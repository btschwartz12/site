-- name: InsertPicture :one
INSERT INTO
    pictures (url, author, extension, description)
VALUES
    (?, ?, ?, ?)
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

-- name: AddLikeToPicture :exec
UPDATE
    pictures
SET
    num_likes = num_likes + 1
WHERE
    id = ?;

-- name: AddDislikeToPicture :exec
UPDATE
    pictures
SET
    num_dislikes = num_dislikes + 1
WHERE
    id = ?;

-- name: UpdateLikesDislikesOfPicture :one
UPDATE
    pictures
SET
    num_likes = ?,
    num_dislikes = ?
WHERE
    id = ?
RETURNING
    *;