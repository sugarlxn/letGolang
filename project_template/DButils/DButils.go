package DButils

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DButils_package_variable = "DButils.package"

func ConnectToDB() (*sql.DB, error) {
	dbPath := "template.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}
