package handler

import (
	"encoding/csv"
	"net/http"
	"regexp"
	"strings"
	"time"
	"fmt"

	"govmonitor-it/backend/internal/domain"
	"govmonitor-it/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

var macPattern = regexp.MustCompile(`^([0-9a-fA-F]{2}[:\-]){5}[0-9a-fA-F]{2}$`)

// ColumnMap maps source CSV header names (lowercased) to Device field names.
var defaultColumnMap = map[string]string{
	"name":         "name",
	"device name":  "name",
	"model (unifi)": "model",
	"model":        "model",
	"mac address":  "mac",
	"mac":          "mac",
	"ip address":   "ip",
	"ip":           "ip",
	"status":       "firmware_status", // "Up to date" = firmware version, NOT monitoring status
	"type":         "type",
	"location":     "location",
}

// ImportHandler handles the CSV/XLSX bulk import endpoint.
type ImportHandler struct {
	deviceRepo repository.DeviceRepository
}

func NewImportHandler(deviceRepo repository.DeviceRepository) *ImportHandler {
	return &ImportHandler{deviceRepo: deviceRepo}
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

	// Parse CSV (XLSX requires excelize which is not yet in go.mod — add as needed)
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
	for i, h := range headers {
		normalized := strings.ToLower(strings.TrimSpace(h))
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

		// Validation
		name := fields["name"]
		mac := normalizeMACAddress(fields["mac"])
		ip := fields["ip"]

		if name == "" {
			result.Status = "failed"
			result.Reason = "Missing device name"
			summary.Failed++
			summary.Results = append(summary.Results, result)
			continue
		}

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

		// Build device — default DHCP addressing (common for APs)
		dev := &domain.Device{
			Name:             name,
			Type:             mapDeviceType(fields["model"]),
			MAC:              mac,
			IP:               ip,
			Model:            fields["model"],
			FirmwareStatus:   fields["firmware_status"],
			Status:           domain.StatusUP,
			AddressingMode:   domain.AddressingDHCP,
			FailureThreshold: 3,
			CheckedSecondsAgo: 0,
			LastChecked:      time.Now().Format("15:04:05"),
			Uptime30d:        100.0,
		}

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

// mapDeviceType infers DeviceType from UniFi model string.
func mapDeviceType(model string) domain.DeviceType {
	m := strings.ToLower(model)
	switch {
	case strings.Contains(m, "u6") || strings.Contains(m, "uap") || strings.Contains(m, "ap"):
		return domain.AccessPoint
	case strings.Contains(m, "usw") || strings.Contains(m, "switch"):
		return domain.Switch
	case strings.Contains(m, "usg") || strings.Contains(m, "router") || strings.Contains(m, "udm"):
		return domain.Router
	default:
		return domain.AccessPoint // most imports are APs
	}
}
