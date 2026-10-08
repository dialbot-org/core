package controller

import (
	model "core/internal/model"
	csv "core/pkg/csv"
)

type InitImportData struct {
	RowsProcessed int            `json:"rowsProcessed"`
	Inserted      int            `json:"inserted"`
	Skipped       int            `json:"skipped"`
	Failed        int            `json:"failed"`
	Errors        []csv.RowError `json:"errors"`
}

type InitImportResponse struct {
	Data InitImportData `json:"data"`
}

type StatusData struct {
	Total       int64       `json:"total"`
	Pending     int64       `json:"pending"`
	Calling     int64       `json:"calling"`
	Completed   int64       `json:"completed"`
	Interested  int64       `json:"interested"`
	Failed      int64       `json:"failed"`
	CurrentLead *model.Lead `json:"current_lead"`
}

type StatusResponse struct {
	Data StatusData `json:"data"`
}

type LeadsResponse struct {
	Data []model.Lead `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
