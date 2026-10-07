package controller

import (
	"log"
	"net/http"

	model "autocall/internal/model"
	csv "autocall/pkg/csv"

	"github.com/gin-gonic/gin"
)

// Init godoc
//
//	@Summary		Import leads from CSV
//	@Description	Upload a CSV with fullname, company, and phone number columns. Duplicate phone numbers are skipped.
//	@Tags			autocall
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"CSV file"
//	@Success		200		{object}	InitImportResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/autocall/init [post]
func (c *Controller) Init(context *gin.Context) {
	file, err := context.FormFile("file")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "csv file is required (form field: file)"})
		return
	}

	opened, err := file.Open()
	if err != nil {
		log.Printf("init: open upload: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read csv"})
		return
	}
	defer opened.Close()

	parsed, err := csv.Parse(opened)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inserted, skipped, err := c.leads.InsertIgnoreDuplicates(parsed.Leads)
	if err != nil {
		log.Printf("init: insert leads: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store leads"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"rowsProcessed": len(parsed.Leads) + len(parsed.Errors),
			"inserted":      inserted,
			"skipped":       skipped,
			"failed":        len(parsed.Errors),
			"errors":        parsed.Errors,
		},
	})
}

// GetStatus godoc
//
//	@Summary		Get calling status
//	@Description	Returns lead counts by status and the current calling lead when applicable.
//	@Tags			autocall
//	@Produce		json
//	@Success		200	{object}	StatusResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/autocall/status [get]
func (c *Controller) GetStatus(context *gin.Context) {
	counts, err := c.leads.StatusCounts()
	if err != nil {
		log.Printf("status: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load status"})
		return
	}

	var current *model.Lead
	if id := c.caller.CurrentLeadID(); id != nil {
		var lead model.Lead
		if err := c.db.First(&lead, "id = ?", *id).Error; err == nil {
			current = &lead
		}
	} else {
		current, _ = c.leads.CurrentCalling()
	}

	context.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total":        counts.Total,
			"pending":      counts.Pending,
			"calling":      counts.Calling,
			"completed":    counts.Completed,
			"interested":   counts.Interested,
			"failed":       counts.Failed,
			"current_lead": current,
		},
	})
}

// GetLeads godoc
//
//	@Summary		List interested leads
//	@Description	Returns leads that pressed DTMF 1 (status interested) from PostgreSQL.
//	@Tags			autocall
//	@Produce		json
//	@Success		200	{object}	LeadsResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/autocall/leads [get]
func (c *Controller) GetLeads(context *gin.Context) {
	leads, err := c.leads.Interested()
	if err != nil {
		log.Printf("leads: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load leads"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"data": leads})
}
