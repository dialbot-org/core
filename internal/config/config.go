package config

import (
	"fmt"
	"strconv"

	"github.com/joho/godotenv"
)

func Load() (*Config, error) {
	godotenv.Load(".env")

	baud, _ := strconv.Atoi(GetEnv("MODEM_BAUD_RATE", "115200"))

	return &Config{
		Server: ServerConfig{
			Port:    ":" + GetEnv("PORT", "8000"),
			Passkey: GetEnv("PASSKEY", ""),
		},
		Database: DatabaseConfig{
			Name:     GetEnv("DB_NAME", "autocall"),
			Username: GetEnv("DB_USER", "postgres"),
			Password: GetEnv("DB_PASSWORD", "password"),
			Host:     GetEnv("DB_HOST", "localhost"),
			Port:     GetEnv("DB_PORT", "5432"),
			SSL:      GetEnv("DB_SSLMODE", "disable"),
		},
		Modem: ModemConfig{
			Port:      GetEnv("MODEM_PORT", "/dev/ttyUSB0"),
			BaudRate:  baud,
			AudioFile: GetEnv("MODEM_AUDIO_PATH", `C:\message.wav`),
		},
		Google: GoogleConfig{
			CredentialsFile: GetEnv("GOOGLE_CREDENTIALS_FILE", ""),
			SpreadsheetID:   GetEnv("GOOGLE_SPREADSHEET_ID", ""),
			SheetName:       GetEnv("GOOGLE_SHEET_NAME", "Leads"),
		},
	}, nil
}

func (cfg *DatabaseConfig) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Name,
		cfg.SSL,
	)
}

func (cfg *DatabaseConfig) GetAdminDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.SSL,
	)
}

func (cfg *DatabaseConfig) GetMigrationDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)
}
