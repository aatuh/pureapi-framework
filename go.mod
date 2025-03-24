module github.com/pureapi/pureapi-framework

go 1.24.0

require (
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pureapi/pureapi-core v1.0.0
	github.com/pureapi/pureapi-mysql v0.0.0-00010101000000-000000000000
	github.com/pureapi/pureapi-sqlite v0.0.0-00010101000000-000000000000
	github.com/pureapi/pureapi-util v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/pureapi/pureapi-core => ../pureapi-core

replace github.com/pureapi/pureapi-util => ../pureapi-util

replace github.com/pureapi/pureapi-sqlite => ../pureapi-sqlite

replace github.com/pureapi/pureapi-mysql => ../pureapi-mysql
