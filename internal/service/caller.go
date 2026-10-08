package service

import (
	"log"
	"sync"
	"time"

	model "core/internal/model"
	repository "core/internal/repository"
	sim800c "core/pkg/caller/sim800c"

	"github.com/google/uuid"
)

type Caller struct {
	repo          *repository.LeadRepository
	modem         *sim800c.Client
	noAnswerRetry time.Duration
	mu            sync.Mutex
	current       *uuid.UUID
}

func NewCaller(repo *repository.LeadRepository, modem *sim800c.Client, noAnswerRetry time.Duration) *Caller {
	if noAnswerRetry <= 0 {
		noAnswerRetry = 6 * time.Hour
	}
	return &Caller{repo: repo, modem: modem, noAnswerRetry: noAnswerRetry}
}

func (c *Caller) Start() {
	if n, err := c.repo.ReleaseStaleCalling(); err != nil {
		log.Printf("caller: release stale calling leads: %v", err)
	} else if n > 0 {
		log.Printf("caller: reset %d lead(s) stuck in calling -> pending", n)
	}

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
	defer c.setCurrent(nil)

	lead.Status = model.LeadStatusCalling
	lead.Error = nil
	if err := c.repo.Update(lead); err != nil {
		log.Printf("caller: mark calling: %v", err)
		return false, err
	}

	result := c.modem.Call(lead.PhoneNumber, lead.FullName)
	status, errMsg := mapCallResult(result)

	if result.Outcome == sim800c.OutcomeNoAnswer {
		if err := c.repo.RequeueNoAnswer(lead.ID, c.noAnswerRetry); err != nil {
			log.Printf("caller: requeue no answer: %v", err)
			return true, err
		}
		log.Printf("caller: no answer for %q; pending again after %s", lead.FullName, c.noAnswerRetry)
		return true, nil
	}

	lead.Status = status
	if errMsg != "" {
		lead.Error = &errMsg
	}
	if err := c.repo.Update(lead); err != nil {
		log.Printf("caller: save result: %v", err)
		return true, err
	}

	return true, nil
}

func mapCallResult(result sim800c.CallResult) (string, string) {
	switch result.Outcome {
	case sim800c.OutcomeInterested:
		return model.LeadStatusInterested, ""
	case sim800c.OutcomeCompleted:
		return model.LeadStatusCompleted, ""
	case sim800c.OutcomeNoAnswer:
		return model.LeadStatusPending, ""
	case sim800c.OutcomeBusy, sim800c.OutcomeNoCarrier,
		sim800c.OutcomePlaybackFailed, sim800c.OutcomeModemError, sim800c.OutcomeSerialError,
		sim800c.OutcomeCallTimeout:
		return model.LeadStatusFailed, result.Detail
	default:
		return model.LeadStatusFailed, result.Detail
	}
}
