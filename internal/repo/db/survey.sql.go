// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: survey.sql

package db

import (
	"context"
)

const getSurveyState = `-- name: GetSurveyState :one
SELECT
    data
FROM
    survey_state
WHERE
    id = 1
`

func (q *Queries) GetSurveyState(ctx context.Context) ([]byte, error) {
	row := q.db.QueryRowContext(ctx, getSurveyState)
	var data []byte
	err := row.Scan(&data)
	return data, err
}

const updateSurveyState = `-- name: UpdateSurveyState :exec
INSERT INTO survey_state (id, data, pit)
VALUES (1, ?, CURRENT_TIMESTAMP)
ON CONFLICT(id) DO UPDATE SET
    data = excluded.data,
    pit = CURRENT_TIMESTAMP
`

func (q *Queries) UpdateSurveyState(ctx context.Context, data []byte) error {
	_, err := q.db.ExecContext(ctx, updateSurveyState, data)
	return err
}
