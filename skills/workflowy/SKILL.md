---
name: workflowy
description: "Interact with WorkFlowy via the wf CLI tool. Use when asked to create, read, update, delete, search, or organize WorkFlowy nodes, or manage a task/note workflow through the CLI."
---

# workflowy

Guide for agents to interact with WorkFlowy using the `wf` CLI tool.

## Authentication

Set the `WF_API_KEY` environment variable before running any command:

```bash
export WF_API_KEY=wf_your_api_key
```

Or authenticate once via `wf auth login` (interactive prompt or piped stdin).

## Output Format

Always pass `--json` for structured, parseable output. All JSON responses follow this envelope:

```json
{"success": true, "data": ...}
{"success": false, "error": "message", "type": "error_type"}
```

Use `--max-output N` to cap output length (rune-safe truncation, 0 = unlimited).

## Key Concepts

- **Node**: A WorkFlowy item (bullet). Has an ID (UUID), name, optional note, layout, and completion status.
- **Target**: A named location — `inbox`, `home`, or a shortcut. Use as parent refs.
- **Parent ref**: Either a node UUID or a target key like `inbox` or `home`.
- **Position**: `top` or `bottom` — where to place a node among siblings.
- **Layout modes**: `bullets`, `todo`, `h1`, `h2`, `h3`, `code-block`, `quote-block`.

## Commands

### Create a node

```bash
wf create "Buy groceries" --parent inbox --position top --json
```

Optional flags: `--parent`, `--note`, `--layout`, `--position`.

Returns: `{"success": true, "data": {"item_id": "..."}}`

### Get a node

```bash
wf get <node-id> --json
```

### List children

```bash
wf list --parent inbox --json
```

Without `--parent`, lists root-level children.

### Update a node

```bash
wf update <node-id> --name "New title" --note "Updated note" --json
```

Flags: `--name`, `--note`, `--layout`. Only changed fields are sent.

### Delete a node

```bash
wf delete <node-id> --json
```

Permanent deletion — no undo.

### Complete / Uncomplete

```bash
wf complete <node-id> --json
wf uncomplete <node-id> --json
```

### Move a node

```bash
wf move <node-id> --parent home --position bottom --json
```

### Search

```bash
wf search "keyword" --json
```

Searches node names and notes (case-insensitive). Uses a local cache (5 min TTL). Pass `--refresh` to force a fresh export from the API.

### Tree view

```bash
wf tree --depth 3 --json
```

Exports all nodes and builds a hierarchical tree. `--depth 0` = unlimited. Pass `--refresh` to bypass cache.

### List targets

```bash
wf targets --json
```

Returns available targets (inbox, home, shortcuts) with their keys and types.

## Workflow Patterns

### Create a task list under inbox

```bash
wf create "Project Tasks" --parent inbox --layout todo --json
# capture item_id from response, then:
wf create "Task 1" --parent <item_id> --json
wf create "Task 2" --parent <item_id> --json
```

### Find and update a node

```bash
wf search "meeting notes" --json
# pick the target node id from results
wf update <node-id> --note "Added action items" --json
```

### Organize: move completed items

```bash
wf list --parent inbox --json
# filter completed items from response (completedAt != null)
wf move <node-id> --parent home --json
```

## Error Handling

Check `success` field in JSON output. Common errors:
- **Not found**: invalid node ID.
- **Rate limited**: too many API calls — back off and retry.
- **Not authenticated**: missing or invalid API key.
