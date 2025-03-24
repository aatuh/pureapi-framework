package custom

import (
	"fmt"
	"sync"

	"github.com/pureapi/pureapi-core/database"
	"github.com/pureapi/pureapi-core/database/types"
	mysqlconfig "github.com/pureapi/pureapi-mysql/config"
	sqliteconfig "github.com/pureapi/pureapi-sqlite/config"
	"github.com/pureapi/pureapi-util/envvar"

	_ "github.com/mattn/go-sqlite3"
)

const (
	SQLite3 = "sqlite3"
	MySQL   = "mysql"
)

const envVarDatabaseDriver = "DATABASE_DRIVER"

// Environment variable names for SQLite.
const (
	envVarSQLiteName = "DATABASE_SQLITE_NAME"
)

// Environment variable names for MySQL.
const (
	envVarMySQLName       = "DATABASE_MYSQL_NAME"
	envVarMySQLDatabase   = "DATABASE_MYSQL_DATABASE"
	envVarMySQLUser       = "DATABASE_MYSQL_USER"
	envVarMySQLPassword   = "DATABASE_MYSQL_PASSWORD"
	envVarMySQLUseUnix    = "DATABASE_MYSQL_USE_UNIX_SOCKET"
	envVarMySQLHost       = "DATABASE_MYSQL_HOST"
	envVarMySQLPort       = "DATABASE_MYSQL_PORT"
	envVarMySQLSocketDir  = "DATABASE_MYSQL_UNIX_SOCKET_DIRECTORY"
	envVarMySQLSocketName = "DATABASE_MYSQL_UNIX_SOCKET_NAME"
)

// connections stores the singleton instances of the database connections
// keyed by a string identifier.
var (
	connCounter int = 1
	connections     = make(map[int]types.DB)
	connMu      sync.Mutex
)

// GetConnectionKey returns a new unique key for a database connection.
func GetConnectionKey() int {
	connMu.Lock()
	defer connMu.Unlock()

	connCounter++
	return connCounter
}

// GetUniqueConnection returns a singleton instance of the database connection for the
// specified key. If the connection doesn't exist, it creates one.
func GetUniqueConnection(
	key int, cfg ...*database.ConnectConfig,
) (types.DB, error) {
	connMu.Lock()
	defer connMu.Unlock()

	if db, exists := connections[key]; exists {
		return db, nil
	}

	db, err := GetConnection(cfg...)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection for key %d: %w", key, err)
	}

	connections[key] = db
	return db, nil
}

func GetConnection(cfg ...*database.ConnectConfig) (types.DB, error) {
	dbDriver := envvar.MustGet(envVarDatabaseDriver)
	switch dbDriver {
	case SQLite3:
		var useCfg *database.ConnectConfig
		if len(cfg) > 0 && cfg[0] != nil {
			useCfg = cfg[0]
		} else {
			var err error
			useCfg, err = DefaultConnectConfig()
			if err != nil {
				return nil, err
			}
		}
		return GetSQLiteConnection(useCfg)
	case MySQL:
		return GetMySQLConnection()
	default:
		return nil, fmt.Errorf("unknown database type: %s", dbDriver)
	}
}

func DefaultConnectConfig() (*database.ConnectConfig, error) {
	dbDriver := envvar.MustGet(envVarDatabaseDriver)
	switch dbDriver {
	case SQLite3:
		return sqliteconfig.DefaultConfig(
			envvar.MustGet(envVarSQLiteName),
		), nil
	case MySQL:
		switch envvar.MustGetBool(envVarMySQLUseUnix) {
		case true:
			return mysqlconfig.DefaultUnixConfig(
				envvar.MustGet(envVarMySQLUser),
				envvar.MustGet(envVarMySQLPassword),
				envvar.MustGet(envVarMySQLDatabase),
				envvar.MustGet(envVarMySQLSocketDir),
				envvar.MustGet(envVarMySQLSocketName),
			), nil
		default:
			return mysqlconfig.DefaultTCPConfig(
				envvar.MustGet(envVarMySQLUser),
				envvar.MustGet(envVarMySQLPassword),
				envvar.MustGet(envVarMySQLName),
			), nil
		}
	default:
		return nil, fmt.Errorf("unknown database type: %s", dbDriver)
	}
}

// SQLiteConfig holds the configuration for a SQLite connection.
type SQLiteConfig struct {
	DatabaseName string
}

// GetConnectConfig returns a *database.ConnectConfig configured for SQLite.
func (cfg SQLiteConfig) GetConnectConfig() *database.ConnectConfig {
	connectCfg := sqliteconfig.DefaultConfig(cfg.DatabaseName)
	connectCfg.Driver = "sqlite3"
	connectCfg.Parameters = "_busy_timeout=5000&_foreign_keys=1&_journal_mode=WAL"
	return connectCfg
}

// GetSQLiteConnection establishes a SQLite database connection.
func GetSQLiteConnection(cfg *database.ConnectConfig) (types.DB, error) {
	dsn, err := sqliteconfig.DSN(*cfg)
	if err != nil {
		return nil, fmt.Errorf("GetSQLiteConnection: %w", err)
	}
	return database.Connect(*cfg, database.NewSQLDBAdapter, dsn)
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

// GetConnectConfig returns a *database.ConnectConfig configured for MySQL.
// Note: This example assumes you have helper functions in an
// "extendeddatabase" package for Unix and TCP configurations.
func (cfg MySQLConfig) GetConnectConfig() *database.ConnectConfig {
	var connectCfg *database.ConnectConfig
	if cfg.UseUnixSocket {
		// Use your own helper to generate a Unix socket DSN.
		connectCfg = mysqlconfig.DefaultUnixConfig(
			cfg.User, cfg.Password, cfg.Database, cfg.SocketDir, cfg.SocketName,
		)
	} else {
		// Use your own helper to generate a TCP DSN.
		connectCfg = mysqlconfig.DefaultTCPConfig(
			cfg.User, cfg.Password, cfg.Database,
		)
		connectCfg.Host = cfg.Host
		connectCfg.Port = cfg.Port
		// Example parameter for MySQL.
		connectCfg.Parameters = "parseTime=true"
	}
	connectCfg.Driver = "mysql"
	return connectCfg
}

// GetMySQLConnection establishes a MySQL database connection.
func GetMySQLConnection(cfg ...*database.ConnectConfig) (types.DB, error) {
	var useCfg *database.ConnectConfig
	if len(cfg) > 0 && cfg[0] != nil {
		useCfg = cfg[0]
	} else {
		// Read configuration from environment variables.
		mysqlCfg := MySQLConfig{
			Database:      envvar.MustGet(envVarMySQLName),
			User:          envvar.MustGet(envVarMySQLUser),
			Password:      envvar.MustGet(envVarMySQLPassword),
			UseUnixSocket: envvar.MustGetBool(envVarMySQLUseUnix),
			Host:          envvar.MustGet(envVarMySQLHost),
			Port:          envvar.MustGetInt(envVarMySQLPort),
			SocketDir:     envvar.MustGet(envVarMySQLSocketDir),
			SocketName:    envvar.MustGet(envVarMySQLSocketName),
		}
		useCfg = mysqlCfg.GetConnectConfig()
	}

	dsn, err := mysqlconfig.DSN(*useCfg)
	if err != nil {
		return nil, fmt.Errorf("GetMySQLConnection: %w", err)
	}
	return database.Connect(*useCfg, database.NewSQLDBAdapter, dsn)
}
