package monitoring

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"sanoc/backend/internal/domain"
)

// RunPublicMonitorCheck dispatches a monitor to the probe that matches its
// configured type. HTTP retains the existing request semantics.
func RunPublicMonitorCheck(ctx context.Context, monitor domain.PublicMonitor) CheckResult {
	switch monitor.MonitorType {
	case domain.PublicMonitorHTTPKeyword:
		return RunHTTPKeywordCheck(ctx, monitor)
	case domain.PublicMonitorHTTPJSON:
		return RunHTTPJSONCheck(ctx, monitor)
	case domain.PublicMonitorTCP:
		return runWithRetry(ctx, monitor, probeTCP)
	case domain.PublicMonitorPing:
		return runWithRetry(ctx, monitor, probePing)
	case domain.PublicMonitorDNS:
		return runWithRetry(ctx, monitor, probeDNS)
	case "", domain.PublicMonitorHTTP:
		return RunHTTPCheck(ctx, monitor)
	default:
		return CheckResult{Status: domain.PublicMonitorDown, Error: fmt.Sprintf("unsupported monitor type: %s", monitor.MonitorType)}
	}
}

type probeFunc func(context.Context, domain.PublicMonitor) (int, error)

func runWithRetry(ctx context.Context, monitor domain.PublicMonitor, probe probeFunc) CheckResult {
	timeout := time.Duration(monitor.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	attempts := monitor.RetryCount + 1
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	var latency int
	for attempt := 0; attempt < attempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, timeout)
		started := time.Now()
		statusCode, err := probe(attemptCtx, monitor)
		latency = int(time.Since(started).Milliseconds())
		cancel()
		if err == nil {
			return CheckResult{Status: domain.PublicMonitorUp, StatusCode: statusCode, LatencyMs: latency}
		}
		lastErr = err
		if attempt+1 < attempts {
			select {
			case <-ctx.Done():
				return CheckResult{Status: domain.PublicMonitorDown, LatencyMs: latency, Error: ctx.Err().Error()}
			case <-time.After(150 * time.Millisecond):
			}
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("probe failed")
	}
	return CheckResult{Status: domain.PublicMonitorDown, LatencyMs: latency, Error: lastErr.Error()}
}

func resolvePublicIPs(ctx context.Context, host string) ([]net.IP, error) {
	if err := ValidatePublicHost(host); err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", strings.Trim(host, "[]"))
	if err != nil {
		return nil, fmt.Errorf("target host could not be resolved")
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return nil, fmt.Errorf("target resolved to a private or reserved address")
		}
	}
	ordered := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if ip.To4() != nil {
			ordered = append(ordered, ip)
		}
	}
	for _, ip := range ips {
		if ip.To4() == nil {
			ordered = append(ordered, ip)
		}
	}
	return ordered, nil
}

func probeTCP(ctx context.Context, monitor domain.PublicMonitor) (int, error) {
	if monitor.TargetPort < 1 || monitor.TargetPort > 65535 {
		return 0, fmt.Errorf("TCP port must be between 1 and 65535")
	}
	ips, err := resolvePublicIPs(ctx, monitor.TargetHost)
	if err != nil {
		return 0, err
	}
	dialer := &net.Dialer{}
	for _, ip := range ips {
		conn, dialErr := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), strconv.Itoa(monitor.TargetPort)))
		if dialErr == nil {
			_ = conn.Close()
			return 0, nil
		}
		err = dialErr
	}
	if err == nil {
		err = fmt.Errorf("TCP connection failed")
	}
	return 0, err
}

func probePing(ctx context.Context, monitor domain.PublicMonitor) (int, error) {
	ips, err := resolvePublicIPs(ctx, monitor.TargetHost)
	if err != nil {
		return 0, err
	}
	host := ips[0].String()
	timeoutSeconds := monitor.TimeoutSeconds
	if timeoutSeconds < 1 {
		timeoutSeconds = 10
	}
	var command *exec.Cmd
	if runtime.GOOS == "windows" {
		command = exec.CommandContext(ctx, "ping", "-n", "1", "-w", strconv.Itoa(timeoutSeconds*1000), host)
	} else {
		command = exec.CommandContext(ctx, "ping", "-c", "1", "-W", strconv.Itoa(timeoutSeconds), host)
	}
	if output, err := command.CombinedOutput(); err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return 0, fmt.Errorf("ping failed: %s", message)
	}
	return 0, nil
}

func probeDNS(ctx context.Context, monitor domain.PublicMonitor) (int, error) {
	host := strings.TrimSpace(monitor.TargetHost)
	recordType := strings.ToUpper(strings.TrimSpace(monitor.DNSRecordType))
	if recordType == "" {
		recordType = "A"
	}
	resolver := net.DefaultResolver
	switch recordType {
	case "A":
		ips, err := resolver.LookupIP(ctx, "ip4", host)
		if err != nil || len(ips) == 0 {
			return 0, fmt.Errorf("DNS A record not found")
		}
	case "AAAA":
		ips, err := resolver.LookupIP(ctx, "ip6", host)
		if err != nil || len(ips) == 0 {
			return 0, fmt.Errorf("DNS AAAA record not found")
		}
	case "CNAME":
		name, err := resolver.LookupCNAME(ctx, host)
		if err != nil || strings.TrimSpace(name) == "" {
			return 0, fmt.Errorf("DNS CNAME record not found")
		}
	default:
		return 0, fmt.Errorf("DNS record type must be A, AAAA, or CNAME")
	}
	return 0, nil
}
