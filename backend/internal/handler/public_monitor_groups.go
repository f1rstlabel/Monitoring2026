package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"sanoc/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type publicMonitorGroupInput struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"displayOrder"`
	Enabled      *bool  `json:"enabled"`
}

func (h *Handler) GetPublicMonitorGroups(c *gin.Context) {
	if h.publicMonitorGroupRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitor groups are not initialized"})
		return
	}
	groups, err := h.publicMonitorGroupRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor groups"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": groups, "data": groups})
}

func (h *Handler) CreatePublicMonitorGroup(c *gin.Context) {
	if h.publicMonitorGroupRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitor groups are not initialized"})
		return
	}
	var input publicMonitorGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group payload"})
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group name is required and must be at most 120 characters"})
		return
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	created, err := h.publicMonitorGroupRepo.Create(&domain.PublicMonitorGroup{
		Name: input.Name, Description: strings.TrimSpace(input.Description), DisplayOrder: input.DisplayOrder, Enabled: enabled,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "A group with this name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create public monitor group"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) UpdatePublicMonitorGroup(c *gin.Context) {
	if h.publicMonitorGroupRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitor groups are not initialized"})
		return
	}
	group, err := h.publicMonitorGroupRepo.GetByID(c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) || group == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Public monitor group not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor group"})
		return
	}
	var input publicMonitorGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group payload"})
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group name is required and must be at most 120 characters"})
		return
	}
	if group.Name == "Ungrouped" && input.Name != "Ungrouped" {
		c.JSON(http.StatusConflict, gin.H{"error": "The Ungrouped group cannot be renamed"})
		return
	}
	group.Name = input.Name
	group.Description = strings.TrimSpace(input.Description)
	group.DisplayOrder = input.DisplayOrder
	if input.Enabled != nil {
		group.Enabled = *input.Enabled
	}
	if err := h.publicMonitorGroupRepo.Update(group); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "A group with this name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update public monitor group"})
		return
	}
	updated, _ := h.publicMonitorGroupRepo.GetByID(group.ID)
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) DeletePublicMonitorGroup(c *gin.Context) {
	if h.publicMonitorGroupRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Public monitor groups are not initialized"})
		return
	}
	group, err := h.publicMonitorGroupRepo.GetByID(c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) || group == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Public monitor group not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public monitor group"})
		return
	}
	if group.Name == "Ungrouped" {
		c.JSON(http.StatusConflict, gin.H{"error": "The Ungrouped group cannot be deleted"})
		return
	}
	if group.MonitorCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Move monitors out of this group before deleting it"})
		return
	}
	if err := h.publicMonitorGroupRepo.Delete(group.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete public monitor group"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
