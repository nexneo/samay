# Samay Web

This document translates the existing Samay time-tracking experience into a standalone web application. It focuses on product capabilities, data contracts, and workflow expectations tailored for a browser-only build that stores data locally (no backend services).

## Product Goals

- Deliver a focused time tracker where individuals can start/stop timers, log work manually, and review history with minimal ceremony.
- Keep the mental model of “projects” owning all time entries, with analytics that summarize effort over weekly and monthly horizons.
- Maintain parity with the original Samay data semantics so workflows behave identically even though storage lives entirely in the browser.

## Core Concepts

- **Project**: Primary container for tracked work. Projects have unique names, optional company metadata, visibility flags, and timestamps for creation/update.
- **Entry**: Represents a block of time spent on a project. Each entry has duration in milliseconds, optional start/end timestamps, text description (supports inline `#tags`), billable flag, and type (`WORK`, `CHORE`, `FUN`).
- **Timer**: A project-scoped record that marks when active tracking began. Only one timer can exist per project at any time. Stopping a timer commits a new entry using the elapsed duration.
- **Tag**: Derived metadata extracted from hashtags in entry descriptions. Tags enable future filtering and reporting.
- **Person**: Optional lookup table for associating entries with contributors. The current feature set is single-user but the schema supports expansion.

## Data Model Expectations

Implement the following logical model inside the browser using IndexedDB (recommended), Web Storage, or another persistent client-side store that survives reloads:

- **Projects store**: metadata plus `is_hidden` flag and timestamps. Indexed by `updated_at` and project name for fast lookups.
- **Timers store**: keyed by `project_id`, containing `started_at`. Deleting a project cascades to its timer entry.
- **Entries store**: references a project and optional person identifier, tracks duration, timestamps, description, entry type, billable flag, and derived tags.
- **Entry tags store**: many-to-many join implemented as an array field on entries or a dedicated store keyed by `entry_id + tag`.
- **People store** (optional): unique email/name records ready for future multi-user features even though the initial release is single-user.

Persist timestamps in UTC (Unix seconds) and durations in milliseconds. Whenever entries change, update the associated project’s `updated_at` so list ordering stays meaningful.

## Feature Breakdown

### Project Management

- **Create project**: Trim whitespace, reject empty names, and block case-insensitive duplicates. Newly created projects should appear immediately in project listings.
- **Rename project**: Validate non-empty, non-duplicate names. Persist the change and refresh dependent views.
- **Delete project**: Require explicit confirmation. Deletion removes all related timers, entries, and tags via cascading rules. After deletion, redirect users to a safe default project list view.
- **List projects**: Present all projects sorted by most recent activity (`updated_at`). Indicate whether the project currently has an active timer.
- **Visibility**: Respect the `is_hidden` flag so hidden projects can be excluded from default lists while remaining queryable when needed.

### Time Tracking

- **Start timer**: Starting a timer records the current UTC timestamp in the `timers` table. If a timer already exists for the project, it should be replaced with the new start time. Timers are billable by default until the user toggles otherwise.
- **Stop timer**: Display the live elapsed duration inside the confirmation dialog, then prompt for a descriptive summary and billable toggle (defaulting to `true`). Calculate elapsed duration between stored start timestamp and current time (minimum zero). Persist a new entry with:
  - Trimmed description (allow empty but store `""` for consistency),
  - Duration in milliseconds,
  - Start/end timestamps and entry type `WORK`,
  - Billable flag,
  - Extracted hashtags saved into `entry_tags`.
    After persisting the entry, remove the timer row. Surface the duration captured so users get immediate feedback.
- **Timer status indicators**: When a timer exists for a project, show elapsed time and make stop actions primary. Hide or disable start actions until the timer is cleared.

### Manual Entry Capture

- Provide a form for users to manually log time with fields for duration, description, billable flag (default `true`), optional start/end override, and entry type (default to `WORK`).
- Accept human-readable duration strings (e.g., `45m`, `1h30m`). Validate and convert them into milliseconds.
- Reject empty duration or description values with inline error messaging.
- Save manual entries using the same persistence path as timer-based entries and update project ordering afterward.

### Entry Management

- **List entries**: Display a project’s entries sorted by recency (running entries first, then ended entries by end time descending). Allow text filtering and pagination for long histories.
- **Entry detail view**: Show normalized timestamps in the user’s locale, duration formatted as `H:MM`, billable state, entry type, and tag list. Render full multiline descriptions.
- **Modify entry**: Provide in-place editing so users can adjust description, billable flag (default `true` for new edits), duration, tags, or timestamps. Persist the changes atomically through the local data layer and update related project metadata.
- **Delete entry**: Require confirmation before removal. After deletion, refresh the list and project totals.
- **Move entry**: Allow reassigning an entry to another project. Present only valid target projects, perform client-side validation, and update both the entry and related project metadata atomically.
- **Tag usage**: Expose tag filters or search, leveraging local queries that reuse the same tag extraction and normalization logic used when saving entries.

### Reporting & Analytics

- **Weekly overview**: Summarize tracked duration for each project over the trailing seven days, month-to-date duration, total billable duration, and whether a timer is currently running. Order projects by weekly effort and highlight the currently selected project when applicable.
- **Monthly report**: For a chosen month, aggregate per-project totals (tracked vs billable time, entry counts) and indicate active timers (with elapsed time). Provide navigation to previous/next months and a quick return to the current month.
- **Totals formatting**: Use the helper that converts raw duration into `H:MM` strings to keep reporting consistent.
- **Future exports**: Provide CSV/JSON exports so users can back up or migrate their locally stored data.

### Configuration & Persistence

- **Local storage strategy**: Use IndexedDB (preferred for structured data) to store projects, entries, timers, tags, and optional people records. Provide a thin abstraction that mirrors the data model section, enabling swapping to other storage if needed.
- **Bootstrapping**: On first launch, seed the database with an empty dataset. Optionally offer import (JSON/CSV) so users can migrate existing records.
- **Backups & restore**: Expose export/import actions so users can copy data between browsers or devices since there is no server-side persistence.
- **Versioning**: Display the application version in a footer/about dialog and migrate local data stores when the schema evolves.
- **Graceful teardown**: Ensure pending writes flush before the tab closes (e.g., using `beforeunload`) and show clear status indicators while operations persist to IndexedDB.

### UX Guidelines

- Keep all primary actions within reach: start/stop timer, create manual entry, rename/delete project, view analytics.
- Provide confirmations for destructive actions and real-time validation for inputs (duration parsing, duplicate names, blank fields).
- Surface stopwatch-style feedback while a timer runs (e.g., mm:ss since start) so users understand elapsed time.
- Preserve keyboard accessibility and quick actions where feasible (e.g., shortcuts or command palette) without referencing terminal key bindings.
- Show explicit success/error states after operations like save, move, or delete to maintain user confidence.

### Technical Notes

- Build a lightweight data access layer that wraps IndexedDB transactions and exposes async CRUD/aggregation helpers mirroring the features above.
- Normalize time handling by storing and comparing in UTC, then formatting for the user’s locale at render time.
- Use deterministic UUID generation on the client (e.g., `crypto.randomUUID()`) and consistent duration formatting/tag extraction utilities to match the original behavior.
- Consider service workers or background sync if future enhancements require multi-tab coordination; for now, serialize writes per store to avoid conflicts.

## Implementation Sequence (Suggested)

1. **Storage abstraction**: Define IndexedDB stores, versioning, and upgrade logic that reflect the data model.
2. **State management**: Choose a client-side state solution (e.g., Redux, Zustand, Vuex) that syncs with IndexedDB and supports optimistic UI updates.
3. **Project dashboard**: Build the primary view that lists projects, timer states, and quick actions.
4. **Timer & manual entry forms**: Implement interactions for starting/stopping timers and capturing manual entries with validation and persistence.
5. **Entry management UI**: Create entry list, detail, edit/move/delete workflows tied to the local data layer.
6. **Reporting screens**: Implement weekly overview and monthly report aggregations client-side, caching intermediate results as needed.
7. **Backups & settings**: Surface export/import, version info, and reset flows.
8. **Testing & QA**: Add unit tests for data access, integration tests for flows (using mocked IndexedDB), and UI tests to guarantee parity with the original experience.
