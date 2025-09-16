package main

import (
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/aatuh/pureapi-core/database"
	coreexamples "github.com/aatuh/pureapi-core/doc/examples"
	"github.com/aatuh/pureapi-framework/db"
	repositoryexamples "github.com/aatuh/pureapi-framework/doc/examples/repository"
)

// This example demonstrates the usage of the Query interface. It creates
// a database table using a custom SchemaManager implementation.
func main() {
	// Connect to the database.
	db, err := coreexamples.Connect(
		coreexamples.Cfg(), coreexamples.DummyConnectionOpen,
	)
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer db.Close()

	// Create the table.
	CreateTable(db)
}

// CreateTable creates the "users" table using the custom SchemaManager
// implementation. This demonstrates that you can use a custom query builder
// implementation and run custom SQL queries without the need for a repository.
//
// Parameters:
//   - dbh: The database handle.
func CreateTable(dbh database.DB) {
	schemaManager := &repositoryexamples.SimpleSchemaManager{}
	columns := []db.ColumnDefinition{
		{
			Name:          "id",
			Type:          "INTEGER",
			NotNull:       true,
			AutoIncrement: true,
			PrimaryKey:    true,
		},
		{
			Name:    "name",
			Type:    "TEXT",
			NotNull: true,
		},
	}
	createTableQuery, _, err := schemaManager.CreateTableQuery(
		"users", true, columns, nil, db.TableOptions{},
	)
	if err != nil {
		log.Printf("Create table query error: %v", err)
		return
	}
	if _, err = dbh.Exec(createTableQuery); err != nil {
		log.Printf("Create table execution error: %v", err)
		return
	}
	log.Println("Table 'users' created.")
}
