// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"database/sql"
	"time"
)

type Picture struct {
	ID          int64
	Url         string
	Description string
	Extension   string
	Pit         time.Time
}

type Visitor struct {
	ID      int64
	Path    string
	Message string
	Ip      sql.NullString
	City    sql.NullString
	Region  sql.NullString
	Country sql.NullString
	Pit     time.Time
}
