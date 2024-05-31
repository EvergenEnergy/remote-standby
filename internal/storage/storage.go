package storage

import (
	"log/slog"
	"sync"
	"time"
)

type Service struct {
	logger *slog.Logger

	mutex                 *sync.Mutex
	latestCommandReceived time.Time
}

func NewService(logger *slog.Logger) *Service {
	return &Service{
		logger:                logger,
		mutex:                 new(sync.Mutex),
		latestCommandReceived: time.Now(),
	}
}

func (s *Service) SetCommandTimestamp(setTime time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.latestCommandReceived = setTime
}

func (s *Service) GetCommandTimestamp() time.Time {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.latestCommandReceived
}
