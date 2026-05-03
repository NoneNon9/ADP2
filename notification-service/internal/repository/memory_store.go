package repository

import "sync"

type InMemoryStore struct {
	processedEvents sync.Map
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{}
}

func (s *InMemoryStore) HasProcessed(eventID string) bool {
	_, exists := s.processedEvents.Load(eventID)
	return exists
}

func (s *InMemoryStore) MarkProcessed(eventID string) {
	s.processedEvents.Store(eventID, true)
}
