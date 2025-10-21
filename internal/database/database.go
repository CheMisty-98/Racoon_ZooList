package database

import (
	"Zoo_List/internal/config"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Db struct {
}

func ConnectDb(config config.Config) (*sql.DB, error) {
	dsn := config.DBUsername + ":" + config.DBPassword +
		"@tcp(" + config.DBHost + ":" + config.DBPort + ")/" + config.DBName +
		"?charset=utf8mb4&parseTime=True&loc=Local" // ← ВОТ ЭТО
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
