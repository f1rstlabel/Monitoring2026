package handler

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var validHostRegex = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`)

func isValidTargetHost(target string) bool {
	target = strings.TrimSpace(target)
	if target == "" || len(target) > 253 {
		return false
	}
	// Check if valid IP
	if ip := net.ParseIP(target); ip != nil {
		return true
	}
	// Check if valid hostname
	return validHostRegex.MatchString(target)
}

type PingRequest struct {
	Target string `json:"target"`
	Count  int    `json:"count"`
}

type TracerouteRequest struct {
	Target string `json:"target"`
}

type PortProbeRequest struct {
	Target string `json:"target"`
	Port   int    `json:"port"`
}

// RunPing executes a safe, sanitized ICMP ping test from the server.
func (h *Handler) RunPing(c *gin.Context) {
	var req PingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	target := strings.TrimSpace(req.Target)
	if !isValidTargetHost(target) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target must be a valid IP address or hostname"})
		return
	}

	count := req.Count
	if count < 1 || count > 10 {
		count = 4
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", strconv.Itoa(count), "-w", "2000", target)
	} else {
		cmd = exec.Command("ping", "-c", strconv.Itoa(count), "-W", "2", target)
	}

	start := time.Now()
	outBytes, err := cmd.CombinedOutput()
	durationMs := time.Since(start).Milliseconds()
	outStr := string(outBytes)

	lines := strings.Split(strings.ReplaceAll(outStr, "\r\n", "\n"), "\n")
	var cleanLines []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			cleanLines = append(cleanLines, l)
		}
	}

	success := (err == nil) && !strings.Contains(strings.ToLower(outStr), "100% loss") && !strings.Contains(strings.ToLower(outStr), "100% packet loss")

	c.JSON(http.StatusOK, gin.H{
		"target":     target,
		"success":    success,
		"durationMs": durationMs,
		"output":     cleanLines,
		"raw":        outStr,
	})
}

// RunTraceroute executes a safe network path traceroute from the server.
func (h *Handler) RunTraceroute(c *gin.Context) {
	var req TracerouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	target := strings.TrimSpace(req.Target)
	if !isValidTargetHost(target) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target must be a valid IP address or hostname"})
		return
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tracert", "-d", "-h", "15", "-w", "1500", target)
	} else {
		cmd = exec.Command("traceroute", "-n", "-m", "15", "-w", "2", target)
	}

	start := time.Now()
	outBytes, _ := cmd.CombinedOutput()
	durationMs := time.Since(start).Milliseconds()
	outStr := string(outBytes)

	lines := strings.Split(strings.ReplaceAll(outStr, "\r\n", "\n"), "\n")
	var cleanLines []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			cleanLines = append(cleanLines, l)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"target":     target,
		"durationMs": durationMs,
		"output":     cleanLines,
		"raw":        outStr,
	})
}

// RunPortProbe tests TCP socket connection to a specific target port.
func (h *Handler) RunPortProbe(c *gin.Context) {
	var req PortProbeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	target := strings.TrimSpace(req.Target)
	if !isValidTargetHost(target) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target must be a valid IP address or hostname"})
		return
	}

	if req.Port < 1 || req.Port > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Port must be between 1 and 65535"})
		return
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(target, strconv.Itoa(req.Port)), 3*time.Second)
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"target":    target,
			"port":      req.Port,
			"open":      false,
			"latencyMs": latencyMs,
			"error":     err.Error(),
			"message":   fmt.Sprintf("Port %d on %s is CLOSED / TIMEOUT", req.Port, target),
		})
		return
	}
	defer conn.Close()

	if latencyMs <= 0 {
		latencyMs = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"target":    target,
		"port":      req.Port,
		"open":      true,
		"latencyMs": latencyMs,
		"message":   fmt.Sprintf("Port %d on %s is OPEN (%dms)", req.Port, target, latencyMs),
	})
}
