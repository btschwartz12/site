-- name: UpdateSurveyState :exec
INSERT INTO survey_state (id, data, pit)
VALUES (1, ?, CURRENT_TIMESTAMP)
ON CONFLICT(id) DO UPDATE SET
    data = excluded.data,
    pit = CURRENT_TIMESTAMP;

-- name: GetSurveyState :one
SELECT
    data
FROM
    survey_state
WHERE
    id = 1;