# Nodes

A node is the fundamental unit of content in WorkFlowy. Each node
represents a single bullet point that can contain text, have child
nodes, and be organized hierarchically. Nodes can be expanded or
collapsed, completed, tagged, and moved around to create flexible
outlines and lists.

{% examples %}

```endpoints
POST   /api/nodes
POST   /api/nodes/:id
GET    /api/nodes/:id
GET    /api/nodes
DELETE /api/nodes/:id
```

{% /examples %}


---

# The Node object

### Attributes

- **`id`** *string*

  Unique identifier of the node.

- **`name`** *string*

  The text content of the node. This is the main bullet text that appears in your outline. Supports [formatting](#formatting).

- **`note`** *string | null*

  Additional note content for the node. Notes appear below the main text and can contain extended descriptions or details.

- **`priority`** *number*

  Sort order of the node among its siblings. Lower values appear first.

- **`data.layoutMode`** *string*

  Display mode of the node. Common values include `"bullets"` (default), `"todo"`, `"h1"`, `"h2"`, `"h3"`.

- **`createdAt`** *number*

  Unix timestamp indicating when the node was created.

- **`modifiedAt`** *number*

  Unix timestamp indicating when the node was last modified.

- **`completedAt`** *number | null*

  Unix timestamp indicating when the node was completed. `null` if the node is not completed.

{% examples %}

```label
The Node object
```

```json
{
  "id": "6ed4b9ca-256c-bf2e-bd70-d8754237b505",
  "name": "This is a test outline for API examples",
  "note": null,
  "priority": 200,
  "data": {
    "layoutMode": "bullets"
  },
  "createdAt": 1753120779,
  "modifiedAt": 1753120850,
  "completedAt": null
}
```

{% /examples %}


---

# Create a node

### Parameters

- **`parent_id`** *string*

  The parent node identifier. Can be either a node UUID (e.g., `"6ed4b9ca-256c-bf2e-bd70-d8754237b505"`), a target key (e.g., `"home"`, `"inbox"`), or `"None"` to create a top-level node. See [List targets](#targets-list) to get your available target keys.

- **`name`** *string* *required*

  The text content of the new node. Supports [formatting](#formatting).

- **`note`** *string*

  Additional note content for the node. Notes appear below the main text.

- **`layoutMode`** *string*

  The display mode of the node. Common values include `"bullets"` (default), `"todo"`, `"h1"`, `"h2"`, `"h3"`.

- **`position`** *string*

  Where to place the new node. Options: `"top"` (default) or `"bottom"`.

## Formatting

The `name` field supports two formatting approaches: markdown syntax (auto-parsed) or direct HTML tags.
You can also set the `layoutMode` parameter explicitly to control node display style.

### Multiline text

When the `name` field contains multiple lines, the first line becomes the parent node and subsequent lines become child nodes.
A single `\n` is joined into a space; use `\n\n` (double newline) to create separate children.

### Layout modes

Set the node's display style using markdown syntax in `name` or the `layoutMode` parameter.

| Markdown | layoutMode | Description |
|---|---|---|
| `# text` | `"h1"` | Level 1 header |
| `## text` | `"h2"` | Level 2 header |
| `### text` | `"h3"` | Level 3 header |
| `#### text` | — | Bold text |
| `- text` | `"bullets"` | Bullet point |
| `- [ ] text` | `"todo"` | Uncompleted todo |
| `- [x] text` | `"todo"` | Completed todo |
| `` ```code``` `` | `"code-block"` | Code block |
| `> text` | `"quote-block"` | Quote block |

### Inline formatting

Style text within the `name` field using markdown or HTML tags.

| Markdown | HTML | Result |
|---|---|---|
| `**text**` | `<b>text</b>` | **bold** |
| `*text*` | `<i>text</i>` | *italic* |
| `~~text~~` | `<s>text</s>` | ~~strikethrough~~ |
| `` `text` `` | `<code>text</code>` | inline code |
| `[text](url)` | `<a href="url">text</a>` | hyperlink |

### Dates

Add dates to nodes using ISO 8601 format in square brackets. Invalid dates remain as plain text.

| Format | Description |
|---|---|
| `[YYYY-MM-DD]` | Date only |
| `[YYYY-MM-DD HH:MM]` | Date with time (24-hour) |

{% examples %}

```endpoint
POST /api/v1/nodes
```

```curl
curl -X POST https://workflowy.com/api/v1/nodes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_API_KEY>" \
  -d '{
    "parent_id": "inbox",
    "name": "Hello API",
    "position": "top"
  }' | jq
```

```response
{
  "item_id": "5b401959-4740-4e1a-905a-62a961daa8c9"
}
```

```endpoint
POST /api/v1/nodes
```

```curl
curl -X POST https://workflowy.com/api/v1/nodes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_API_KEY>" \
  -d '{
    "parent_id": "inbox",
    "name": "# Project\n\n- [ ] Task one\n\n- [ ] Task two"
  }' | jq
```

```endpoint
POST /api/v1/nodes
```

```curl
curl -X POST https://workflowy.com/api/v1/nodes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_API_KEY>" \
  -d '{
    "parent_id": "inbox",
    "name": "**Important**: See [docs](https://example.com)"
  }' | jq
```

```endpoint
POST /api/v1/nodes
```

```curl
curl -X POST https://workflowy.com/api/v1/nodes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_API_KEY>" \
  -d '{
    "parent_id": "inbox",
    "name": "Meeting [2025-12-15 14:30]"
  }' | jq
```

{% /examples %}


---

# Update a node

Updates the specified node by setting the values of the parameters passed. Any parameters not provided will be left unchanged.

### Parameters

- **`id`** *string* *required*

  The identifier of the node to update.

- **`name`** *string*

  The text content of the node. Supports [formatting](#formatting).

- **`note`** *string*

  The note content of the node.

- **`layoutMode`** *string*

  The display mode of the node.

{% examples %}

```endpoint
POST /api/v1/nodes/:id
```

```curl
curl -X POST https://workflowy.com/api/v1/nodes/:id \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_API_KEY>" \
  -d '{
    "name": "Updated node title"
  }' | jq
```

```response
{
  "status": "ok"
}
```

{% /examples %}


---

# Retrieve a node

Retrieves the details of an existing node. Supply the unique node ID and WorkFlowy will return the corresponding node information.

### Parameters

- **`id`** *string* *required*

  The identifier of the node to retrieve.

{% examples %}

```endpoint
GET /api/v1/nodes/:id
```

```curl
curl -X GET https://workflowy.com/api/v1/nodes/:id \
  -H "Authorization: Bearer <YOUR_API_KEY>" | jq
```

```response
{
  "node": {
    "id": "6ed4b9ca-256c-bf2e-bd70-d8754237b505",
    "name": "This is a test outline for API examples",
    "note": null,
    "priority": 200,
    "data": {
      "layoutMode": "bullets"
    },
    "createdAt": 1753120779,
    "modifiedAt": 1753120850,
    "completedAt": null
  }
}
```

{% /examples %}


---

# List nodes

Returns a list of child nodes for a given parent. The nodes are returned unordered - you need to sort them yourself based on the `priority` field.

### Parameters

- **`parent_id`** *string*

  The parent node identifier. Can be either a node UUID (e.g., `"6ed4b9ca-256c-bf2e-bd70-d8754237b505"`), a target key (e.g., `"home"`, `"inbox"`), or `"None"` to list top-level nodes. See [List targets](#targets-list) to get your available target keys.

{% examples %}

```endpoint
GET /api/v1/nodes
```

```curl
curl -G https://workflowy.com/api/v1/nodes \
  -H "Authorization: Bearer <YOUR_API_KEY>" \
  -d "parent_id=inbox" | jq
```

```response
{
  "nodes": [
    {
      "id": "ee1ac4c4-775e-1983-ae98-a8eeb92b1aca",
      "name": "Bullet A",
      "note": null,
      "priority": 100,
      "data": {
        "layoutMode": "bullets"
      },
      "createdAt": 1753120787,
      "modifiedAt": 1753120815,
      "completedAt": null
    }
  ]
}
```

{% /examples %}


---

# Delete a node

Permanently deletes a node. This cannot be undone.

### Parameters

- **`id`** *string* *required*

  The identifier of the node to delete.

{% examples %}

```endpoint
DELETE /api/v1/nodes/:id
```

```curl
curl -X DELETE https://workflowy.com/api/v1/nodes/:id \
  -H "Authorization: Bearer <YOUR_API_KEY>" | jq
```

```response
{
  "status": "ok"
}
```

{% /examples %}


---

# Move a node

### Parameters

- **`parent_id`** *string*

  The new parent node identifier. Can be either a node UUID (e.g., `"6ed4b9ca-256c-bf2e-bd70-d8754237b505"`), a target key (e.g., `"home"`, `"inbox"`), or `"None"` to move to top-level. See [List targets](#targets-list) to get your available target keys.

- **`position`** *string*

  Where to place the node. Options: `"top"` (default) or `"bottom"`.

{% examples %}

```endpoint
POST /api/v1/nodes/:id/move
```

```curl
curl -X POST https://workflowy.com/api/v1/nodes/<NODE_ID>/move \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_API_KEY>" \
  -d '{
    "parent_id": "inbox",
    "position": "top"
  }' | jq
```

```response
{
  "status": "ok"
}
```

{% /examples %}


---

# Complete a node

Marks a node as completed. This sets the completion timestamp and updates the node's status.

### Parameters

- **`id`** *string* *required*

  The identifier of the node to complete.

{% examples %}

```endpoint
POST /api/v1/nodes/:id/complete
```

```curl
curl -X POST https://workflowy.com/api/v1/nodes/:id/complete \
  -H "Authorization: Bearer <YOUR_API_KEY>" | jq
```

```response
{
  "status": "ok"
}
```

{% /examples %}


---

# Uncomplete a node

Marks a node as not completed. This removes the completion timestamp and updates the node's status.

### Parameters

- **`id`** *string* *required*

  The identifier of the node to uncomplete.

{% examples %}

```endpoint
POST /api/v1/nodes/:id/uncomplete
```

```curl
curl -X POST https://workflowy.com/api/v1/nodes/:id/uncomplete \
  -H "Authorization: Bearer <YOUR_API_KEY>" | jq
```

```response
{
  "status": "ok"
}
```

{% /examples %}


---

# Export all nodes

Returns all user's nodes as a flat list. Each node includes its `parent_id` field to reconstruct the hierarchy. The nodes are returned unordered - you need to build the tree structure yourself based on the `parent_id` and `priority` fields.

**Note:** This endpoint has a rate limit of 1 request per minute due to the potentially large response size.

{% examples %}

```endpoint
GET /api/v1/nodes-export
```

```curl
curl https://workflowy.com/api/v1/nodes-export \
  -H "Authorization: Bearer <YOUR_API_KEY>" | jq
```

```response
{
  "nodes": [
    {
      "id": "ee1ac4c4-775e-1983-ae98-a8eeb92b1aca",
      "name": "Top Level Item",
      "note": "This is a note",
      "parent_id": null,
      "priority": 100,
      "completed": false,
      "data": {
        "layoutMode": "bullets"
      },
      "createdAt": 1753120787,
      "modifiedAt": 1753120815,
      "completedAt": null
    },
    {
      "id": "ff2bd5d5-886f-2094-bf09-b9ffa93c2bdb",
      "name": "Child Item",
      "note": null,
      "parent_id": "ee1ac4c4-775e-1983-ae98-a8eeb92b1aca",
      "priority": 200,
      "completed": false,
      "data": {
        "layoutMode": "bullets"
      },
      "createdAt": 1753120820,
      "modifiedAt": 1753120830,
      "completedAt": null
    }
  ]
}
```

{% /examples %}


---

# Targets

Targets provide quick access to specific nodes in your outline. They include both system targets (like "inbox") and custom shortcuts you create (like "home").

Learn more about shortcuts in the [shortcuts documentation](https://workflowy.com/learn/shortcuts/).

{% examples %}

```endpoints
GET /api/v1/targets
```

{% /examples %}


---

# The Target object

### Attributes

- **`key`** *string*

  The unique identifier for this target (e.g., "home", "inbox", "today").

- **`type`** *string*

  The type of target:
  - `"shortcut"` - User-defined shortcuts.
  - `"system"` - System-managed locations like inbox. Always returned, even if the target node hasn't been created yet.

- **`name`** *string | null*

  The name of the node that this target points to. Returns `null` only for system targets when the target node hasn't been created yet.

{% examples %}

```label
User-defined shortcut
```

```json
{
  "key": "home",
  "type": "shortcut",
  "name": "My Home Page"
}
```

```label
System target (before node created)
```

```json
{
  "key": "inbox",
  "type": "system",
  "name": null
}
```

{% /examples %}


---

# List targets

Returns all available targets, including user-defined shortcuts (like "home") and system targets (like "inbox").

### Parameters

No parameters required.

{% examples %}

```endpoint
GET /api/v1/targets
```

```curl
curl https://workflowy.com/api/v1/targets \
  -H "Authorization: Bearer <YOUR_API_KEY>" | jq
```

```response
{
  "targets": [
    {
      "key": "home",
      "type": "shortcut",
      "name": "My Home Page"
    },
    {
      "key": "inbox",
      "type": "system",
      "name": "Inbox"
    }
  ]
}
```

{% /examples %}
