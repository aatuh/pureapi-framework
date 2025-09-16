package defaults

import (
	"fmt"
	"sync/atomic"

	"github.com/aatuh/pureapi-core/database"
	mysqlconfig "github.com/pureapi/pureapi-mysql/config"
	sqliteconfig "github.com/pureapi/pureapi-sqlite/config"
	"github.com/pureapi/pureapi-util/envvar"
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
	return envvar.MustGet(GetDatabaseEnvConfig().DatabaseDriver)
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
	switch dbDriver {
	case SQLite3:
		return newSQLiteConfigFromEnv(envCfg).ConnectConfig(), nil
	case MySQL:
		mysqlCfg := newMySQLConfigFromEnv(envCfg)
		return mysqlCfg.ConnectConfig(), nil
	default:
		return nil, fmt.Errorf(
			"ConnectConfig: unknown database type: %s", dbDriver,
		)
	}
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
		useCfg = newMySQLConfigFromEnv(envCfg).ConnectConfig()
	}
	dsn, err := mysqlconfig.DSN(*useCfg)
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
		useCfg = newSQLiteConfigFromEnv(envCfg).ConnectConfig()
	}
	dsn, err := sqliteconfig.DSN(*useCfg)
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
	return sqliteconfig.DefaultConfig(cfg.DatabaseName)
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
	var connectCfg *database.ConnectConfig
	if cfg.UseUnixSocket {
		connectCfg = mysqlconfig.DefaultUnixConfig(
			cfg.User, cfg.Password, cfg.Database, cfg.SocketDir, cfg.SocketName,
		)
	} else {
		connectCfg = mysqlconfig.DefaultTCPConfig(
			cfg.User, cfg.Password, cfg.Database,
		)
		connectCfg.Host = cfg.Host
		connectCfg.Port = cfg.Port
		connectCfg.Parameters = "parseTime=true"
	}
	connectCfg.Driver = "mysql"
	return connectCfg
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
