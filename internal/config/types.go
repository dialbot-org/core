package config

type ServerConfig struct {
	Port    string
	Passkey string
}

type DatabaseConfig struct {
	SSL      string
	Name     string
	Host     string
	Port     string
	Password string
	Username string
}

type ModemConfig struct {
	Port      string
	BaudRate  int
	AudioFile string
}

type GoogleConfig struct {
	CredentialsFile string
	SpreadsheetID   string
	SheetName       string
}

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Modem    ModemConfig
	Google   GoogleConfig
}
