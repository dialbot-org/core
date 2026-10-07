package sheets

import (
	"context"
	"fmt"
	"log"
	"os"

	model "autocall/internal/model"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Config struct {
	CredentialsFile string
	SpreadsheetID   string
	SheetName       string
}

type Client struct {
	config    Config
	service    *sheets.Service
	ready  bool
}

func New(cfg Config) *Client {
	if cfg.SheetName == "" {
		cfg.SheetName = "Leads"
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
	c.ready = true
	return nil
}

func (c *Client) AppendLead(lead model.Lead) error {
	if err := c.initialize(); err != nil {
		return err
	}

	rangeName := fmt.Sprintf("%s!A:D", c.config.SheetName)
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
	).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()

	if err != nil {
		log.Printf("sheets: append lead %s: %v", lead.ID, err)
		return err
	}
	
	return nil
}
