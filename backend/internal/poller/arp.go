package poller

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"sanoc/backend/internal/domain"

	"github.com/gosnmp/gosnmp"
)

// ResolveIPFromARP parses the OS ARP table to find the current IP for a MAC address.
// Works on Windows (arp -a) and Linux (arp -n / ip neigh).
// Returns empty string if MAC is not in the ARP cache.
func ResolveIPFromARP(mac string) string {
	return ResolveIPFromARPWarm(mac, "")
}

func ResolveIPFromARPWarm(mac, lastKnownIP string) string {
	return ResolveIPFromARPWithCoreSwitch(mac, lastKnownIP, domain.CoreSwitchConfig{})
}

func ResolveIPFromARPWithCoreSwitch(mac, lastKnownIP string, cfg domain.CoreSwitchConfig) string {
	normalizedTarget := normalizeMACForSearch(mac)
	if normalizedTarget == "" {
		return ""
	}

	entries := readARPTable()
	for _, e := range entries {
		if normalizeMACForSearch(e.MAC) == normalizedTarget {
			return e.IP
		}
	}

	// Warm ARP table if lastKnownIP is provided
	if lastKnownIP != "" {
		probeHost(lastKnownIP)
		entries = readARPTable()
		for _, e := range entries {
			if normalizeMACForSearch(e.MAC) == normalizedTarget {
				return e.IP
			}
		}
	}

	// Fall back to Core Switch SNMP ARP table query if configured
	if cfg.IP != "" {
		snmpEntries, err := QueryCoreSwitchARPTable(cfg)
		if err == nil {
			for _, e := range snmpEntries {
				if normalizeMACForSearch(e.MAC) == normalizedTarget {
					return e.IP
				}
			}
		}
	}

	return ""
}

// ResolveMACFromARP parses the OS ARP table to find the MAC for a given IP.
// Sends a quick ICMP/TCP probe to force ARP resolution if the cache is cold.
// Returns empty string if IP is not in the ARP cache.
func ResolveMACFromARP(ip string) string {
	return ResolveMACFromARPWithCoreSwitch(ip, domain.CoreSwitchConfig{})
}

func ResolveMACFromARPWithCoreSwitch(ip string, cfg domain.CoreSwitchConfig) string {
	if ip == "" {
		return ""
	}

	// 1. Send a quick probe to force ARP resolution (warm the local ARP cache)
	probeHost(ip)

	// 2. Read local ARP table immediately after
	entries := readARPTable()
	for _, e := range entries {
		if e.IP == ip {
			return normalizeMACForSearch(e.MAC) // Normalize output to colon format
		}
	}

	// 3. Fall back to Core Switch SNMP ARP table walk if local ARP misses
	if cfg.IP != "" {
		snmpEntries, err := QueryCoreSwitchARPTable(cfg)
		if err == nil {
			for _, e := range snmpEntries {
				if e.IP == ip {
					return normalizeMACForSearch(e.MAC)
				}
			}
		}
	}

	return ""
}

// QueryCoreSwitchARPTable walks ipNetToMediaPhysAddress (.1.3.6.1.2.1.4.22.1.2) on the configured Core Switch/Router.
// If empty, falls back to ipNetToPhysicalPhysAddress (.1.3.6.1.2.1.4.35.1.4).
func QueryCoreSwitchARPTable(cfg domain.CoreSwitchConfig) ([]ARPEntry, error) {
	if cfg.IP == "" {
		return nil, nil
	}
	community := cfg.Community
	if community == "" {
		community = "public"
	}
	port := cfg.Port
	if port == 0 {
		port = 161
	}
	version := gosnmp.Version2c
	if cfg.Version == "v1" {
		version = gosnmp.Version1
	}

	snmpClient := &gosnmp.GoSNMP{
		Target:    cfg.IP,
		Port:      uint16(port),
		Community: community,
		Version:   version,
		Timeout:   time.Duration(2) * time.Second,
		Retries:   1,
	}

	err := snmpClient.Connect()
	if err != nil {
		log.Printf("[SNMP ARP] Failed to connect to core switch %s: %v", cfg.IP, err)
		return nil, err
	}
	defer snmpClient.Conn.Close()

	var entries []ARPEntry

	// 1. Primary MIB-2 OID: ipNetToMediaPhysAddress (.1.3.6.1.2.1.4.22.1.2.<ifIndex>.<ip1>.<ip2>.<ip3>.<ip4>)
	oid := ".1.3.6.1.2.1.4.22.1.2"
	_ = snmpClient.Walk(oid, func(pdu gosnmp.SnmpPDU) error {
		parts := strings.Split(strings.TrimPrefix(pdu.Name, "."), ".")
		if len(parts) >= 14 {
			ipParts := parts[len(parts)-4:]
			ipStr := strings.Join(ipParts, ".")
			if parsedIP := net.ParseIP(ipStr); parsedIP != nil && parsedIP.To4() != nil {
				var macBytes []byte
				switch v := pdu.Value.(type) {
				case []byte:
					macBytes = v
				case string:
					macBytes = []byte(v)
				}
				if len(macBytes) == 6 {
					macStr := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
						macBytes[0], macBytes[1], macBytes[2], macBytes[3], macBytes[4], macBytes[5])
					entries = append(entries, ARPEntry{IP: ipStr, MAC: macStr})
				}
			}
		}
		return nil
	})

	// 2. Fallback if no entries found: ipNetToPhysicalPhysAddress (.1.3.6.1.2.1.4.35.1.4)
	if len(entries) == 0 {
		fallbackOID := ".1.3.6.1.2.1.4.35.1.4"
		_ = snmpClient.Walk(fallbackOID, func(pdu gosnmp.SnmpPDU) error {
			parts := strings.Split(strings.TrimPrefix(pdu.Name, "."), ".")
			if len(parts) >= 15 {
				ipParts := parts[len(parts)-4:]
				ipStr := strings.Join(ipParts, ".")
				if parsedIP := net.ParseIP(ipStr); parsedIP != nil && parsedIP.To4() != nil {
					var macBytes []byte
					switch v := pdu.Value.(type) {
					case []byte:
						macBytes = v
					case string:
						macBytes = []byte(v)
					}
					if len(macBytes) == 6 {
						macStr := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
							macBytes[0], macBytes[1], macBytes[2], macBytes[3], macBytes[4], macBytes[5])
						entries = append(entries, ARPEntry{IP: ipStr, MAC: macStr})
					}
				}
			}
			return nil
		})
	}

	log.Printf("[SNMP ARP] Core switch %s table walk returned %d entries", cfg.IP, len(entries))
	return entries, nil
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

type ARPEntry struct {
	IP  string
	MAC string
}


func readARPTable() []ARPEntry {
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
func parseARPOutput(raw string) []ARPEntry {
	ipRe := regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
	macRe := regexp.MustCompile(`\b([0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2})\b`)

	var entries []ARPEntry
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

		entries = append(entries, ARPEntry{IP: ipMatch, MAC: normalizedMAC})
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
