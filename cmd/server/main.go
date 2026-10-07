// @title AutoCall API
// @version 1.0
// @description Outbound SIM800C calling API for CSV lead import, status, and interested leads.
// @BasePath /api/v1
package main

import (
	router "autocall/api/router"
	config "autocall/internal/config"
	database "autocall/internal/database"
	repository "autocall/internal/repository"
	service "autocall/internal/service"
	sim800c "autocall/pkg/caller/sim800c"
	logger "autocall/pkg/log"
	sheets "autocall/pkg/sheets"

	"net/http"
	"time"

	"github.com/ory/graceful"
)

func main() {
	logger.Start()

	logger.Info("🧩 Configuring environment")
	cfg, err := config.Load()
	logger.Fatal(err, "Failed to load env")

	logger.Info("📡 Bootstrapping database")
	gormDB := database.Init(&cfg.Database)
	sqlDB, err := gormDB.DB()
	logger.Fatal(err, "Failed to create connection with database")
	defer sqlDB.Close()

	leads := repository.NewLeads(gormDB)
	modem := sim800c.New(sim800c.Config{
		Port:      cfg.Modem.Port,
		BaudRate:  cfg.Modem.BaudRate,
		AudioFile: cfg.Modem.AudioFile,
	})

	caller := service.NewCaller(leads, modem)
	caller.Start()

	sheetClient := sheets.New(sheets.Config{
		CredentialsFile: cfg.Google.CredentialsFile,
		SpreadsheetID:   cfg.Google.SpreadsheetID,
		SheetName:       cfg.Google.SheetName,
	})
	service.StartSheetsSync(leads, sheetClient)

	server := graceful.WithDefaults(&http.Server{
		Addr:         cfg.Server.Port,
		Handler:      router.Config(gormDB, caller),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	logger.Info("🚀 Starting server on port %s", cfg.Server.Port)
	err = graceful.Graceful(server.ListenAndServe, server.Shutdown)
	logger.Fatal(err, "Failed to gracefully shutdown")
	logger.Info("🛑 Server exited")
}
