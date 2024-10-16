-- name: InsertFile :one
INSERT INTO
    files (uuid, url, notes, extension)
VALUES
    (?, ?, ?, ?)
RETURNING
    *;

-- name: GetFile :one
SELECT
    *
FROM
    files
WHERE
    uuid = ?;

-- name: GetAllFiles :many
SELECT
    *
FROM
    files;

-- name: InsertPermalink :one
INSERT INTO
    permalinks (uuid, file_uuid, duration_seconds, expires)
VALUES
    (?, ?, ?, ?)
RETURNING
    *;

-- name: GetPermalink :one
SELECT
    *
FROM
    permalinks
WHERE
    uuid = ?;

-- name: GetAllPermalinks :many
SELECT
    *
FROM
    permalinks;