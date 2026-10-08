package controller

import (
	repository "core/internal/repository"
	service "core/internal/service"

	"gorm.io/gorm"
)

type Controller struct {
	db     *gorm.DB
	leads  *repository.LeadRepository
	caller *service.Caller
}

func Create(db *gorm.DB, caller *service.Caller) *Controller {
	return &Controller{
		db:     db,
		leads:  repository.NewLeads(db),
		caller: caller,
	}
}
