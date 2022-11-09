-- name: InsertTask :one
INSERT INTO tasks (name) 
VALUES ($1)
RETURNING *;

-- name: FindAllTasks :many
SELECT * FROM tasks
ORDER BY name;

-- name: FindTask :one
SELECT * FROM tasks
WHERE id = $1 LIMIT 1;

-- name: UpdateTask :exec
UPDATE tasks
  set name = $2,
  done = $3
WHERE id = $1;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;
