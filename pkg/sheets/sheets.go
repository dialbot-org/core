package sheets

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	model "core/internal/model"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Config struct {
	CredentialsFile string
	SpreadsheetID   string
	SheetName       string
	StartRow        int // first data row (rows above are left for headers, etc.)
}

type Client struct {
	config  Config
	service *sheets.Service
	ready   bool
}

func New(cfg Config) *Client {
	cfg.SheetName = normalizeSheetTab(cfg.SheetName)
	if cfg.StartRow < 1 {
		cfg.StartRow = 3
	}
	return &Client{config: cfg}
}

func (c *Client) Enabled() bool {
	return c.config.SpreadsheetID != "" && c.config.CredentialsFile != ""
}

func (c *Client) initialize() error {
	if c.ready {
		return nil
	}
	if !c.Enabled() {
		return fmt.Errorf("google sheets not configured")
	}

	data, err := os.ReadFile(c.config.CredentialsFile)
	if err != nil {
		return err
	}

	creds, err := google.CredentialsFromJSON(
		context.Background(),
		data,
		sheets.SpreadsheetsScope,
	)
	if err != nil {
		return err
	}

	srv, err := sheets.NewService(
		context.Background(),
		option.WithCredentials(creds),
	)
	if err != nil {
		return err
	}

	c.service = srv

	if err := c.validateSheetTab(context.Background()); err != nil {
		return err
	}

	c.ready = true
	log.Printf(
		"sheets: syncing to tab %q (range %s)",
		c.config.SheetName,
		sheetAppendRange(c.config.SheetName, c.config.StartRow),
	)
	return nil
}

func (c *Client) AppendLead(lead model.Lead) error {
	if err := c.initialize(); err != nil {
		return err
	}

	rangeName := sheetAppendRange(c.config.SheetName, c.config.StartRow)
	values := [][]any{
		{
			lead.FullName,
			lead.Company,
			lead.PhoneNumber,
			lead.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	_, err := c.service.Spreadsheets.Values.Append(
		c.config.SpreadsheetID,
		rangeName,
		&sheets.ValueRange{Values: values},
	).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Do()

	if err != nil {
		log.Printf("sheets: append lead %s (range %s): %v", lead.ID, rangeName, err)
		return err
	}

	return nil
}

func normalizeSheetTab(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Trim(name, "'\"")
	if i := strings.Index(name, "!"); i >= 0 {
		name = strings.TrimSpace(name[:i])
	}
	if name == "" {
		return "Leads"
	}
	return name
}

func (c *Client) validateSheetTab(ctx context.Context) error {
	meta, err := c.service.Spreadsheets.Get(c.config.SpreadsheetID).
		Fields("sheets.properties.title").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("load spreadsheet metadata: %w", err)
	}

	var titles []string
	for _, sh := range meta.Sheets {
		if sh.Properties == nil {
			continue
		}
		title := sh.Properties.Title
		titles = append(titles, title)
		if title == c.config.SheetName {
			return nil
		}
	}

	return fmt.Errorf(
		"sheet tab %q not found; set GOOGLE_SHEET_NAME to one of: %s",
		c.config.SheetName,
		strings.Join(titles, ", "),
	)
}

// sheetAppendRange anchors the append table at startRow (e.g. row 3 with headers in 1–2).
func sheetAppendRange(tab string, startRow int) string {
	tab = normalizeSheetTab(tab)
	escaped := strings.ReplaceAll(tab, "'", "''")
	return fmt.Sprintf("'%s'!A%d:D%d", escaped, startRow, startRow)
}
