# WorkFlowy Go SDK 设计分析

## 一、WorkFlowy API 资源模型分析

在动手设计 SDK 之前，我们先用 Google AIP-121 资源导向设计的视角来审视 WorkFlowy 的 API，把它的资源、方法和层级关系梳理清楚。

### 1.1 资源识别

WorkFlowy API 暴露了两个核心资源：

| 资源 | 说明 | 是否集合 |
| --- | --- | --- |
| **Node** | 大纲中的一个节点（bullet），是基本内容单元 | 是——节点本身形成树状集合 |
| **Target** | 快捷入口（inbox、home 等），指向特定节点 | 是——只读集合 |

从 AIP-121 的角度看，Node 是一个典型的**层级资源（hierarchical resource）**——每个 Node 既是资源本身，也是其子节点的集合。这与 Google API 中 `folders/123/documents/456` 这样的父子层级如出一辙，只不过 WorkFlowy 的层级可以无限嵌套。Target 则更像是一个辅助性的只读资源，类似于 Google API 中的 “locations” 或 “regions”——它为 Node 操作提供命名锚点。

### 1.2 方法映射

将 WorkFlowy API 的端点与 AIP 标准方法对照：

| AIP 标准方法 | WorkFlowy 端点 | 说明 |
| --- | --- | --- |
| **Create** (AIP-133) | `POST /api/nodes` | 创建节点，parent_id 指定父级 |
| **Get** (AIP-131) | `GET /api/nodes/: id` | 获取单个节点 |
| **List** (AIP-132) | `GET /api/nodes? parent_id=` | 列出子节点（需客户端按 priority 排序） |
| **Update** (AIP-134) | `POST /api/nodes/: id` | 更新节点属性（部分更新语义） |
| **Delete** (AIP-135) | `DELETE /api/nodes/: id` | 永久删除节点 |
| **Custom: Move** (AIP-136) | `POST /api/nodes/: id/move` | 移动节点到新父级 |
| **Custom: Complete** (AIP-136) | `POST /api/nodes/: id/complete` | 标记完成 |
| **Custom: Uncomplete** (AIP-136) | `POST /api/nodes/: id/uncomplete` | 取消完成 |
| **Custom: Export** (AIP-136) | `GET /api/nodes-export` | 导出全部节点（限流 1 次/分钟） |
| **List** (AIP-132) | `GET /api/targets` | 列出所有 Target |

这里有几个值得注意的设计特征：

**Move、Complete、Uncomplete 是典型的自定义方法。** 按照 AIP-136 的指导，它们是“无法用标准 CRUD 干净表达的操作”——Move 改变的是资源在层级中的位置而非属性本身，Complete/Uncomplete 是状态转换操作（AIP-216 中 state field 的理念：状态字段不应通过 Update 直接写入，而应通过专用的自定义方法触发）。

**Export 是一个集合级别的无状态方法。** 它不操作单个资源，而是将整个树扁平化导出。这对应 AIP-136 中的 “stateless methods” 或 AIP-159 中的 “aggregated list” 概念。

### 1.3 资源层级的特殊性

WorkFlowy 的 Node 层级有一个不寻常之处：`parent_id` 既可以是节点 UUID，也可以是 Target key（如 `"inbox"`、`"home"`），还可以是 `"None"` 表示顶层。这意味着 SDK 需要提供一种统一的方式来表达“父级”这个概念，而不是简单地用 string 类型一笔带过。

---

## 二、SDK 架构设计

### 2.1 整体结构：`client. Service. Operation` 模式

你提到希望采用 `client. Service. Operation` 的调用风格，这正是 Google 官方 Go 客户端库（`google-api-go-client`）的经典模式。以 Google Cloud Resource Manager 为例，它的调用链是：

```go
svc, _ := cloudresourcemanager.NewService(ctx)
project, _ := svc.Projects.Get("my-project").Do()
```

这个模式的核心思想是**三层结构**：

1. **Service（顶层客户端）**——持有 HTTP client、认证信息、base URL，并暴露各个资源的子服务

2. **ResourceService（资源服务）**——对应一类资源，暴露该资源上的所有操作方法

3. **Call（操作调用）**——每个操作返回一个 Call 对象，支持链式设置可选参数，最终通过 `. Do(ctx)` 执行

对于 WorkFlowy SDK，我建议的调用体验如下：

```go
// 创建客户端
client, err := workflowy.NewClient(
    workflowy.WithAPIKey("wf_xxx"),
)

// 标准 CRUD
node, err := client.Nodes.Get(ctx, "6ed4b9ca-...")
nodes, err := client.Nodes.List(ctx, workflowy.Parent("inbox"))
created, err := client.Nodes.Create(ctx, &workflowy.CreateNodeRequest{
    ParentID: workflowy.TargetInbox,
    Name:     "Hello API",
    Position: workflowy.PositionTop,
})
err = client.Nodes.Update(ctx, "6ed4b9ca-...", &workflowy.UpdateNodeRequest{
    Name: workflowy.String("Updated title"),
})
err = client.Nodes.Delete(ctx, "6ed4b9ca-...")

// 自定义方法
err = client.Nodes.Move(ctx, "6ed4b9ca-...", &workflowy.MoveNodeRequest{
    ParentID: workflowy.TargetInbox,
    Position: workflowy.PositionTop,
})
err = client.Nodes.Complete(ctx, "6ed4b9ca-...")
err = client.Nodes.Uncomplete(ctx, "6ed4b9ca-...")

// 导出（集合级别操作）
allNodes, err := client.Nodes.Export(ctx)

// Targets
targets, err := client.Targets.List(ctx)
```

### 2.2 为什么选择简化的 Call 模式

Google 的 `google-api-go-client` 使用了一个重量级的 Call 对象模式：每个操作返回一个 `*XxxCall`，通过链式方法设置参数，最后调用 `. Do()`。这种模式的优势在于可以处理极其复杂的 API（几十个可选参数、分页、字段掩码等），但代价是类型爆炸——每个操作都需要一个独立的 Call 类型。

WorkFlowy 的 API 相当精简（总共约 10 个端点，每个端点最多 4-5 个参数），如果照搬完整的 Call 模式会显得过度工程化。我的建议是采用**简化版**：直接在方法签名中接受 Request struct 或必需参数，可选参数通过 Request struct 的零值/指针语义处理。这样既保持了 `client. Nodes. Get(...)` 的调用风格，又避免了不必要的复杂度。

不过，如果你更偏好完整的 Call 链式风格（为了未来扩展性或风格一致性），也完全可以实现：

```go
// 完整 Call 模式（备选方案）
node, err := client.Nodes.Get("6ed4b9ca-...").
    Fields("name", "note").  // 未来可能支持的字段选择
    Do(ctx)

created, err := client.Nodes.Create(&workflowy.Node{Name: "Hello"}).
    ParentID(workflowy.TargetInbox).
    Position(workflowy.PositionTop).
    Do(ctx)
```

---

## 三、类型系统设计

### 3.1 资源类型

```go
package workflowy

import "time"

// Node 是 WorkFlowy 的核心资源，代表大纲中的一个节点。
type Node struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    Note        *string    `json:"note"`          // nullable
    Priority    int        `json:"priority"`
    Data        NodeData   `json:"data"`
    CreatedAt   Timestamp  `json:"createdAt"`
    ModifiedAt  Timestamp  `json:"modifiedAt"`
    CompletedAt *Timestamp `json:"completedAt"`   // nullable

    // 仅在 Export 响应中出现
    ParentID    *string    `json:"parent_id,omitempty"`
    Completed   *bool      `json:"completed,omitempty"`
}

type NodeData struct {
    LayoutMode LayoutMode `json:"layoutMode"`
}

// Target 表示一个快捷入口。
type Target struct {
    Key  string     `json:"key"`
    Type TargetType `json:"type"`
    Name *string    `json:"name"` // system target 未创建时为 null
}
```

这里有几个关键的类型设计决策值得展开讨论：

**nullable 字段用指针表示。** `Note`、`CompletedAt`、`Name`（Target 的）在 API 中可以是 `null`。Go 中惯用的做法是用指针类型来区分“零值”和“未设置/null”。这与 AIP-134 中 Update 方法的部分更新语义直接相关——你需要区分“用户想把 note 清空”和“用户没有传 note 字段”。

**自定义 Timestamp 类型。** WorkFlowy 返回的是 Unix 时间戳（整数），而不是 RFC 3339 字符串。定义一个 `Timestamp` 类型封装 `time. Time`，实现自定义的 JSON 序列化/反序列化，可以让用户在代码中直接使用 `time. Time` 的方法，同时正确处理 wire format。

### 3.2 枚举与常量

```go
// LayoutMode 定义节点的显示模式。
type LayoutMode string

const (
    LayoutBullets    LayoutMode = "bullets"
    LayoutTodo       LayoutMode = "todo"
    LayoutH1         LayoutMode = "h1"
    LayoutH2         LayoutMode = "h2"
    LayoutH3         LayoutMode = "h3"
    LayoutCodeBlock  LayoutMode = "code-block"
    LayoutQuoteBlock LayoutMode = "quote-block"
)

// TargetType 定义 Target 的类型。
type TargetType string

const (
    TargetTypeShortcut TargetType = "shortcut"
    TargetTypeSystem   TargetType = "system"
)

// Position 定义节点的插入位置。
type Position string

const (
    PositionTop    Position = "top"
    PositionBottom Position = "bottom"
)

// ParentRef 统一表达"父级"的概念。
// 可以是节点 UUID、Target key 或 None（顶层）。
type ParentRef string

const (
    ParentNone    ParentRef = "None"
    TargetHome    ParentRef = "home"
    TargetInbox   ParentRef = "inbox"
)

// NodeParent 从节点 ID 构造 ParentRef。
func NodeParent(id string) ParentRef {
    return ParentRef(id)
}
```

`ParentRef` 的设计是一个关键抉择。WorkFlowy 的 `parent_id` 是一个多态字段——它可以是 UUID、target key 或字面量 `"None"`。直接用 `string` 虽然简单，但丢失了类型安全性和自文档化能力。用一个命名类型加上预定义常量和构造函数，既保持了灵活性，又让调用者在 IDE 中能看到所有合法选项。

### 3.3 请求/响应类型

```go
// CreateNodeRequest 对应 Create 操作的参数。
type CreateNodeRequest struct {
    ParentID   ParentRef  `json:"parent_id,omitempty"`
    Name       string     `json:"name"`
    Note       string     `json:"note,omitempty"`
    LayoutMode LayoutMode `json:"layoutMode,omitempty"`
    Position   Position   `json:"position,omitempty"`
}

// UpdateNodeRequest 对应 Update 操作的参数。
// 使用指针类型实现部分更新语义——nil 表示不修改该字段。
type UpdateNodeRequest struct {
    Name       *string     `json:"name,omitempty"`
    Note       *string     `json:"note,omitempty"`
    LayoutMode *LayoutMode `json:"layoutMode,omitempty"`
}

// MoveNodeRequest 对应 Move 自定义方法的参数。
type MoveNodeRequest struct {
    ParentID ParentRef `json:"parent_id,omitempty"`
    Position Position  `json:"position,omitempty"`
}

// ListNodesOptions 对应 List 操作的查询参数。
type ListNodesOptions struct {
    ParentID ParentRef
}
```

UpdateNodeRequest 中全部使用指针字段，这是 Go SDK 处理 PATCH 语义的标准做法。AIP-134 强调 Update 应该是 PATCH 而非 PUT——只修改显式传入的字段。指针为 nil 意味着“不碰这个字段”，指针指向零值（如空字符串）意味着“把这个字段清空”。为了方便用户构造指针值，提供辅助函数：

```go
// 辅助函数，方便构造指针值
func String(v string) *string { return &v }
func Layout(v LayoutMode) *LayoutMode { return &v }
```

---

## 四、客户端内部架构

### 4.1 核心结构

```go
package workflowy

import (
    "context"
    "net/http"
)

// Client 是 WorkFlowy API 的顶层客户端。
type Client struct {
    // 子服务——对外暴露
    Nodes   *NodesService
    Targets *TargetsService

    // 内部基础设施
    httpClient *http.Client
    baseURL    string
    apiKey     string
    userAgent  string
}

// NewClient 创建一个新的 WorkFlowy 客户端。
func NewClient(opts ...Option) (*Client, error) {
    c := &Client{
        httpClient: http.DefaultClient,
        baseURL:    "https://workflowy.com/api/v1",
        userAgent:  "workflowy-go/0.1.0",
    }
    for _, opt := range opts {
        opt(c)
    }
    if c.apiKey == "" {
        return nil, ErrMissingAPIKey
    }

    // 初始化子服务，共享同一个内部 caller
    s := &service{client: c}
    c.Nodes = &NodesService{s: s}
    c.Targets = &TargetsService{s: s}

    return c, nil
}

// service 是所有子服务共享的内部基础设施。
type service struct {
    client *Client
}
```

### 4.2 Functional Options 模式

```go
type Option func(*Client)

func WithAPIKey(key string) Option {
    return func(c *Client) { c.apiKey = key }
}

func WithHTTPClient(hc *http.Client) Option {
    return func(c *Client) { c.httpClient = hc }
}

func WithBaseURL(url string) Option {
    return func(c *Client) { c.baseURL = url }
}

func WithUserAgent(ua string) Option {
    return func(c *Client) { c.userAgent = ua }
}
```

Functional Options 是 Go 社区公认的客户端配置最佳实践。相比 Config struct，它的优势在于零值即合理默认值、向后兼容地添加新选项、以及调用侧的可读性。Google Cloud Go 客户端库（`cloud.google.com/go`）就采用了这种模式（`option. WithAPIKey`、`option. WithHTTPClient` 等）。

### 4.3 内部 HTTP 层

```go
// do 是所有 API 调用的统一出口。
func (s *service) do(ctx context.Context, req *http.Request, v interface{}) error {
    req.Header.Set("Authorization", "Bearer "+s.client.apiKey)
    req.Header.Set("User-Agent", s.client.userAgent)
    if req.Body != nil {
        req.Header.Set("Content-Type", "application/json")
    }

    resp, err := s.client.httpClient.Do(req.WithContext(ctx))
    if err != nil {
        return fmt.Errorf("workflowy: request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return parseAPIError(resp)
    }

    if v != nil {
        if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
            return fmt.Errorf("workflowy: decode response: %w", err)
        }
    }
    return nil
}
```

所有的认证、请求头注入、错误处理、响应解码都集中在这一个方法里。子服务只需要构造 `*http. Request` 并调用 `s.do()`，不需要关心这些横切关注点。

---

## 五、子服务实现

### 5.1 NodesService

```go
type NodesService struct {
    s *service
}

func (ns *NodesService) Get(ctx context.Context, id string) (*Node, error) {
    req, err := ns.s.newRequest("GET", "/nodes/"+id, nil)
    if err != nil {
        return nil, err
    }
    var resp struct {
        Node *Node `json:"node"`
    }
    if err := ns.s.do(ctx, req, &resp); err != nil {
        return nil, err
    }
    return resp.Node, nil
}

func (ns *NodesService) List(ctx context.Context, opts *ListNodesOptions) ([]*Node, error) {
    params := url.Values{}
    if opts != nil && opts.ParentID != "" {
        params.Set("parent_id", string(opts.ParentID))
    }
    req, err := ns.s.newRequest("GET", "/nodes?"+params.Encode(), nil)
    if err != nil {
        return nil, err
    }
    var resp struct {
        Nodes []*Node `json:"nodes"`
    }
    if err := ns.s.do(ctx, req, &resp); err != nil {
        return nil, err
    }
    // 按 AIP-132 的精神，List 应返回有序结果
    // WorkFlowy API 明确说返回无序，需客户端排序
    sort.Slice(resp.Nodes, func(i, j int) bool {
        return resp.Nodes[i].Priority < resp.Nodes[j].Priority
    })
    return resp.Nodes, nil
}

func (ns *NodesService) Create(ctx context.Context, r *CreateNodeRequest) (*CreateNodeResponse, error) {
    req, err := ns.s.newRequest("POST", "/nodes", r)
    if err != nil {
        return nil, err
    }
    var resp CreateNodeResponse
    if err := ns.s.do(ctx, req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (ns *NodesService) Update(ctx context.Context, id string, r *UpdateNodeRequest) error {
    req, err := ns.s.newRequest("POST", "/nodes/"+id, r)
    if err != nil {
        return err
    }
    return ns.s.do(ctx, req, nil)
}

func (ns *NodesService) Delete(ctx context.Context, id string) error {
    req, err := ns.s.newRequest("DELETE", "/nodes/"+id, nil)
    if err != nil {
        return err
    }
    return ns.s.do(ctx, req, nil)
}

// --- 自定义方法 ---

func (ns *NodesService) Move(ctx context.Context, id string, r *MoveNodeRequest) error {
    req, err := ns.s.newRequest("POST", "/nodes/"+id+"/move", r)
    if err != nil {
        return err
    }
    return ns.s.do(ctx, req, nil)
}

func (ns *NodesService) Complete(ctx context.Context, id string) error {
    req, err := ns.s.newRequest("POST", "/nodes/"+id+"/complete", nil)
    if err != nil {
        return err
    }
    return ns.s.do(ctx, req, nil)
}

func (ns *NodesService) Uncomplete(ctx context.Context, id string) error {
    req, err := ns.s.newRequest("POST", "/nodes/"+id+"/uncomplete", nil)
    if err != nil {
        return err
    }
    return ns.s.do(ctx, req, nil)
}

func (ns *NodesService) Export(ctx context.Context) ([]*Node, error) {
    req, err := ns.s.newRequest("GET", "/nodes-export", nil)
    if err != nil {
        return nil, err
    }
    var resp struct {
        Nodes []*Node `json:"nodes"`
    }
    if err := ns.s.do(ctx, req, &resp); err != nil {
        return nil, err
    }
    return resp.Nodes, nil
}
```

### 5.2 TargetsService

```go
type TargetsService struct {
    s *service
}

func (ts *TargetsService) List(ctx context.Context) ([]*Target, error) {
    req, err := ts.s.newRequest("GET", "/targets", nil)
    if err != nil {
        return nil, err
    }
    var resp struct {
        Targets []*Target `json:"targets"`
    }
    if err := ts.s.do(ctx, req, &resp); err != nil {
        return nil, err
    }
    return resp.Targets, nil
}
```

---

## 六、错误处理设计

AIP-193 对错误处理有明确的指导：错误应该是结构化的，包含 code、message 和可选的 details。WorkFlowy API 的错误格式文档中没有明确说明，但一个好的 SDK 应该提供结构化的错误类型：

```go
// Error 表示 WorkFlowy API 返回的错误。
type Error struct {
    StatusCode int    // HTTP 状态码
    Message    string // 错误消息
    RawBody    string // 原始响应体（用于调试）
}

func (e *Error) Error() string {
    return fmt.Sprintf("workflowy: %d - %s", e.StatusCode, e.Message)
}

// IsNotFound 判断是否为 404 错误。
func IsNotFound(err error) bool {
    var apiErr *Error
    if errors.As(err, &apiErr) {
        return apiErr.StatusCode == 404
    }
    return false
}

// IsRateLimited 判断是否为限流错误（Export 端点 1次/分钟）。
func IsRateLimited(err error) bool {
    var apiErr *Error
    if errors.As(err, &apiErr) {
        return apiErr.StatusCode == 429
    }
    return false
}
```

提供 `IsNotFound`、`IsRateLimited` 这样的判断函数，比让用户自己做类型断言和状态码比较要友好得多。这也是 Google Cloud Go 客户端库的常见模式。

---

## 七、进阶设计考量

### 7.1 树操作辅助方法

WorkFlowy 的 Export 端点返回扁平列表，但用户通常需要树形结构。SDK 可以提供一个辅助方法来重建树：

```go
// NodeTree 是带有子节点引用的树形节点。
type NodeTree struct {
    *Node
    Children []*NodeTree
}

// BuildTree 将扁平的节点列表重建为树形结构。
func BuildTree(nodes []*Node) []*NodeTree {
    // 按 parent_id 分组，按 priority 排序，递归构建
    // ...
}
```

这不是 API 的直接映射，而是 SDK 层面的增值功能。Google 的客户端库通常不做这种事（它们是自动生成的），但手写的 SDK 完全应该提供这样的便利。

### 7.2 重试与限流

Export 端点有明确的 1 次/分钟限流。SDK 应该内置重试逻辑：

```go
func WithRetry(maxRetries int) Option {
    return func(c *Client) { c.maxRetries = maxRetries }
}
```

对于 429 响应，自动等待 `Retry-After` 头指定的时间后重试。对于网络错误和 5xx 错误，使用指数退避策略。

### 7.3 Context 的使用

所有公开方法的第一个参数都是 `context. Context`，这是 Go 的惯例，也是 Google Cloud Go 客户端库的标准做法。它支持超时控制、取消传播和请求级别的元数据传递。

### 7.4 接口抽象（可测试性）

为了让用户能够 mock SDK 进行单元测试，可以提供接口定义：

```go
type NodesAPI interface {
    Get(ctx context.Context, id string) (*Node, error)
    List(ctx context.Context, opts *ListNodesOptions) ([]*Node, error)
    Create(ctx context.Context, r *CreateNodeRequest) (*CreateNodeResponse, error)
    Update(ctx context.Context, id string, r *UpdateNodeRequest) error
    Delete(ctx context.Context, id string) error
    Move(ctx context.Context, id string, r *MoveNodeRequest) error
    Complete(ctx context.Context, id string) error
    Uncomplete(ctx context.Context, id string) error
    Export(ctx context.Context) ([]*Node, error)
}

type TargetsAPI interface {
    List(ctx context.Context) ([]*Target, error)
}
```

不过，Go 社区的惯例是“在消费侧定义接口”（Accept interfaces, return structs），所以这些接口可以放在一个单独的子包中（如 `workflowytest`），或者干脆让用户自己定义他们需要的接口子集。

---

## 八、项目结构

```plaintext
workflowy-go/
├── workflowy.go          // Client, NewClient, Option, service
├── nodes.go              // NodesService, Node, 请求/响应类型
├── targets.go            // TargetsService, Target
├── types.go              // ParentRef, LayoutMode, Position, Timestamp 等公共类型
├── errors.go             // Error, IsNotFound, IsRateLimited
├── tree.go               // BuildTree 等辅助方法
├── workflowy_test.go     // 集成测试
├── examples/
│   ├── basic/main.go
│   └── export-tree/main.go
├── go.mod
└── README.md
```

模块路径建议：`github.com/{你的用户名}/workflowy-go`，包名 `workflowy`。

---

## 九、完整使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/example/workflowy-go"
)

func main() {
    ctx := context.Background()

    // 1. 创建客户端
    client, err := workflowy.NewClient(
        workflowy.WithAPIKey("wf_your_api_key"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // 2. 查看可用的 Targets
    targets, err := client.Targets.List(ctx)
    if err != nil {
        log.Fatal(err)
    }
    for _, t := range targets {
        fmt.Printf("Target: %s (%s)
", t.Key, t.Type)
    }

    // 3. 在 inbox 中创建一个项目大纲
    created, err := client.Nodes.Create(ctx, &workflowy.CreateNodeRequest{
        ParentID: workflowy.TargetInbox,
        Name:     "# API SDK 开发计划",  // 使用 markdown 语法，自动设为 h1
    })
    if err != nil {
        log.Fatal(err)
    }
    projectID := created.ItemID

    // 4. 在项目下创建待办事项
    for _, task := range []string{"设计类型系统", "实现 HTTP 层", "编写测试"} {
        _, err := client.Nodes.Create(ctx, &workflowy.CreateNodeRequest{
            ParentID:   workflowy.NodeParent(projectID),
            Name:       task,
            LayoutMode: workflowy.LayoutTodo,
            Position:   workflowy.PositionBottom,
        })
        if err != nil {
            log.Fatal(err)
        }
    }

    // 5. 完成第一个任务
    children, err := client.Nodes.List(ctx, &workflowy.ListNodesOptions{
        ParentID: workflowy.NodeParent(projectID),
    })
    if err != nil {
        log.Fatal(err)
    }
    if len(children) > 0 {
        err = client.Nodes.Complete(ctx, children[0].ID)
        if err != nil {
            log.Fatal(err)
        }
    }

    // 6. 导出并构建树
    allNodes, err := client.Nodes.Export(ctx)
    if err != nil {
        log.Fatal(err)
    }
    tree := workflowy.BuildTree(allNodes)
    fmt.Printf("顶层节点数: %d
", len(tree))
}
```

---

## 十、设计决策总结

| 决策点 | 选择 | 理由 |
| --- | --- | --- |
| 调用风格 | `client. Nodes. Get(ctx, id)` | 符合 Google Go SDK 惯例，资源导向，清晰的命名空间 |
| 可选参数 | Request struct + 指针字段 | 比 Call 链式更简洁，适合参数少的 API |
| 配置方式 | Functional Options | Go 社区最佳实践，向后兼容 |
| 父级引用 | `ParentRef` 命名类型 | 类型安全，自文档化，IDE 友好 |
| nullable 字段 | 指针类型 | Go 惯用法，正确处理 PATCH 语义 |
| 时间戳 | 自定义 `Timestamp` 类型 | 封装 Unix 时间戳 ↔ `time. Time` 转换 |
| 错误处理 | 结构化 `*Error` + 判断函数 | 符合 AIP-193 精神，用户友好 |
| List 排序 | SDK 内部自动按 priority 排序 | API 返回无序，SDK 应屏蔽这个细节 |
| 树重建 | `BuildTree` 辅助方法 | Export 返回扁平列表，树是用户的真实需求 |
| 可测试性 | 接口定义（可选） | 支持 mock，但不强制 |

这套设计在 Google AIP 资源导向理念和 Go 语言惯用法之间取得了平衡——它不是对 Google 重量级客户端生成器的照搬，而是针对 WorkFlowy 这个精简 API 的量身定制。
