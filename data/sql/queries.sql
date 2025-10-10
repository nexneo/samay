-- Projects

-- name: ListProjects :many
SELECT id,
       name,
       company,
       is_hidden,
       created_at,
       updated_at
FROM projects
ORDER BY name COLLATE NOCASE;

-- name: ListVisibleProjects :many
SELECT id,
       name,
       company,
       is_hidden,
       created_at,
       updated_at
FROM projects
WHERE is_hidden = 0
ORDER BY name COLLATE NOCASE;

-- name: GetProject :one
SELECT id,
       name,
       company,
       is_hidden,
       created_at,
       updated_at
FROM projects
WHERE id = ?1;

-- name: GetProjectByName :one
SELECT id,
       name,
       company,
       is_hidden,
       created_at,
       updated_at
FROM projects
WHERE name = ?1 COLLATE NOCASE;

-- name: CreateProject :one
INSERT INTO projects (name, company, is_hidden)
VALUES (?1, ?2, ?3)
RETURNING id,
          name,
          company,
          is_hidden,
          created_at,
          updated_at;

-- name: UpdateProject :one
UPDATE projects
SET name = ?2,
    company = ?3,
    is_hidden = ?4,
    updated_at = unixepoch()
WHERE id = ?1
RETURNING id,
          name,
          company,
          is_hidden,
          created_at,
          updated_at;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = ?1;


-- People

-- name: GetPerson :one
SELECT id,
       email,
       name,
       created_at,
       updated_at
FROM people
WHERE id = ?1;

-- name: GetPersonByEmail :one
SELECT id,
       email,
       name,
       created_at,
       updated_at
FROM people
WHERE email = ?1 COLLATE NOCASE;

-- name: UpsertPerson :one
INSERT INTO people (email, name)
VALUES (?1, ?2)
ON CONFLICT(email) DO UPDATE
SET name = excluded.name,
    updated_at = unixepoch()
RETURNING id,
          email,
          name,
          created_at,
          updated_at;


-- Timers

-- name: GetTimer :one
SELECT project_id,
       started_at,
       created_at,
       updated_at
FROM timers
WHERE project_id = ?1;

-- name: UpsertTimer :one
INSERT INTO timers (project_id, started_at)
VALUES (?1, ?2)
ON CONFLICT(project_id) DO UPDATE
SET started_at = excluded.started_at,
    updated_at = unixepoch()
RETURNING project_id,
          started_at,
          created_at,
          updated_at;

-- name: DeleteTimer :exec
DELETE FROM timers
WHERE project_id = ?1;


-- Entries

-- name: ListEntriesByProject :many
SELECT id,
       project_id,
       creator_id,
       content,
       duration_ms,
       started_at,
       ended_at,
       entry_type,
       is_billable,
       created_at,
       updated_at
FROM entries
WHERE project_id = ?1
ORDER BY ended_at IS NULL,
         ended_at DESC,
         started_at DESC,
         created_at DESC;

-- name: ListEntriesByTag :many
SELECT e.id,
       e.project_id,
       e.creator_id,
       e.content,
       e.duration_ms,
       e.started_at,
       e.ended_at,
       e.entry_type,
       e.is_billable,
       e.created_at,
       e.updated_at
FROM entries e
JOIN entry_tags t ON t.entry_id = e.id
WHERE t.tag = ?1 COLLATE NOCASE
ORDER BY e.ended_at IS NULL,
         e.ended_at DESC,
         e.started_at DESC,
         e.created_at DESC;

-- name: GetEntry :one
SELECT id,
       project_id,
       creator_id,
       content,
       duration_ms,
       started_at,
       ended_at,
       entry_type,
       is_billable,
       created_at,
       updated_at
FROM entries
WHERE id = ?1;

-- name: CreateEntry :one
INSERT INTO entries (
    id,
    project_id,
    creator_id,
    content,
    duration_ms,
    started_at,
    ended_at,
    entry_type,
    is_billable
) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)
RETURNING id,
          project_id,
          creator_id,
          content,
          duration_ms,
          started_at,
          ended_at,
          entry_type,
          is_billable,
          created_at,
          updated_at;

-- name: UpdateEntry :one
UPDATE entries
SET project_id = ?2,
    creator_id = ?3,
    content = ?4,
    duration_ms = ?5,
    started_at = ?6,
    ended_at = ?7,
    entry_type = ?8,
    is_billable = ?9,
    updated_at = unixepoch()
WHERE id = ?1
RETURNING id,
          project_id,
          creator_id,
          content,
          duration_ms,
          started_at,
          ended_at,
          entry_type,
          is_billable,
          created_at,
          updated_at;

-- name: DeleteEntry :exec
DELETE FROM entries
WHERE id = ?1;

-- name: ProjectTotalsInRange :one
SELECT COALESCE(SUM(duration_ms), 0) AS total_duration_ms,
       COALESCE(SUM(CASE WHEN is_billable = 1 THEN duration_ms ELSE 0 END), 0) AS billable_duration_ms,
       COUNT(*) AS entry_count
FROM entries
WHERE project_id = ?1
  AND ended_at IS NOT NULL
  AND ended_at >= ?2
  AND ended_at < ?3;


-- Entry Tags

-- name: ListTagsForEntry :many
SELECT tag,
       created_at
FROM entry_tags
WHERE entry_id = ?1
ORDER BY tag;

-- name: ListAllTags :many
SELECT DISTINCT tag
FROM entry_tags
ORDER BY tag;

-- name: InsertEntryTag :exec
INSERT INTO entry_tags (entry_id, tag)
VALUES (?1, ?2)
ON CONFLICT(entry_id, tag) DO NOTHING;

-- name: DeleteEntryTags :exec
DELETE FROM entry_tags
WHERE entry_id = ?1;
