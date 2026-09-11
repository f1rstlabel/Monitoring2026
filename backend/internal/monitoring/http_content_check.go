package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sanoc/backend/internal/domain"
)

func RunHTTPKeywordCheck(ctx context.Context, monitor domain.PublicMonitor) CheckResult {
	return runHTTPAssertionCheck(ctx, monitor, func(body []byte) error {
		if !strings.Contains(string(body), monitor.Keyword) {
			return fmt.Errorf("response body does not contain configured keyword")
		}
		return nil
	})
}

func RunHTTPJSONCheck(ctx context.Context, monitor domain.PublicMonitor) CheckResult {
	return runHTTPAssertionCheck(ctx, monitor, func(body []byte) error {
		var payload interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			return fmt.Errorf("response body is not valid JSON")
		}
		value, ok := lookupJSONPath(payload, monitor.JSONPath)
		if !ok {
			return fmt.Errorf("JSON path %q was not found", monitor.JSONPath)
		}
		if fmt.Sprint(value) != monitor.ExpectedValue {
			return fmt.Errorf("JSON path %q expected %q, got %q", monitor.JSONPath, monitor.ExpectedValue, fmt.Sprint(value))
		}
		return nil
	})
}

func runHTTPAssertionCheck(ctx context.Context, monitor domain.PublicMonitor, assertion func([]byte) error) CheckResult {
	if err := ValidateTargetURL(monitor.TargetURL); err != nil {
		return CheckResult{Status: domain.PublicMonitorDown, Error: err.Error()}
	}
	timeout := monitor.TimeoutSeconds
	if timeout < 1 {
		timeout = 10
	}
	client := safeHTTPClient(time.Duration(timeout) * time.Second)
	attempts := monitor.RetryCount + 1
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	lastCode, lastLatency := 0, 0
	for attempt := 0; attempt < attempts; attempt++ {
		requestCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, monitor.TargetURL, nil)
		if err == nil {
			SetBrowserHeaders(request)
			started := time.Now()
			response, requestErr := client.Do(request)
			lastLatency = int(time.Since(started).Milliseconds())
			if requestErr == nil {
				lastCode = response.StatusCode
				body, readErr := io.ReadAll(io.LimitReader(response.Body, 2<<20))
				response.Body.Close()
				switch {
				case readErr != nil:
					lastErr = readErr
				case response.StatusCode < 200 || response.StatusCode >= 400:
					lastErr = fmt.Errorf("HTTP status %d", response.StatusCode)
				default:
					lastErr = assertion(body)
					if lastErr == nil {
						cancel()
						return CheckResult{Status: domain.PublicMonitorUp, StatusCode: response.StatusCode, LatencyMs: lastLatency}
					}
				}
			} else {
				lastErr = requestErr
			}
		} else {
			lastErr = err
		}
		cancel()
		if attempt+1 < attempts {
			select {
			case <-ctx.Done():
				return CheckResult{Status: domain.PublicMonitorDown, StatusCode: lastCode, LatencyMs: lastLatency, Error: ctx.Err().Error()}
			case <-time.After(150 * time.Millisecond):
			}
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("HTTP assertion failed")
	}
	return CheckResult{Status: domain.PublicMonitorDown, StatusCode: lastCode, LatencyMs: lastLatency, Error: lastErr.Error()}
}

// lookupJSONPath supports simple dot notation (data.status) and array indexes (items.0).
func lookupJSONPath(value interface{}, path string) (interface{}, bool) {
	current := value
	for _, segment := range strings.Split(strings.TrimPrefix(strings.TrimSpace(path), "."), ".") {
		if segment == "" {
			continue
		}
		if index, err := strconv.Atoi(strings.Trim(segment, "[]")); err == nil {
			items, ok := current.([]interface{})
			if !ok || index < 0 || index >= len(items) {
				return nil, false
			}
			current = items[index]
			continue
		}
		object, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = object[segment]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
