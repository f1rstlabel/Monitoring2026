package handler

import (
	"encoding/csv"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"govmonitor-it/backend/internal/domain"
	"govmonitor-it/backend/internal/middleware"
	"govmonitor-it/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

var macPattern = regexp.MustCompile(`^([0-9a-fA-F]{2}[:\-]){5}[0-9a-fA-F]{2}$`)

// ColumnMap maps source CSV header names (lowercased) to Device field names.
var defaultColumnMap = map[string]string{
	"device name":                       "name",
	"device type":                       "type",
	"addressing mode":                   "addressing_mode",
	"ip address":                        "ip",
	"mac address":                       "mac",
	"location":                          "location",
	"rack":                              "rack",
	"snmp polling enabled":              "snmp_enabled",
	"snmp community":                    "snmp_community",
	"snmp port":                         "snmp_port",
	"custom failure threshold override": "use_custom_threshold",
	"custom failure threshold value":    "custom_failure_threshold",
}

// ImportHandler handles the CSV/XLSX bulk import endpoint.
type ImportHandler struct {
	deviceRepo   repository.DeviceRepository
	locationRepo repository.LocationRepository
}

func NewImportHandler(deviceRepo repository.DeviceRepository, locationRepo repository.LocationRepository) *ImportHandler {
	return &ImportHandler{
		deviceRepo:   deviceRepo,
		locationRepo: locationRepo,
	}
}

// ImportDevices handles POST /api/v1/devices/import
// Content-Type: multipart/form-data
// Role required: superadmin (enforced by RequireRole middleware on the route)
func (h *ImportHandler) ImportDevices(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'file' field in form data"})
		return
	}
	defer file.Close()

	filename := strings.ToLower(header.Filename)
	if !strings.HasSuffix(filename, ".csv") && !strings.HasSuffix(filename, ".xlsx") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .csv and .xlsx files are supported"})
		return
	}

	// Parse CSV
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to parse CSV: " + err.Error()})
		return
	}
	if len(records) < 2 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "CSV has no data rows"})
		return
	}

	// Map headers
	headers := records[0]
	fieldMap := make(map[int]string) // colIndex → fieldName
	for i, hdr := range headers {
		normalized := strings.ToLower(strings.TrimSpace(hdr))
		if field, ok := defaultColumnMap[normalized]; ok {
			fieldMap[i] = field
		}
	}

	summary := domain.ImportSummary{}
	seenMACs := make(map[string]bool)

	for rowIdx, row := range records[1:] {
		result := domain.ImportResult{RowIndex: rowIdx}

		// Extract mapped fields
		fields := make(map[string]string)
		for colIdx, field := range fieldMap {
			if colIdx < len(row) {
				fields[field] = strings.TrimSpace(row[colIdx])
			}
		}

		name := fields["name"]
		mac := normalizeMACAddress(fields["mac"])
		ip := fields["ip"]
		rawType := fields["type"]
		rawAddrMode := fields["addressing_mode"]

		if name == "" {
			result.Status = "failed"
			result.Reason = "Missing device name"
			summary.Failed++
			summary.Results = append(summary.Results, result)
			continue
		}

		// Validate Device Type
		var dType domain.DeviceType
		switch rawType {
		case "Access Point", "Switch", "Router", "SmartPower", "CCTV", "NVR":
			dType = domain.DeviceType(rawType)
		default:
			result.Status = "failed"
			result.Reason = fmt.Sprintf("Unrecognized device type: %s (must be one of: Access Point, Switch, Router, SmartPower, CCTV, NVR)", rawType)
			summary.Failed++
			summary.Results = append(summary.Results, result)
			continue
		}

		// Validate Addressing Mode
		var addrMode domain.AddressingMode
		switch strings.ToLower(rawAddrMode) {
		case "static ip", "static":
			addrMode = domain.AddressingStatic
		case "dhcp reservation", "dhcp":
			addrMode = domain.AddressingDHCP
		default:
			result.Status = "failed"
			result.Reason = fmt.Sprintf("Invalid addressing mode: %s (must be 'Static IP' or 'DHCP Reservation')", rawAddrMode)
			summary.Failed++
			summary.Results = append(summary.Results, result)
			continue
		}

		// Validate IP Address
		if addrMode == domain.AddressingStatic && ip == "" {
			result.Status = "failed"
			result.Reason = "IP Address is required when Addressing Mode is Static IP"
			summary.Failed++
			summary.Results = append(summary.Results, result)
			continue
		}
		if ip != "" && !isValidIPv4(ip) {
			result.Status = "failed"
			result.Reason = fmt.Sprintf("Invalid IP format: %s", ip)
			summary.Failed++
			summary.Results = append(summary.Results, result)
			continue
		}

		// Validate MAC Address
		if mac == "" {
			result.Status = "failed"
			result.Reason = "Missing MAC address"
			summary.Failed++
			summary.Results = append(summary.Results, result)
			continue
		}

		if !macPattern.MatchString(mac) {
			result.Status = "failed"
			result.Reason = fmt.Sprintf("Invalid MAC format: %s", mac)
			summary.Failed++
			summary.Results = append(summary.Results, result)
			continue
		}

		// Duplicate check (in-batch)
		if seenMACs[mac] {
			result.Status = "skipped"
			result.Reason = fmt.Sprintf("Duplicate MAC in import file: %s", mac)
			summary.Skipped++
			summary.Results = append(summary.Results, result)
			continue
		}
		seenMACs[mac] = true

		// Duplicate check (against DB)
		exists, _ := h.deviceRepo.ExistsByMAC(mac)
		if exists {
			result.Status = "skipped"
			result.Reason = fmt.Sprintf("MAC already registered: %s", mac)
			summary.Skipped++
			summary.Results = append(summary.Results, result)
			continue
		}

		// Auto-resolve or create location if location string provided
		var locationID string
		locationName := fields["location"]
		if locationName != "" && h.locationRepo != nil {
			loc, _ := h.locationRepo.GetByName(locationName)
			if loc != nil {
				locationID = loc.ID
			} else {
				newLoc, err := h.locationRepo.Create(&domain.Location{Name: locationName})
				if err == nil && newLoc != nil {
					locationID = newLoc.ID
				}
			}
		}

		// Build device
		dev := &domain.Device{
			Name:              name,
			Type:              dType,
			MAC:               mac,
			IP:                ip,
			Status:            domain.StatusUP,
			AddressingMode:    addrMode,
			LocationID:        locationID,
			Location:          locationName,
			Rack:              fields["rack"],
			FailureThreshold:  3,
			CheckedSecondsAgo: 0,
			LastChecked:       time.Now().Format("15:04:05"),
			Uptime30d:         100.0,
		}

		// SNMP settings mapping
		if fields["snmp_enabled"] != "" {
			enabled, _ := strconv.ParseBool(fields["snmp_enabled"])
			dev.SNMPEnabled = enabled
			if enabled {
				dev.SNMPCommunity = fields["snmp_community"]
				if port, err := strconv.Atoi(fields["snmp_port"]); err == nil {
					dev.SNMPPort = port
				} else {
					dev.SNMPPort = 161
				}
			}
		}

		// Custom threshold mapping
		if fields["use_custom_threshold"] != "" {
			useOverride, _ := strconv.ParseBool(fields["use_custom_threshold"])
			dev.UseCustomThreshold = useOverride
			if useOverride {
				if thresh, err := strconv.Atoi(fields["custom_failure_threshold"]); err == nil {
					dev.CustomFailureThreshold = &thresh
					dev.FailureThreshold = thresh
				}
			}
		}

		dev.CreatedByUserID = middleware.GetUserID(c)
		created, err := h.deviceRepo.Create(dev)
		if err != nil {
			result.Status = "failed"
			result.Reason = "DB insert error: " + err.Error()
			summary.Failed++
		} else {
			result.Status = "imported"
			result.DeviceID = created.ID
			summary.Imported++
		}
		summary.Results = append(summary.Results, result)
	}

	c.JSON(http.StatusOK, summary)
}

// normalizeMACAddress converts any MAC format to lowercase colon-separated.
func normalizeMACAddress(raw string) string {
	// Strip all separators, lowercase
	cleaned := strings.ToLower(strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') {
			return r
		}
		return -1
	}, raw))

	if len(cleaned) != 12 {
		return raw // Return as-is if not 12 hex chars — validation will catch it
	}

	// Insert colons
	parts := make([]string, 6)
	for i := 0; i < 6; i++ {
		parts[i] = cleaned[i*2 : i*2+2]
	}
	return strings.Join(parts, ":")
}

func isValidIPv4(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.To4() != nil
}
