-- name: InsertPicture :one
INSERT INTO
    pictures (url, extension, description)
VALUES
    (?, ?, ?)
RETURNING
    *;

-- name: GetAllPictures :many
SELECT
    *
FROM
    pictures;

-- name: GetPicture :one
SELECT
    *
FROM
    pictures
WHERE
    id = ?;

-- name: DeletePicture :one
DELETE FROM
    pictures
WHERE
    id = ?
RETURNING
    url;