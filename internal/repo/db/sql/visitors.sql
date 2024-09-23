-- name: InsertVisitor :exec
INSERT INTO
    visitors (ip, path, message, city, region, country)
VALUES
    (?, ?, ?, ?, ?, ?);

-- name: GetAllVisitors :many
SELECT
    *
FROM
    visitors
