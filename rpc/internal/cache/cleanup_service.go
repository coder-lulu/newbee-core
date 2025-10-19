package cache

import (
	"context"
	"sync"
	"time"
)

// CleanupService provides background cleanup for cache entries
type CleanupService struct {
	cacheManager *CacheManager
	interval     time.Duration
	ticker       *time.Ticker
	stopCh       chan struct{}
	running      bool
	mu           sync.RWMutex
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(cacheManager *CacheManager, interval time.Duration) *CleanupService {
	return &CleanupService{
		cacheManager: cacheManager,
		interval:     interval,
		stopCh:       make(chan struct{}),
	}
}

// Start begins the background cleanup process
func (cs *CleanupService) Start() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.running {
		return
	}

	cs.running = true
	cs.ticker = time.NewTicker(cs.interval)

	go cs.run()
}

// Stop halts the background cleanup process
func (cs *CleanupService) Stop() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if !cs.running {
		return
	}

	cs.running = false
	close(cs.stopCh)
	
	if cs.ticker != nil {
		cs.ticker.Stop()
	}
}

// IsRunning returns whether the cleanup service is currently running
func (cs *CleanupService) IsRunning() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.running
}

// SetInterval updates the cleanup interval
func (cs *CleanupService) SetInterval(interval time.Duration) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.interval = interval
	
	if cs.running && cs.ticker != nil {
		cs.ticker.Stop()
		cs.ticker = time.NewTicker(interval)
	}
}

// GetInterval returns the current cleanup interval
func (cs *CleanupService) GetInterval() time.Duration {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.interval
}

// run executes the background cleanup loop
func (cs *CleanupService) run() {
	for {
		select {
		case <-cs.ticker.C:
			cs.performCleanup()
		case <-cs.stopCh:
			return
		}
	}
}

// performCleanup executes the actual cleanup process
func (cs *CleanupService) performCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Clean up expired cache entries
	expiredCount := cs.cacheManager.CleanupExpiredEntries(ctx)
	
	// TODO: Add proper logging
	_ = expiredCount // Prevent unused variable error for now
}

// ForceCleanup manually triggers a cleanup operation
func (cs *CleanupService) ForceCleanup() int {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return cs.cacheManager.CleanupExpiredEntries(ctx)
}

// Global cleanup service instance
var globalCleanupService *CleanupService

// InitGlobalCleanupService initializes the global cleanup service
func InitGlobalCleanupService(cacheManager *CacheManager, interval time.Duration) {
	globalCleanupService = NewCleanupService(cacheManager, interval)
}

// GetGlobalCleanupService returns the global cleanup service instance
func GetGlobalCleanupService() *CleanupService {
	return globalCleanupService
}

// StartGlobalCleanupService starts the global cleanup service
func StartGlobalCleanupService() {
	if globalCleanupService != nil {
		globalCleanupService.Start()
	}
}

// StopGlobalCleanupService stops the global cleanup service
func StopGlobalCleanupService() {
	if globalCleanupService != nil {
		globalCleanupService.Stop()
	}
}