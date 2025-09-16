package main

import (
	"context"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/aatuh/pureapi-core/database"

	coreexamples "github.com/aatuh/pureapi-core/doc/examples"
	"github.com/aatuh/pureapi-framework/db"
	repositoryexamples "github.com/aatuh/pureapi-framework/doc/examples/repository"
)

// This example demonstrates the usage of the Repository interface. It creates
// a database table and inserts and retrieves a user using repositories and uses
// custom implementations of the Query and ErrorChecker interfaces to
// handle the database operations.
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
	// Insert a new user.
	CreateUser(db)
	// Retrieve the inserted user.
	GetUser(db)
}

// CreateTable creates the "users" table.
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

// CreateUser inserts a new user into the database using the custom QueryBuilder
// and ErrorChecker implementations.
//
// Parameters:
//   - dbh: The database handle.
func CreateUser(dbh database.DB) {
	mutatorRepo := db.NewMutatorRepo[*repositoryexamples.User](
		&repositoryexamples.SimpleQuery{},
		&repositoryexamples.SimpleErrorChecker{},
	)
	newUser := &repositoryexamples.User{Name: "Alice"}
	insertedUser, err := mutatorRepo.Insert(context.Background(), dbh, newUser)
	if err != nil {
		log.Printf("Insert error: %v", err)
		return
	}
	log.Printf("Inserted user: %+v\n", insertedUser)
}

// GetUser retrieves a user from the database using the custom QueryBuilder
// and ErrorChecker implementations.
//
// Parameters:
//   - dbh: The database handle.
func GetUser(dbh database.DB) {
	readerRepo := db.NewReaderRepo[*repositoryexamples.User](
		&repositoryexamples.SimpleQuery{},
		&repositoryexamples.SimpleErrorChecker{},
	)
	retrievedUser, err := readerRepo.GetOne(
		context.Background(),
		dbh,
		func() *repositoryexamples.User { return &repositoryexamples.User{} },
		nil,
	)
	if err != nil {
		log.Printf("GetOne error: %v", err)
		return
	}
	log.Printf("Retrieved user: %+v\n", retrievedUser)
}
