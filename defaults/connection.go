package defaults

import (
	"fmt"
	"sync/atomic"

	"github.com/aatuh/envvar"
	"github.com/aatuh/pureapi-core/database"
)

// Supported database drivers.
const (
	SQLite3 = "sqlite3"
	MySQL   = "mysql"
)

// DatabaseEnvConfig holds the names of environment variables for database
// connections.
type DatabaseEnvConfig struct {
	DatabaseDriver  string
	SQLiteName      string
	MySQLName       string
	MySQLDatabase   string
	MySQLUser       string
	MySQLPassword   string
	MySQLUseUnix    string
	MySQLHost       string
	MySQLPort       string
	MySQLSocketDir  string
	MySQLSocketName string
}

// dbEnvConfig holds the database environment variable config.
var dbEnvConfig atomic.Value

// init sets the default database environment variable config.
func init() {
	dbEnvConfig.Store(DatabaseEnvConfig{
		DatabaseDriver:  "DATABASE_DRIVER",
		SQLiteName:      "DATABASE_SQLITE_NAME",
		MySQLName:       "DATABASE_MYSQL_NAME",
		MySQLDatabase:   "DATABASE_MYSQL_DATABASE",
		MySQLUser:       "DATABASE_MYSQL_USER",
		MySQLPassword:   "DATABASE_MYSQL_PASSWORD",
		MySQLUseUnix:    "DATABASE_MYSQL_USE_UNIX_SOCKET",
		MySQLHost:       "DATABASE_MYSQL_HOST",
		MySQLPort:       "DATABASE_MYSQL_PORT",
		MySQLSocketDir:  "DATABASE_MYSQL_UNIX_SOCKET_DIRECTORY",
		MySQLSocketName: "DATABASE_MYSQL_UNIX_SOCKET_NAME",
	})
}

// SetDatabaseEnvConfig allows clients to override the default database env var
// names.
func SetDatabaseEnvConfig(cfg DatabaseEnvConfig) {
	dbEnvConfig.Store(cfg)
}

// GetDatabaseEnvConfig returns the current database env var configuration.
func GetDatabaseEnvConfig() DatabaseEnvConfig {
	return dbEnvConfig.Load().(DatabaseEnvConfig)
}

// GetConnection returns a database connection.
// If a configuration is provided it is used, otherwise a configuration is
// built from environment variables.
//
// Parameters:
//   - cfg: An optional connection configuration to use.
//
// Returns:
//   - DB: A database connection.
//   - error: An error if the connection fails.
func GetConnection(cfg ...*database.ConnectConfig) (database.DB, error) {
	dbDriver := GetDBDriverName(cfg...)
	switch dbDriver {
	case SQLite3:
		return GetSQLiteConnection(cfg...)
	case MySQL:
		return GetMySQLConnection(cfg...)
	default:
		return nil, fmt.Errorf(
			"GetConnection: unknown database type: %s", dbDriver,
		)
	}
}

// GetDBDriverName returns the name of the database driver.
// If a configuration is provided it is used, otherwise a configuration is
// built from environment variables.
//
// Parameters:
//   - cfg: An optional connection configuration to use.
//
// Returns:
//   - string: The name of the database driver.
func GetDBDriverName(cfg ...*database.ConnectConfig) string {
	if len(cfg) > 0 && cfg[0] != nil {
		return cfg[0].Driver
	}
	name := envvar.MustGet(GetDatabaseEnvConfig().DatabaseDriver)
	setCurrentDriverName(name)
	return name
}

// ConnectConfig returns a connection config based on environment variables.
// It wil determine the database driver based on an environment variable.
// If a configuration is provided it is used, otherwise a configuration is
// built from environment variables.
//
// Parameters:
//   - cfg: An optional connection configuration to use.
//
// Returns:
//   - *database.ConnectConfig: The connection configuration.
//   - error: An error if the connection fails.
func ConnectConfig() (*database.ConnectConfig, error) {
	envCfg := GetDatabaseEnvConfig()
	dbDriver := envvar.MustGet(envCfg.DatabaseDriver)
	provider, ok := getDriverProvider(dbDriver)
	if !ok || provider.NewConfigFromEnv == nil {
		return nil, fmt.Errorf(
			"ConnectConfig: no provider registered for driver: %s",
			dbDriver,
		)
	}
	return provider.NewConfigFromEnv(envCfg), nil
}

// GetMySQLConnection establishes a MySQL database connection.
// If a configuration is provided it is used, otherwise a configuration is
// built from environment variables.
//
// Parameters:
//   - cfg: An optional connection configuration to use.
//
// Returns:
//   - DB: A database connection.
//   - error: An error if the connection fails.
func GetMySQLConnection(cfg ...*database.ConnectConfig) (database.DB, error) {
	envCfg := GetDatabaseEnvConfig()
	var useCfg *database.ConnectConfig
	if len(cfg) > 0 && cfg[0] != nil {
		useCfg = cfg[0]
	} else {
		provider, ok := getDriverProvider(MySQL)
		if !ok || provider.NewConfigFromEnv == nil {
			return nil, fmt.Errorf(
				"GetMySQLConnection: no provider registered for driver: %s",
				MySQL,
			)
		}
		useCfg = provider.NewConfigFromEnv(envCfg)
	}
	provider, ok := getDriverProvider(MySQL)
	if !ok || provider.DSN == nil {
		return nil, fmt.Errorf(
			"GetMySQLConnection: no DSN function for driver: %s", MySQL,
		)
	}
	dsn, err := provider.DSN(*useCfg)
	if err != nil {
		return nil, fmt.Errorf("GetMySQLConnection: %w", err)
	}
	return database.Connect(*useCfg, database.NewSQLDBAdapter, dsn)
}

// GetSQLiteConnection establishes a SQLite database connection.
// If a configuration is provided it is used, otherwise a configuration is
// built from environment variables.
//
// Parameters:
//   - cfg: An optional connection configuration to use.
//
// Returns:
//   - DB: A database connection.
//   - error: An error if the connection fails.
func GetSQLiteConnection(cfg ...*database.ConnectConfig) (database.DB, error) {
	envCfg := GetDatabaseEnvConfig()
	var useCfg *database.ConnectConfig
	if len(cfg) > 0 && cfg[0] != nil {
		useCfg = cfg[0]
	} else {
		provider, ok := getDriverProvider(SQLite3)
		if !ok || provider.NewConfigFromEnv == nil {
			return nil, fmt.Errorf(
				"GetSQLiteConnection: no provider registered for driver: %s",
				SQLite3,
			)
		}
		useCfg = provider.NewConfigFromEnv(envCfg)
	}
	provider, ok := getDriverProvider(SQLite3)
	if !ok || provider.DSN == nil {
		return nil, fmt.Errorf(
			"GetSQLiteConnection: no DSN function for driver: %s", SQLite3,
		)
	}
	dsn, err := provider.DSN(*useCfg)
	if err != nil {
		return nil, fmt.Errorf("GetSQLiteConnection: %w", err)
	}
	return database.Connect(*useCfg, database.NewSQLDBAdapter, dsn)
}

// SQLiteConfig holds the configuration for a SQLite connection.
type SQLiteConfig struct {
	DatabaseName string
}

// ConnectConfig returns a connection config configured for SQLite.
//
// Returns:
//   - *database.ConnectConfig: The connection configuration.
func (cfg SQLiteConfig) ConnectConfig() *database.ConnectConfig {
	provider, ok := getDriverProvider(SQLite3)
	if !ok || provider.NewConfigFromEnv == nil {
		return &database.ConnectConfig{Driver: SQLite3, Database: cfg.DatabaseName}
	}
	// Build from provided fields without env; reusing env builder is acceptable
	// if the driver reads only fields.
	env := GetDatabaseEnvConfig()
	// Temporarily override via env-like struct.
	_ = env
	cc := provider.NewConfigFromEnv(env)
	cc.Driver = SQLite3
	cc.Database = cfg.DatabaseName
	return cc
}

// MySQLConfig holds the configuration for a MySQL connection.
type MySQLConfig struct {
	Database      string
	User          string
	Password      string
	UseUnixSocket bool
	Host          string
	Port          int
	SocketDir     string
	SocketName    string
}

// ConnectConfig returns a connection config configured for MySQL.
//
// Returns:
//   - *database.ConnectConfig: The connection configuration.
func (cfg MySQLConfig) ConnectConfig() *database.ConnectConfig {
	provider, ok := getDriverProvider(MySQL)
	if !ok || provider.NewConfigFromEnv == nil {
		return &database.ConnectConfig{Driver: MySQL, Database: cfg.Database}
	}
	// Build from provided fields similarly as above.
	env := GetDatabaseEnvConfig()
	_ = env
	cc := provider.NewConfigFromEnv(env)
	cc.Driver = MySQL
	cc.Database = cfg.Database
	cc.User = cfg.User
	cc.Password = cfg.Password
	cc.Host = cfg.Host
	cc.Port = cfg.Port
	if cfg.UseUnixSocket {
		cc.ConnectionType = "unix"
		cc.SocketDirectory = cfg.SocketDir
		cc.SocketName = cfg.SocketName
	} else {
		cc.ConnectionType = "tcp"
		if cc.Parameters == "" {
			cc.Parameters = "parseTime=true"
		}
	}
	return cc
}

// newMySQLConfigFromEnv creates a MySQLConfig from environment variables.
//
// Parameters:
//   - envCfg DatabaseEnvConfig: The environment variable names.
//
// Returns:
//   - MySQLConfig: The MySQL configuration.
func newMySQLConfigFromEnv(envCfg DatabaseEnvConfig) MySQLConfig {
	return MySQLConfig{
		Database:      envvar.MustGet(envCfg.MySQLDatabase),
		User:          envvar.MustGet(envCfg.MySQLUser),
		Password:      envvar.MustGet(envCfg.MySQLPassword),
		UseUnixSocket: envvar.MustGetBool(envCfg.MySQLUseUnix),
		Host:          envvar.MustGet(envCfg.MySQLHost),
		Port:          envvar.MustGetInt(envCfg.MySQLPort),
		SocketDir:     envvar.MustGet(envCfg.MySQLSocketDir),
		SocketName:    envvar.MustGet(envCfg.MySQLSocketName),
	}
}

// newSQLiteConfigFromEnv creates a SQLiteConfig from environment variables.
//
// Parameters:
//   - envCfg DatabaseEnvConfig: The environment variable names.
//
// Returns:
//   - SQLiteConfig: The SQLite configuration.
func newSQLiteConfigFromEnv(envCfg DatabaseEnvConfig) SQLiteConfig {
	return SQLiteConfig{
		DatabaseName: envvar.MustGet(envCfg.SQLiteName),
	}
}
