package policies

import (
	"context"
	"sync/atomic"
	"time"

	"flowdb/backend/store"
)

type Store struct {
	store   *store.Store
	refresh time.Duration
	value   atomic.Value
}

func NewStore(st *store.Store, refresh time.Duration) *Store {
	return &Store{store: st, refresh: refresh}
}

func (s *Store) Start(ctx context.Context) error {
	if err := s.Refresh(ctx); err != nil {
		return err
	}
	ticker := time.NewTicker(s.refresh)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = s.Refresh(ctx)
			}
		}
	}()
	return nil
}

func (s *Store) Refresh(ctx context.Context) error {
	list, err := s.store.ListPolicies(ctx)
	if err != nil {
		return err
	}
	raw := make([][]byte, 0, len(list))
	for _, p := range list {
		raw = append(raw, p.Doc)
	}
	engine, err := NewEngine(raw)
	if err != nil {
		return err
	}
	s.value.Store(engine)
	return nil
}

func (s *Store) Engine() *Engine {
	val := s.value.Load()
	if val == nil {
		return &Engine{}
	}
	return val.(*Engine)
}
