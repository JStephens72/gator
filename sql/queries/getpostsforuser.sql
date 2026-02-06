-- name: GetPostsForUser :many
SELECT posts.*, feeds.name AS feed_name FROM posts
INNER JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
INNER JOIN feeds ON feeds.id = feed_follows.feed_id
WHERE feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2;