package persistence

import "database/sql"

func OpenDBConnection() (*sql.DB, error) {
	return sql.Open("superbot", "")
}
