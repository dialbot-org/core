package service

import (
	"log"
	"sync"
	"time"

	model "autocall/internal/model"
	repository "autocall/internal/repository"
	sim800c "autocall/pkg/caller/sim800c"

	"github.com/google/uuid"
)

type Caller struct {
	repo    *repository.LeadRepository
	modem   *sim800c.Client
	mu      sync.Mutex
	current *uuid.UUID
}

func NewCaller(repo *repository.LeadRepository, modem *sim800c.Client) *Caller {
	return &Caller{repo: repo, modem: modem}
}

func (c *Caller) Start() {
	go func() {
		for {
			processed, err := c.processNext()
			if err != nil {
				log.Printf("caller: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}
			if processed {
				time.Sleep(1 * time.Second)
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}()
}

func (c *Caller) CurrentLeadID() *uuid.UUID {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.current
}

func (c *Caller) setCurrent(id *uuid.UUID) {
	c.mu.Lock()
	c.current = id
	c.mu.Unlock()
}

func (c *Caller) processNext() (bool, error) {
	lead, err := c.repo.NextPending()
	if err != nil {
		return false, err
	}
	if lead == nil {
		return false, nil
	}

	c.setCurrent(&lead.ID)

	lead.Status = model.LeadStatusCalling
	lead.Error = nil
	if err := c.repo.Update(lead); err != nil {
		c.setCurrent(nil)
		log.Printf("caller: mark calling: %v", err)
		return false, err
	}

	result := c.modem.Call(lead.PhoneNumber)
	status, errMsg := mapCallResult(result)

	lead.Status = status
	if errMsg != "" {
		lead.Error = &errMsg
	}
	if err := c.repo.Update(lead); err != nil {
		c.setCurrent(nil)
		log.Printf("caller: save result: %v", err)
		return true, err
	}

	c.setCurrent(nil)
	return true, nil
}

func mapCallResult(result sim800c.CallResult) (string, string) {
	switch result.Outcome {
	case sim800c.OutcomeInterested:
		return model.LeadStatusInterested, ""
	case sim800c.OutcomeCompleted:
		return model.LeadStatusCompleted, ""
	case sim800c.OutcomeBusy, sim800c.OutcomeNoAnswer, sim800c.OutcomeNoCarrier,
		sim800c.OutcomePlaybackFailed, sim800c.OutcomeModemError, sim800c.OutcomeSerialError:
		return model.LeadStatusFailed, result.Detail
	default:
		return model.LeadStatusFailed, result.Detail
	}
}
