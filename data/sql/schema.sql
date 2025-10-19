PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;

CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL COLLATE NOCASE UNIQUE,
    company TEXT,
    is_hidden INTEGER NOT NULL DEFAULT 0 CHECK (is_hidden IN (0, 1)),
    position INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;

CREATE TABLE IF NOT EXISTS people (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL COLLATE NOCASE UNIQUE,
    name TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;

CREATE TABLE IF NOT EXISTS timers (
    project_id INTEGER PRIMARY KEY,
    started_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) STRICT, WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS entries (
    id TEXT PRIMARY KEY,
    project_id INTEGER NOT NULL,
    creator_id INTEGER,
    content TEXT NOT NULL DEFAULT '',
    duration_ms INTEGER NOT NULL CHECK (duration_ms >= 0),
    started_at INTEGER,
    ended_at INTEGER,
    entry_type TEXT NOT NULL DEFAULT 'WORK' CHECK (entry_type IN ('CHORE', 'FUN', 'WORK')),
    is_billable INTEGER NOT NULL DEFAULT 1 CHECK (is_billable IN (0, 1)),
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES people(id) ON DELETE SET NULL
) STRICT;

CREATE TABLE IF NOT EXISTS entry_tags (
    entry_id TEXT NOT NULL,
    tag TEXT NOT NULL COLLATE NOCASE,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    CHECK (tag = trim(tag) AND tag <> ''),
    FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE,
    PRIMARY KEY (entry_id, tag)
) STRICT, WITHOUT ROWID;

CREATE INDEX IF NOT EXISTS idx_projects_ordering ON projects(position ASC, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_entries_project ON entries(project_id);
CREATE INDEX IF NOT EXISTS idx_entries_project_ended ON entries(project_id, ended_at DESC, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_entries_creator ON entries(creator_id);
CREATE INDEX IF NOT EXISTS idx_entry_tags_tag ON entry_tags(tag);
