package query

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID           string
	ConnectionID uuid.UUID
	Statement    string
	Action       string
	Resource     string
	UserID       uuid.UUID
	CreatedAt    time.Time
	Options      Options
	ApprovalID   *uuid.UUID
}

type Options struct {
	MaxRows      int
	TimeoutMs    int
	ReadOnly     bool
	RequireWhere bool
}

type JobStore struct {
	mu   sync.Mutex
	jobs map[string]Job
	ttl  time.Duration
}

func NewJobStore(ttl time.Duration) *JobStore {
	return &JobStore{jobs: map[string]Job{}, ttl: ttl}
}

func (s *JobStore) Create(job Job) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	job.ID = uuid.NewString()
	job.CreatedAt = time.Now().UTC()
	s.jobs[job.ID] = job
	return job.ID
}

func (s *JobStore) Get(id string) (Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return Job{}, false
	}
	if time.Since(job.CreatedAt) > s.ttl {
		delete(s.jobs, id)
		return Job{}, false
	}
	return job, true
}

func (s *JobStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
}
