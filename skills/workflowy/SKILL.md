---
name: workflowy
description: "Use when asked to interact with Workflowy via the wf CLI: create, read, list, update, move, complete, uncomplete, delete, search, or reorganize Workflowy nodes, inbox tasks, notes, projects, targets, or tree views."
---

# Workflowy

Operate on a real Workflowy account through the `wf` CLI.

## Operating Rules

1. Always use `--json` unless the user explicitly wants human-readable output.
2. Treat delete as permanent — resolve ambiguous targets with `search`, `list`, or `get` first.
3. Anywhere a node ID is accepted you can pass a UUID, a 12-character short ID, a Workflowy internal link (`https://workflowy.com/#/<short-id>`), or a target ref (`inbox`, `home`). Pass the user's raw reference through unchanged — the CLI resolves it. Use `wf id <ref>` when you need the canonical UUID.
4. **Rate limits**: `wf search` and `wf tree` rely on an export API cache. The export endpoint is rate-limited to ~1 req/min.
   - Never pass `--refresh` unless the user says data is stale or you just created/modified nodes and need fresh results.
   - On `429 Too Many Requests`, retry **without** `--refresh` to reuse the cache.
5. Do not suggest re-authentication unless the CLI actually returns an auth error.
6. For large trees, always set `--depth` (e.g. `--depth 2`) to avoid overwhelming output. `--depth 0` means unlimited.

## Authentication

```bash
export WF_API_KEY=wf_your_api_key   # env var
wf auth login                        # or interactive login
```

## Global Flags

| Flag | Purpose |
|------|---------|
| `--json` | Structured JSON output |
| `--max-output N` | Cap text output length (0 = unlimited) |

JSON envelope: `{"success": true, "data": ...}` or `{"success": false, "error": "...", "type": "..."}`.
Always check `success` before reading `data`.

## Commands

### create

```bash
wf create "<name>" [--parent <uuid|target>] [--note "..."] [--layout <mode>] [--position top|bottom] --json
```

Layout modes: `bullets`, `todo`, `h1`, `h2`, `h3`, `code-block`, `quote-block`.

### get

```bash
wf get <id> --json
```

### list vs tree

| | `wf list` | `wf tree` |
|---|-----------|-----------|
| Scope | Direct children only | Full subtree |
| Source | Live API | Cached export |
| Rate limit | No | Yes (~1/min with `--refresh`) |
| Key flag | `--parent <uuid\|target>` | `--depth N`, `--refresh` |

**Use `list`** when you only need one level of children (e.g. check inbox contents).
**Use `tree`** when you need the full hierarchy or deep structure.

```bash
wf list [--parent <uuid|target>] --json
wf tree [--depth N] [--refresh] --json
```

### update

```bash
wf update <id> [--name "..."] [--note "..."] [--layout <mode>] --json
```

### move

```bash
wf move <id> --parent <uuid|target> [--position top|bottom] --json
```

### complete / uncomplete

```bash
wf complete <id> --json
wf uncomplete <id> --json
```

### delete

```bash
wf delete <id> --json
```

### targets

```bash
wf targets --json
```

Returns available target refs (inbox, home, shortcuts).

### id

```bash
wf id <ref> --json
```

Resolves a UUID, 12-character short ID, internal link, or target ref to a canonical UUID. Use only when you specifically need the full UUID (e.g. for stable cross-session references); regular commands accept any of these forms directly.

## Search

Uses cached export data. Supports a Workflowy-style query subset:

```bash
wf search '<query>' [--refresh] --json
```

Query syntax:

| Syntax | Meaning |
|--------|---------|
| `word1 word2` | Implicit AND |
| `word1 OR word2` | Alternatives |
| `-term` | Exclude |
| `"exact phrase"` | Exact match |
| `ancestor > child` | Nested ancestor search |
| `is:todo`, `is:complete`, `is:bullets` | Layout/state filters |
| `is:h1`, `is:h2`, `is:h3` | Heading filters |
| `is:code-block`, `is:quote-block` | Block type filters |
| `has:note` | Has a note |
| `created:7d`, `changed:24h` | Age filters (`30m`, `12h`, `7d`, `2w`) |

**Unsupported** (web-only, not in export data): `text:`, `highlight:`, attachments, mirrors, backlinks, sharing state — these return explicit errors.

## Common Workflows

### Find → inspect → update

```bash
wf search '"meeting notes"' --json        # find by content
wf get <id> --json                         # inspect details
wf update <id> --note "Action items" --json
```

### Build a project in inbox

```bash
wf create "#project Launch" --parent inbox --json
# use returned item_id as parent for sub-tasks
wf create "Draft brief" --parent <project-id> --layout todo --json
wf create "Review assets" --parent <project-id> --layout todo --json
```

### Archive completed items

```bash
wf search 'is:complete' --json
wf move <id> --parent home --json
```

## Error Handling

Common `type` values in error responses:

- **Not authenticated** — missing or invalid API key / session
- **Not found** — bad node ID or unknown target ref
- **Rate limited** — too many export API requests; wait or drop `--refresh`
