package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/internal/interfaces"
	"github.com/coder-lulu/newbee-core/rpc/internal/provider"
	"github.com/coder-lulu/newbee-core/rpc/internal/validation"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusUnknown   HealthStatus = "unknown"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Component   string                        `json:"component"`
	Status      HealthStatus                  `json:"status"`
	Message     string                        `json:"message,omitempty"`
	Details     map[string]interface{}        `json:"details,omitempty"`
	LastChecked time.Time                     `json:"last_checked"`
	Duration    time.Duration                 `json:"duration"`
	Error       string                        `json:"error,omitempty"`
}

// ProviderHealthCheck represents health check for a specific provider
type ProviderHealthCheck struct {
	TenantID     uint64                        `json:"tenant_id"`
	ProviderType interfaces.OAuthProviderType `json:"provider_type"`
	HealthCheck
}

// HealthChecker provides health checking functionality for OAuth providers
type HealthChecker struct {
	mu              sync.RWMutex
	providerService *provider.ProviderService
	validator       *validation.ProviderValidator
	lastResults     map[string]*ProviderHealthCheck
	checkInterval   time.Duration
	ticker          *time.Ticker
	stopCh          chan struct{}
	running         bool
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(providerService *provider.ProviderService, validator *validation.ProviderValidator, checkInterval time.Duration) *HealthChecker {
	return &HealthChecker{
		providerService: providerService,
		validator:       validator,
		lastResults:     make(map[string]*ProviderHealthCheck),
		checkInterval:   checkInterval,
		stopCh:          make(chan struct{}),
	}
}

// CheckProviderHealth performs a health check on a specific provider
func (hc *HealthChecker) CheckProviderHealth(ctx context.Context, tenantID uint64, providerType interfaces.OAuthProviderType) *ProviderHealthCheck {
	start := time.Now()
	key := fmt.Sprintf("%d:%s", tenantID, providerType)

	check := &ProviderHealthCheck{
		TenantID:     tenantID,
		ProviderType: providerType,
		HealthCheck: HealthCheck{
			Component:   fmt.Sprintf("oauth_provider_%s", providerType),
			LastChecked: start,
		},
	}

	// Get provider configuration
	config, err := hc.providerService.GetProviderConfig(ctx, tenantID, providerType)
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = "Failed to retrieve provider configuration"
		check.Error = err.Error()
		check.Duration = time.Since(start)
		
		hc.mu.Lock()
		hc.lastResults[key] = check
		hc.mu.Unlock()
		
		return check
	}

	// Validate provider configuration
	validationResult := hc.validator.ValidateProvider(ctx, config, validation.ValidationLevelExtended)
	
	// Determine health status based on validation
	if !validationResult.Valid {
		check.Status = StatusUnhealthy
		check.Message = "Provider configuration validation failed"
		check.Details = map[string]interface{}{
			"validation_errors": validationResult.Errors,
			"validation_warnings": validationResult.Warnings,
		}
	} else if len(validationResult.Warnings) > 0 {
		check.Status = StatusDegraded
		check.Message = "Provider configuration has warnings"
		check.Details = map[string]interface{}{
			"validation_warnings": validationResult.Warnings,
		}
	} else {
		check.Status = StatusHealthy
		check.Message = "Provider is healthy"
	}

	// Add additional health details
	check.Details["enabled"] = config.Enabled
	check.Details["provider_type"] = config.Type
	check.Details["validation_duration"] = validationResult.Duration

	check.Duration = time.Since(start)

	// Cache the result
	hc.mu.Lock()
	hc.lastResults[key] = check
	hc.mu.Unlock()

	return check
}

// CheckAllProvidersHealth performs health checks on all providers for a tenant
func (hc *HealthChecker) CheckAllProvidersHealth(ctx context.Context, tenantID uint64) ([]*ProviderHealthCheck, error) {
	// Get all provider configurations for the tenant
	configs, err := hc.providerService.ListProviderConfigs(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list provider configs: %v", err)
	}

	checks := make([]*ProviderHealthCheck, 0, len(configs))
	for _, config := range configs {
		check := hc.CheckProviderHealth(ctx, tenantID, config.Type)
		checks = append(checks, check)
	}

	return checks, nil
}

// GetLastHealthCheck returns the last health check result for a provider
func (hc *HealthChecker) GetLastHealthCheck(tenantID uint64, providerType interfaces.OAuthProviderType) *ProviderHealthCheck {
	key := fmt.Sprintf("%d:%s", tenantID, providerType)
	
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	result, exists := hc.lastResults[key]
	if !exists {
		return nil
	}
	
	return result
}

// GetAllLastHealthChecks returns all cached health check results
func (hc *HealthChecker) GetAllLastHealthChecks() []*ProviderHealthCheck {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	checks := make([]*ProviderHealthCheck, 0, len(hc.lastResults))
	for _, check := range hc.lastResults {
		checks = append(checks, check)
	}
	
	return checks
}

// StartPeriodicHealthChecks starts periodic health checks
func (hc *HealthChecker) StartPeriodicHealthChecks(ctx context.Context, tenantIDs []uint64) {
	hc.mu.Lock()
	if hc.running {
		hc.mu.Unlock()
		return
	}
	hc.running = true
	hc.ticker = time.NewTicker(hc.checkInterval)
	hc.mu.Unlock()

	go func() {
		for {
			select {
			case <-hc.ticker.C:
				hc.performPeriodicHealthChecks(ctx, tenantIDs)
			case <-hc.stopCh:
				return
			}
		}
	}()
}

// StopPeriodicHealthChecks stops periodic health checks
func (hc *HealthChecker) StopPeriodicHealthChecks() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	
	if !hc.running {
		return
	}
	
	hc.running = false
	close(hc.stopCh)
	
	if hc.ticker != nil {
		hc.ticker.Stop()
	}
}

// performPeriodicHealthChecks performs health checks for all tenants
func (hc *HealthChecker) performPeriodicHealthChecks(ctx context.Context, tenantIDs []uint64) {
	for _, tenantID := range tenantIDs {
		// Run health checks for this tenant
		_, err := hc.CheckAllProvidersHealth(ctx, tenantID)
		if err != nil {
			// Log error but continue with other tenants
			// TODO: Add proper logging
		}
	}
}

// GetHealthSummary returns a summary of health status across all providers
func (hc *HealthChecker) GetHealthSummary() *HealthSummary {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	summary := &HealthSummary{
		TotalProviders: len(hc.lastResults),
		StatusCounts:   make(map[HealthStatus]int),
		LastUpdated:    time.Now(),
	}
	
	for _, check := range hc.lastResults {
		summary.StatusCounts[check.Status]++
		
		// Update last check time if this is the most recent
		if check.LastChecked.After(summary.LastChecked) {
			summary.LastChecked = check.LastChecked
		}
	}
	
	// Determine overall status
	if summary.StatusCounts[StatusUnhealthy] > 0 {
		summary.OverallStatus = StatusUnhealthy
	} else if summary.StatusCounts[StatusDegraded] > 0 {
		summary.OverallStatus = StatusDegraded
	} else if summary.StatusCounts[StatusHealthy] > 0 {
		summary.OverallStatus = StatusHealthy
	} else {
		summary.OverallStatus = StatusUnknown
	}
	
	return summary
}

// IsRunning returns whether periodic health checks are running
func (hc *HealthChecker) IsRunning() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.running
}

// SetCheckInterval updates the health check interval
func (hc *HealthChecker) SetCheckInterval(interval time.Duration) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	
	hc.checkInterval = interval
	
	// Restart ticker with new interval if running
	if hc.running && hc.ticker != nil {
		hc.ticker.Stop()
		hc.ticker = time.NewTicker(interval)
	}
}

// GetCheckInterval returns the current health check interval
func (hc *HealthChecker) GetCheckInterval() time.Duration {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.checkInterval
}

// ClearHealthChecks clears all cached health check results
func (hc *HealthChecker) ClearHealthChecks() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.lastResults = make(map[string]*ProviderHealthCheck)
}

// HealthSummary represents a summary of health status
type HealthSummary struct {
	OverallStatus   HealthStatus             `json:"overall_status"`
	TotalProviders  int                      `json:"total_providers"`
	StatusCounts    map[HealthStatus]int     `json:"status_counts"`
	LastChecked     time.Time                `json:"last_checked"`
	LastUpdated     time.Time                `json:"last_updated"`
}

// SystemHealthChecker performs system-wide health checks
type SystemHealthChecker struct {
	providerHealthChecker *HealthChecker
	components            map[string]func(context.Context) *HealthCheck
}

// NewSystemHealthChecker creates a new system health checker
func NewSystemHealthChecker(providerHealthChecker *HealthChecker) *SystemHealthChecker {
	shc := &SystemHealthChecker{
		providerHealthChecker: providerHealthChecker,
		components:            make(map[string]func(context.Context) *HealthCheck),
	}
	
	// Register default health check components
	shc.registerDefaultComponents()
	
	return shc
}

// RegisterComponent registers a health check component
func (shc *SystemHealthChecker) RegisterComponent(name string, checkFunc func(context.Context) *HealthCheck) {
	shc.components[name] = checkFunc
}

// CheckSystemHealth performs a comprehensive system health check
func (shc *SystemHealthChecker) CheckSystemHealth(ctx context.Context) map[string]*HealthCheck {
	results := make(map[string]*HealthCheck)
	
	// Check all registered components
	for name, checkFunc := range shc.components {
		results[name] = checkFunc(ctx)
	}
	
	return results
}

// registerDefaultComponents registers default health check components
func (shc *SystemHealthChecker) registerDefaultComponents() {
	// Provider health check component
	shc.RegisterComponent("providers", func(ctx context.Context) *HealthCheck {
		summary := shc.providerHealthChecker.GetHealthSummary()
		
		return &HealthCheck{
			Component:   "oauth_providers",
			Status:      summary.OverallStatus,
			Message:     fmt.Sprintf("Total providers: %d", summary.TotalProviders),
			Details:     map[string]interface{}{"summary": summary},
			LastChecked: time.Now(),
		}
	})
}

// Global health checker instances
var (
	globalHealthChecker       *HealthChecker
	globalSystemHealthChecker *SystemHealthChecker
)

// InitGlobalHealthChecker initializes the global health checker
func InitGlobalHealthChecker(providerService *provider.ProviderService, validator *validation.ProviderValidator, checkInterval time.Duration) {
	globalHealthChecker = NewHealthChecker(providerService, validator, checkInterval)
	globalSystemHealthChecker = NewSystemHealthChecker(globalHealthChecker)
}

// GetGlobalHealthChecker returns the global health checker instance
func GetGlobalHealthChecker() *HealthChecker {
	return globalHealthChecker
}

// GetGlobalSystemHealthChecker returns the global system health checker instance
func GetGlobalSystemHealthChecker() *SystemHealthChecker {
	return globalSystemHealthChecker
}