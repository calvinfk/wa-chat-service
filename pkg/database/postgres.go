package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Establishes a connection to the PostgreSQL database using GORM. The OpenPostgresConnection function takes a Data Source Name (DSN) as input, which contains the necessary information to connect to the database (such as host, port, user, password, and database name). It attempts to open a connection using GORM's postgres driver and returns a *gorm.DB instance if successful. If there is an error during the connection process, it panics with an error message indicating that the database connection failed.
func OpenPostgresConnection(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to open database: " + err.Error())
	}
	return db
}
