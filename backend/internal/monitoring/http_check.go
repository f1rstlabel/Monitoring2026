package monitoring

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"sanoc/backend/internal/domain"
)

type CheckResult struct {
	Status     domain.PublicMonitorStatus
	StatusCode int
	LatencyMs  int
	Error      string
}

// ValidateTargetURL validates the URL shape. The transport additionally
// validates resolved IPs to prevent monitoring targets from reaching private
// network ranges through DNS rebinding.
func ValidateTargetURL(raw string) error {
	u, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Hostname() == "" {
		return fmt.Errorf("target URL must be a valid HTTP or HTTPS URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS targets are supported")
	}
	if u.User != nil {
		return fmt.Errorf("target URL must not contain embedded credentials")
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if err := ValidatePublicHost(host); err != nil {
		return err
	}
	if port := u.Port(); port != "" {
		if p, err := strconv.Atoi(port); err != nil || p < 1 || p > 65535 {
			return fmt.Errorf("target URL contains an invalid port")
		}
	}
	return nil
}

// ValidatePublicHost prevents a monitor from being used as an SSRF primitive.
// Domains are resolved before they are accepted, and every returned address
// must be publicly routable. The transport repeats this check immediately
// before dialing to reduce DNS-rebinding risk.
func ValidatePublicHost(raw string) error {
	host := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(raw, ".")))
	host = strings.Trim(host, "[]")
	if host == "" {
		return fmt.Errorf("target host is required")
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return fmt.Errorf("private hostnames are not allowed")
	}
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicIP(ip) {
			return fmt.Errorf("private or reserved IP addresses are not allowed")
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("target host could not be resolved")
	}
	if len(ips) == 0 {
		return fmt.Errorf("target host has no address")
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return fmt.Errorf("target resolved to a private or reserved address")
		}
	}
	return nil
}

func isPublicIP(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}
	return true
}

func safeHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: timeout}
	transport := &http.Transport{
		// Public monitoring must connect directly so a process-wide proxy cannot
		// bypass the target IP safety checks.
		Proxy:                 nil,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          20,
		MaxIdleConnsPerHost:   4,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, err
			}
			var lastErr error
			// Prefer IPv4: some Windows environments allow outbound IPv4 while
			// blocking IPv6 sockets even when DNS returns IPv6 first.
			for _, preferIPv4 := range []bool{true, false} {
				for _, ip := range ips {
					if !isPublicIP(ip) || (ip.To4() != nil) != preferIPv4 {
						continue
					}
					conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
					if dialErr == nil {
						return conn, nil
					}
					lastErr = dialErr
				}
			}
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, fmt.Errorf("target resolved to a private or reserved address")
		},
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			if err := ValidateTargetURL(req.URL.String()); err != nil {
				return err
			}
			SetBrowserHeaders(req)
			return nil
		},
	}
}

// SetBrowserHeaders populates standard modern browser headers on an outgoing HTTP request
// to prevent legitimate public monitors from being blocked by strict WAFs or anti-bot rules.
func SetBrowserHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Cache-Control", "max-age=0")
}

func RunHTTPCheck(ctx context.Context, monitor domain.PublicMonitor) CheckResult {
	if err := ValidateTargetURL(monitor.TargetURL); err != nil {
		return CheckResult{Status: domain.PublicMonitorDown, Error: err.Error()}
	}
	timeout := time.Duration(monitor.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := safeHTTPClient(timeout)
	attempts := monitor.RetryCount + 1
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	lastStatusCode := 0
	lastLatencyMs := 0
	for attempt := 0; attempt < attempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, timeout)
		request, err := http.NewRequestWithContext(attemptCtx, http.MethodGet, monitor.TargetURL, nil)
		if err == nil {
			SetBrowserHeaders(request)
			started := time.Now()
			response, requestErr := client.Do(request)
			lastLatencyMs = int(time.Since(started).Milliseconds())
			if requestErr == nil {
				lastStatusCode = response.StatusCode
				_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
				response.Body.Close()
				cancel()
				if response.StatusCode >= 200 && response.StatusCode < 400 {
					return CheckResult{Status: domain.PublicMonitorUp, StatusCode: response.StatusCode, LatencyMs: lastLatencyMs}
				}
				lastErr = fmt.Errorf("HTTP status %d", response.StatusCode)
			} else {
				lastErr = requestErr
				cancel()
			}
		} else {
			lastErr = err
			cancel()
		}
		if attempt+1 < attempts {
			select {
			case <-ctx.Done():
				return CheckResult{Status: domain.PublicMonitorDown, StatusCode: lastStatusCode, LatencyMs: lastLatencyMs, Error: ctx.Err().Error()}
			case <-time.After(150 * time.Millisecond):
			}
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("request failed")
	}
	return CheckResult{Status: domain.PublicMonitorDown, StatusCode: lastStatusCode, LatencyMs: lastLatencyMs, Error: lastErr.Error()}
}
