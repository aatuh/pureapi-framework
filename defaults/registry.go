package defaults

import (
	"sync"
	"sync/atomic"

	"github.com/aatuh/pureapi-core/database"
	"github.com/aatuh/pureapi-framework/db"
)

// DriverProvider defines how a driver integrates with the framework.
//
// It provides a way to construct a ConnectConfig from env vars and to
// produce a DSN string for the given configuration.
type DriverProvider struct {
	NewConfigFromEnv func(DatabaseEnvConfig) *database.ConnectConfig
	DSN              func(database.ConnectConfig) (string, error)
}

var (
	providersMu   sync.RWMutex
	providersByID = map[string]DriverProvider{}

	queryFactoriesMu   sync.RWMutex
	queryFactoriesByID = map[string]func() db.Query{}

	errorFactoriesMu   sync.RWMutex
	errorFactoriesByID = map[string]func() database.ErrorChecker{}

	// Default driver name cache for fast, lock-free reads. This is a best
	// effort optimization for Query()/QueryErrorChecker() which often use
	// the current driver. Writes go through GetDBDriverName.
	currentDriverName atomic.Value // string
)

// RegisterDriver registers a database driver provider by name.
func RegisterDriver(name string, provider DriverProvider) {
	providersMu.Lock()
	defer providersMu.Unlock()
	providersByID[name] = provider
}

// getDriverProvider looks up a registered driver provider.
func getDriverProvider(name string) (DriverProvider, bool) {
	providersMu.RLock()
	defer providersMu.RUnlock()
	p, ok := providersByID[name]
	return p, ok
}

// RegisterQueryFactory registers a query factory for a specific driver.
func RegisterQueryFactory(driver string, f func() db.Query) {
	queryFactoriesMu.Lock()
	defer queryFactoriesMu.Unlock()
	queryFactoriesByID[driver] = f
}

// getQueryFactory returns a query factory for a specific driver.
func getQueryFactory(driver string) (func() db.Query, bool) {
	queryFactoriesMu.RLock()
	defer queryFactoriesMu.RUnlock()
	f, ok := queryFactoriesByID[driver]
	return f, ok
}

// RegisterErrorCheckerFactory registers an error checker factory for a
// specific driver.
func RegisterErrorCheckerFactory(
	driver string,
	f func() database.ErrorChecker,
) {
	errorFactoriesMu.Lock()
	defer errorFactoriesMu.Unlock()
	errorFactoriesByID[driver] = f
}

// getErrorCheckerFactory returns an error checker factory for a driver.
func getErrorCheckerFactory(driver string) (func() database.ErrorChecker, bool) {
	errorFactoriesMu.RLock()
	defer errorFactoriesMu.RUnlock()
	f, ok := errorFactoriesByID[driver]
	return f, ok
}

// setCurrentDriverName caches the latest detected driver name.
func setCurrentDriverName(name string) {
	currentDriverName.Store(name)
}

// getCurrentDriverName retrieves the cached current driver name.
func getCurrentDriverName() (string, bool) {
	v := currentDriverName.Load()
	if v == nil {
		return "", false
	}
	name, _ := v.(string)
	return name, name != ""
}
