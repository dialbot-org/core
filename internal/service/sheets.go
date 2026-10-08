package service

import (
	"log"
	"time"

	repository "core/internal/repository"
	sheets "core/pkg/sheets"
)

func StartSheetsSync(repo *repository.LeadRepository, client *sheets.Client) {
	if client == nil || !client.Enabled() {
		log.Printf("sheets sync: disabled (missing configuration)")
		return
	}

	sync := func() {
		leads, err := repo.UnsyncedInterested()
		if err != nil {
			log.Printf("sheets sync: query: %v", err)
			return
		}

		for _, lead := range leads {
			if err := client.AppendLead(lead); err != nil {
				log.Printf("sheets sync: append %s: %v", lead.ID, err)
				continue
			}

			if err := repo.MarkGoogleSheetSynced(lead.ID); err != nil {
				log.Printf("sheets sync: mark synced %s: %v", lead.ID, err)
			}
		}
	}

	go func() {
		sync()

		ticker := time.NewTicker(40 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			sync()
		}
	}()
}
