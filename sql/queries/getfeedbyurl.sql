-- name: GetFeedByUrl :one
SELECT *
FROM feeds
WHERE URL=$1;