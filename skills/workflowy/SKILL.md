---
name: workflowy
description: "Use when asked to interact with Workflowy via the wf CLI: create, read, list, update, move, complete, uncomplete, delete, search, or reorganize Workflowy nodes, inbox tasks, notes, projects, targets, or tree views."
---

# Workflowy

Use this skill to operate on a real Workflowy account through the `wf` CLI.

## When to Use

- Create or update Workflowy tasks, notes, projects, or outlines
- Inspect `inbox`, `home`, shortcuts, or a node subtree
- Search Workflowy content from the CLI
- Reorganize items by moving, completing, or deleting nodes
- Build quick automation steps around `wf --json`

## Operating Rules

1. Prefer `--json` unless the user explicitly wants human-readable output.
2. Treat delete as permanent. If the target is ambiguous, resolve it with `search`, `list`, or `get` before deleting.
3. Prefer target refs like `inbox` and `home` when they fit; otherwise use node UUIDs.
4. `wf search` and `wf tree` rely on the export API cache. Avoid repeated `--refresh` calls because the export endpoint is rate-limited to roughly **1 request/minute**.
5. If search or tree hits `429 Too Many Requests`, retry without `--refresh` first so the cached export can be reused.
6. Do not tell the user to log in again unless the CLI actually returns an authentication error.

## Authentication

The CLI works with either:

```bash
export WF_API_KEY=wf_your_api_key
```

or a previously saved login:

```bash
wf auth login
```

## JSON Envelope

When `--json` is used, responses follow this shape:

```json
{"success": true, "data": ...}
{"success": false, "error": "message", "type": "error_type"}
```

Useful global flags:

- `--json` for structured output
- `--max-output N` to cap text output size

## Core Commands

### Create

```bash
wf create "Buy groceries" --parent inbox --position top --layout todo --json
```

Optional flags: `--parent`, `--note`, `--layout`, `--position`

### Get

```bash
wf get <node-id> --json
```

### List children

```bash
wf list --parent inbox --json
```

If `--parent` is omitted, this lists top-level nodes.

### Update

```bash
wf update <node-id> --name "New title" --note "Updated note" --layout todo --json
```

### Move

```bash
wf move <node-id> --parent home --position bottom --json
```

### Complete / Uncomplete

```bash
wf complete <node-id> --json
wf uncomplete <node-id> --json
```

### Delete

```bash
wf delete <node-id> --json
```

### Tree

```bash
wf tree --depth 3 --json
wf tree --refresh --depth 3 --json
```

`--depth 0` means unlimited depth.

### Targets

```bash
wf targets --json
```

## Search

`wf search` uses cached export data and supports a Workflowy-style subset:

```bash
wf search 'project alpha' --json
wf search '@me OR @you' --json
wf search '-#done' --json
wf search '"exact phrase"' --json
wf search '#project > is:todo -is:complete' --json
wf search 'has:note OR is:code-block' --json
wf search 'created:7d' --json
wf search 'changed:24h is:todo' --json
```

Supported query features:

- implicit `AND`
- `OR`
- unary `-`
- quoted phrases
- nested ancestor search with `>`
- `is:todo`, `is:complete`, `is:bullets`, `is:h1`, `is:h2`, `is:h3`, `is:code-block`, `is:quote-block`
- `has:note`
- `created:<age>` and `changed:<age>` where age looks like `30m`, `12h`, `7d`, `2w`

Unsupported web-only search operators such as `text:`, `highlight:`, attachments, mirrors, backlinks, and sharing state return explicit errors because they are not present in export data.

## Recommended Procedures

### Find then update

```bash
wf search '"meeting notes"' --json
wf get <node-id> --json
wf update <node-id> --note "Added action items" --json
```

### Create a project under inbox

```bash
wf create "#project Launch plan" --parent inbox --json
wf create "Draft brief" --parent <project-id> --layout todo --json
wf create "Review assets" --parent <project-id> --layout todo --json
```

### Reorganize completed work

```bash
wf search 'is:complete' --json
wf move <node-id> --parent home --json
```

### Inspect structure safely

```bash
wf list --parent inbox --json
wf tree --depth 2 --json
```

Use `--refresh` only when cached export data is too stale for the task.

## Error Handling

Check `success` first in JSON mode.

Common failures:

- **Not authenticated**: missing or invalid API key
- **Not found**: bad node ID or parent ref
- **Rate limited**: too many export requests or rapid API activity
