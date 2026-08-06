package poller

import (
	"log"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// ResolveIPFromARP parses the OS ARP table to find the current IP for a MAC address.
// Works on Windows (arp -a) and Linux (arp -n / ip neigh).
// Returns empty string if MAC is not in the ARP cache.
func ResolveIPFromARP(mac string) string {
	return ResolveIPFromARPWarm(mac, "")
}

func ResolveIPFromARPWarm(mac, lastKnownIP string) string {
	normalizedTarget := normalizeMACForSearch(mac)
	if normalizedTarget == "" {
		return ""
	}

	entries := readARPTable()
	for _, e := range entries {
		if normalizeMACForSearch(e.mac) == normalizedTarget {
			return e.ip
		}
	}

	// Warm ARP table if lastKnownIP is provided
	if lastKnownIP != "" {
		probeHost(lastKnownIP)
		entries = readARPTable()
		for _, e := range entries {
			if normalizeMACForSearch(e.mac) == normalizedTarget {
				return e.ip
			}
		}
	}

	return ""
}

// ResolveMACFromARP parses the OS ARP table to find the MAC for a given IP.
// Sends a quick ICMP/TCP probe to force ARP resolution if the cache is cold.
// Returns empty string if IP is not in the ARP cache.
func ResolveMACFromARP(ip string) string {
	if ip == "" {
		return ""
	}

	// 1. Send a quick probe to force ARP resolution (warm the ARP cache)
	probeHost(ip)

	// 2. Read ARP table immediately after
	entries := readARPTable()
	for _, e := range entries {
		if e.ip == ip {
			return normalizeMACForSearch(e.mac) // Normalize output to colon format
		}
	}
	return ""
}

// probeHost sends a short ICMP ping or TCP connect to force the OS to resolve ARP
func probeHost(ip string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "500", ip)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "1", ip)
	}
	_ = cmd.Run() // Ignore error — the goal is just to trigger an ARP request

	// Fallback TCP probes across common ports if ping fails or ICMP is filtered
	ports := []string{"80", "443", "22", "135", "445", "8080"}
	for _, port := range ports {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), 300*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
	}
}

type arpEntry struct {
	ip  string
	mac string
}

func readARPTable() []arpEntry {
	var output []byte
	var err error

	if runtime.GOOS == "windows" {
		output, err = exec.Command("arp", "-a").Output()
	} else {
		output, err = exec.Command("arp", "-n").Output()
		if err != nil {
			// Fallback to 'ip neigh' on modern Linux
			output, err = exec.Command("ip", "neigh").Output()
		}
	}

	if err != nil {
		log.Printf("[ARP] Failed to read ARP table: %v", err)
		return nil
	}

	return parseARPOutput(string(output))
}

// parseARPOutput extracts (ip, mac) pairs from ARP table output.
// Windows format: "  10.20.0.1    dynamic    00-1a-2b-3c-4d-5e"
// Linux arp -n:   "10.20.0.1  ether  00:1a:2b:3c:4d:5e  C  eth0"
// Linux ip neigh: "10.20.0.1 dev eth0 lladdr 00:1a:2b:3c:4d:5e REACHABLE"
func parseARPOutput(raw string) []arpEntry {
	ipRe := regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
	macRe := regexp.MustCompile(`\b([0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2})\b`)

	var entries []arpEntry
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		ipMatch := ipRe.FindString(line)
		macMatch := macRe.FindString(line)

		if ipMatch == "" || macMatch == "" {
			continue
		}

		// Skip Windows header lines and broadcast/multicast MACs
		if strings.Contains(line, "Interface") || strings.Contains(line, "ff-ff-ff") || strings.Contains(line, "ff:ff:ff") {
			continue
		}

		// Normalize MAC address to colon format (e.g. 00-1a-2b... -> 00:1a:2b...)
		normalizedMAC := strings.ReplaceAll(strings.ToLower(macMatch), "-", ":")

		entries = append(entries, arpEntry{ip: ipMatch, mac: normalizedMAC})
	}
	return entries
}

// normalizeMACForSearch converts a MAC address to lowercase colon-separated for comparison.
// e.g. "00-1A-2B-3C-4D-5E" → "00:1a:2b:3c:4d:5e"
func normalizeMACForSearch(mac string) string {
	mac = strings.ToLower(strings.TrimSpace(mac))
	mac = strings.ReplaceAll(mac, "-", ":")
	// Validate it looks like a MAC
	if len(mac) < 11 {
		return ""
	}
	return mac
}
