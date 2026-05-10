package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS workspaces (
	id 				TEXT PRIMARY KEY,
	user_id 		TEXT NOT NULL,
	name 			TEXT NOT NULL,
	status 			TEXT NOT NULL,
	container_id 	TEXT NOT NULL DEFAULT '',
	port INTEGER 	NOT NULL DEFAULT 0,
	created_at 		DATETIME NOT NULL,
	updated_at 		DATETIME NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_workspaces_user_name
	ON workspaces(user_id, name)
	WHERE status != 'DELETED';

CREATE TABLE IF NOT EXISTS ports (
	port 			INTEGER PRIMARY KEY,
	workspace_id 	TEXT NOT NULL
);
`

func New(path string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("Open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Ping database: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("running schema: %w", err)
	}
	return db, nil
}