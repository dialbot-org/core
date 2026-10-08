package repository

import (
	"time"

	model "core/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LeadRepository struct {
	db *gorm.DB
}

func NewLeads(db *gorm.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

type StatusCounts struct {
	Total      int64
	Pending    int64
	Calling    int64
	Completed  int64
	Interested int64
	Failed     int64
}

func (r *LeadRepository) InsertIgnoreDuplicates(leads []model.Lead) (inserted int, skipped int, err error) {
	for _, lead := range leads {
		result := r.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "phone_number"}},
			DoNothing: true,
		}).Create(&lead)

		if result.Error != nil {
			return inserted, skipped, result.Error
		}
		if result.RowsAffected == 0 {
			skipped++
		} else {
			inserted++
		}
	}
	return inserted, skipped, nil
}

func (r *LeadRepository) NextPending() (*model.Lead, error) {
	var lead model.Lead
	err := r.db.
		Where("status = ?", model.LeadStatusPending).
		Where("updated_at <= ?", time.Now()).
		Order("updated_at ASC").
		First(&lead).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

func (r *LeadRepository) Update(lead *model.Lead) error {
	return r.db.Save(lead).Error
}

// RequeueNoAnswer puts the lead back in the queue after retryAfter (not picked until then).
func (r *LeadRepository) RequeueNoAnswer(id uuid.UUID, retryAfter time.Duration) error {
	retryAt := time.Now().Add(retryAfter)
	return r.db.Model(&model.Lead{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     model.LeadStatusPending,
		"error":      nil,
		"updated_at": retryAt,
	}).Error
}

// ReleaseStaleCalling moves leads left in "calling" (e.g. after a crash or hung modem call)
// back to pending so the worker can pick them up again.
func (r *LeadRepository) ReleaseStaleCalling() (int64, error) {
	result := r.db.Model(&model.Lead{}).
		Where("status = ?", model.LeadStatusCalling).
		Updates(map[string]interface{}{
			"status": model.LeadStatusPending,
			"error":  nil,
		})
	return result.RowsAffected, result.Error
}

func (r *LeadRepository) SetStatus(id uuid.UUID, status string, errMsg *string) error {
	updates := map[string]interface{}{
		"status": status,
		"error":  errMsg,
	}
	return r.db.Model(&model.Lead{}).Where("id = ?", id).Updates(updates).Error
}

func (r *LeadRepository) MarkGoogleSheetSynced(id uuid.UUID) error {
	return r.db.Model(&model.Lead{}).
		Where("id = ?", id).
		Update("google_sheet_synced", true).Error
}

func (r *LeadRepository) StatusCounts() (StatusCounts, error) {
	var counts StatusCounts
	if err := r.db.Model(&model.Lead{}).Count(&counts.Total).Error; err != nil {
		return counts, err
	}
	statuses := map[string]*int64{
		model.LeadStatusPending:    &counts.Pending,
		model.LeadStatusCalling:    &counts.Calling,
		model.LeadStatusCompleted:  &counts.Completed,
		model.LeadStatusInterested: &counts.Interested,
		model.LeadStatusFailed:     &counts.Failed,
	}
	for status, target := range statuses {
		if err := r.db.Model(&model.Lead{}).Where("status = ?", status).Count(target).Error; err != nil {
			return counts, err
		}
	}
	return counts, nil
}

func (r *LeadRepository) CurrentCalling() (*model.Lead, error) {
	var lead model.Lead
	err := r.db.Where("status = ?", model.LeadStatusCalling).First(&lead).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

func (r *LeadRepository) Interested() ([]model.Lead, error) {
	var leads []model.Lead
	err := r.db.
		Where("status = ?", model.LeadStatusInterested).
		Order("updated_at DESC").
		Find(&leads).Error
	return leads, err
}

func (r *LeadRepository) UnsyncedInterested() ([]model.Lead, error) {
	var leads []model.Lead
	err := r.db.
		Where("status = ? AND google_sheet_synced = ?", model.LeadStatusInterested, false).
		Order("updated_at ASC").
		Find(&leads).Error
	return leads, err
}
