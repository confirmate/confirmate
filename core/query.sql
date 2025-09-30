-- name: GetTargetOfEvaluation :one
SELECT * FROM targets_of_evaluation
WHERE id = $1 LIMIT 1;

-- name: ListTargetOfEvaluation :many
SELECT * FROM targets_of_evaluation
ORDER BY name;

-- name: CreateTargetOfEvaluation :one
INSERT INTO targets_of_evaluation (
  name
) VALUES (
  $1
)
RETURNING *;

-- name: UpdateTargetOfEvaluation :exec
UPDATE targets_of_evaluation
set name = $1
WHERE id = $1;

-- name: DeleteTargetOfEvaluation :exec
DELETE FROM targets_of_evaluation
WHERE id = $1;