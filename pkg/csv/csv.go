package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	model "core/internal/model"
)

type RowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type ParseResult struct {
	Leads  []model.Lead
	Errors []RowError
}

func Parse(reader io.Reader) (ParseResult, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	header, err := csvReader.Read()
	if err != nil {
		return ParseResult{}, fmt.Errorf("read csv header: %w", err)
	}

	index := mapHeaders(header)
	if index.fullName < 0 || index.company < 0 || index.phone < 0 {
		return ParseResult{}, fmt.Errorf(
			"csv must include fullname, company, and phone number columns",
		)
	}

	result := ParseResult{}
	rowNum := 1

	for {
		rowNum++
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, RowError{
				Row:     rowNum,
				Message: err.Error(),
			})
			continue
		}

		fullName := fieldAt(record, index.fullName)
		company := fieldAt(record, index.company)
		phone := normalizePhone(fieldAt(record, index.phone))

		if fullName == "" || company == "" || phone == "" {
			result.Errors = append(result.Errors, RowError{
				Row:     rowNum,
				Message: "missing required field",
			})
			continue
		}

		result.Leads = append(result.Leads, model.Lead{
			FullName:    fullName,
			Company:     company,
			PhoneNumber: phone,
			Status:      model.LeadStatusPending,
		})
	}

	return result, nil
}

type headerIndex struct {
	fullName int
	company  int
	phone    int
}

func mapHeaders(header []string) headerIndex {
	idx := headerIndex{fullName: -1, company: -1, phone: -1}
	for i, col := range header {
		key := normalizeHeader(col)
		switch key {
		case "fullname", "full_name", "name":
			idx.fullName = i
		case "company":
			idx.company = i
		case "phonenumber", "phone_number", "phone", "mobile", "number":
			idx.phone = i
		}
	}
	return idx
}

func normalizeHeader(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "_", "")
	return value
}

func fieldAt(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

// normalizePhone stores an Indian national number: digits only, last 10.
func normalizePhone(value string) string {
	var b strings.Builder
	for _, ch := range strings.TrimSpace(value) {
		if ch >= '0' && ch <= '9' {
			b.WriteRune(ch)
		}
	}
	digits := b.String()
	if len(digits) > 10 {
		digits = digits[len(digits)-10:]
	}
	return digits
}
