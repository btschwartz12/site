-- name: InsertFile :one
INSERT INTO
    files (uuid, url, notes, extension, expires)
VALUES
    (?, ?, ?, ?, ?)
RETURNING
    *;

-- name: GetFile :one
SELECT
    *
FROM
    files
WHERE
    uuid = ?;

-- name: UpdateFileExpires :one
UPDATE
    files
SET
    expires = ?
WHERE
    uuid = ?
RETURNING
    *;

-- name: GetAllFiles :many
SELECT
    *
FROM
    files;