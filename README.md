# ddd-go

Go 语言版 DDD（领域驱动设计）框架，提供 CQRS 风格的命令总线、事件总线及配套领域原语。

* [架构风格](#架构风格)
* [使用方式](#使用方式)
* [重要组件介绍](#重要组件介绍)
   * [Command（命令）](#command命令)
   * [Event（事件）](#event事件)
   * [CommandBus（命令总线）](#commandbus命令总线)
   * [EventBus（事件总线）](#eventbus事件总线)
   * [CommandHandler（命令处理器）](#commandhandler命令处理器)
   * [CommandInterceptor（命令拦截器）](#commandinterceptor命令拦截器)
   * [CommandValidator（命令验证器）](#commandvalidator命令验证器)
   * [Assembler（组装器）](#assembler组装器)
   * [Converter（转换器）](#converter转换器)
   * [ConcurrencyConflicts（并发校验工具）](#concurrencyconflicts并发校验工具)
* [参考](#参考)

---

### 架构风格

结合了[整洁架构风格](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)、CQRS 风格以及分层架构风格，采用依赖原则分 4 层：

* **adapter** 接口适配层（适配 HTTP、gRPC、Task、Consumer 等接口）
* **application** 应用层（实现用例，如用户下单、商家发布商品等）
* **domain** 领域层（编写领域逻辑，如订单逻辑、商品逻辑、优惠逻辑等）
* **infrastructure** 基础层（Cache、MQ、数据库持久化等实现）

---

### 使用方式

```bash
go get github.com/runssnail/ddd-go
```

**初始化总线（以应用启动为例）：**

```go
import (
    "github.com/runssnail/ddd-go/common"
    "github.com/runssnail/ddd-go/core"
)

eventBus := core.NewEventBus()
eventBus.RegisterHandler(myEventHandler)

bus := core.NewCommandBus()
bus.RegisterHandler(myCommandHandler)
bus.RegisterValidator(myValidator)
bus.RegisterInterceptor(myInterceptor)
bus.Init() // 必须在所有注册完成后调用
```

---

### 重要组件介绍

#### Command（命令）

一个 Command 对象对应一个用例的请求数据。在结构体中嵌入 `common.BaseCommand[R]`，泛型参数 `R` 为该命令的返回结果类型。

```go
type CreateProductCommand struct {
    common.BaseCommand[common.Result[string]]
    Name        string
    Description string
}
```

只读查询场景可嵌入 `common.Query[R]`；需要分页时嵌入 `common.PagingQuery[R]`：

```go
type QueryProductCommand struct {
    common.PagingQuery[common.PagingResult[ProductDTO]]
    Name string
}
```

---

#### Event（事件）

表示一个领域事件，用例完成后发布。嵌入 `common.BaseEvent` 并实现 `Tag()` 以标识事件类型。

```go
type ProductCreatedEvent struct {
    common.BaseEvent
    ProductID string
}

func (e *ProductCreatedEvent) Tag() string { return "product.created" }

// 构造时初始化时间戳
event := &ProductCreatedEvent{
    BaseEvent: common.NewBaseEvent(),
    ProductID: "id-1",
}
```

---

#### CommandBus（命令总线）

将 Command 分发到对应的 CommandHandler 处理业务。使用类型安全的 `core.Dispatch[R]` 函数发送命令并获取类型化结果。

```go
result, err := core.Dispatch[common.Result[string]](ctx, bus, &CreateProductCommand{
    Name: "widget",
})
if err != nil {
    // 处理错误
}
```

---

#### EventBus（事件总线）

用来发布领域事件。支持同步发布（`Publish`）和异步发布（`AsyncPublish`，每个 handler 在独立 goroutine 中执行）。

```go
// 同步发布
eventBus.Publish(ctx, &ProductCreatedEvent{
    BaseEvent: common.NewBaseEvent(),
    ProductID: product.ID,
})

// 异步发布
eventBus.AsyncPublish(ctx, &ProductCreatedEvent{
    BaseEvent: common.NewBaseEvent(),
    ProductID: product.ID,
})
```

---

#### CommandHandler（命令处理器）

实现用例，一个 Command 对应一个 CommandHandler。有两种风格：

**结构体方式**（适合有依赖注入的处理器）：

```go
type CreateProductHandler struct {
    repo     ProductRepository
    eventBus core.EventBus
}

func (h *CreateProductHandler) SupportCommand() reflect.Type {
    return reflect.TypeOf((*CreateProductCommand)(nil))
}

func (h *CreateProductHandler) Handle(ctx context.Context, cmd any) (any, error) {
    c := cmd.(*CreateProductCommand)

    product := h.repo.Save(ctx, c.Name, c.Description)

    h.eventBus.Publish(ctx, &ProductCreatedEvent{
        BaseEvent: common.NewBaseEvent(),
        ProductID: product.ID,
    })

    return common.Success(product.ID), nil
}
```

**函数方式**（轻量无依赖的处理器）：

```go
bus.RegisterHandler(core.NewHandlerFunc(
    func(ctx context.Context, cmd *CreateProductCommand) (common.Result[string], error) {
        return common.Success("id-" + cmd.Name), nil
    },
))
```

---

#### CommandInterceptor（命令拦截器）

拦截 Command 的执行，支持多个拦截器作用于同一 Command。嵌入 `core.CommandInterceptorBase` 并只覆盖需要的钩子方法。

```go
type LoggingInterceptor struct {
    core.CommandInterceptorBase
}

func NewLoggingInterceptor() *LoggingInterceptor {
    return &LoggingInterceptor{
        // 作用于所有命令，执行顺序为 1（值越小越先执行）
        CommandInterceptorBase: core.NewGlobalInterceptorBase(1),
    }
}

func (i *LoggingInterceptor) BeforeHandle(ctx context.Context, cmd any) error {
    slog.Info("before handle", "cmd", cmd)
    return nil
}

func (i *LoggingInterceptor) AfterHandle(ctx context.Context, cmd any, result any) error {
    slog.Info("after handle", "cmd", cmd, "result", result)
    return nil
}
```

如需将拦截器限定到某个具体命令类型，使用 `core.NewInterceptorBase[*YourCommand](order)`。

---

#### CommandValidator（命令验证器）

验证 Command 的参数完整性或业务前置条件。实现 `common.CommandValidator` 接口并通过 `bus.RegisterValidator` 注册。

```go
type CreateProductValidator struct{}

func (v *CreateProductValidator) SupportType() reflect.Type {
    return reflect.TypeOf((*CreateProductCommand)(nil))
}

func (v *CreateProductValidator) Validate(_ context.Context, cmd any) error {
    c := cmd.(*CreateProductCommand)
    if c.Name == "" {
        return common.NewBizError(common.ErrParams, "name is required")
    }
    if len(c.Name) > 10 {
        return common.NewBizErrorf(common.ErrParams, "name too long: %d chars", len(c.Name))
    }
    return nil
}
```

验证器在 `bus.Init()` 时被自动织入为第一个拦截器，无需手动排序。

---

#### Assembler（组装器）

将领域实体对象转换成 DTO 后返回给外部，隔离领域内部实现。

```go
type ProductAssembler struct{}

func (a *ProductAssembler) Assemble(product *domain.Product) *ProductDTO {
    return &ProductDTO{
        ProductID:   product.UUID,
        Name:        product.Name,
        Description: product.Description,
    }
}
```

`Assembler[Source, Target]` 接口定义在 `core` 包中：

```go
type Assembler[Source, Target any] interface {
    Assemble(source Source) Target
}
```

---

#### Converter（转换器）

实现领域实体对象与数据持久化对象之间的双向转换。

```go
type ProductConverter struct{}

func (c *ProductConverter) Serialize(product *domain.Product) *ProductDO {
    return &ProductDO{
        ID:          product.ID,
        UUID:        product.UUID,
        Name:        product.Name,
        Description: product.Description,
    }
}

func (c *ProductConverter) Deserialize(do *ProductDO) *domain.Product {
    p := &domain.Product{}
    p.ID = do.ID
    p.UUID = do.UUID
    p.Name = do.Name
    p.Description = do.Description
    return p
}
```

`Converter[Domain, Data]` 接口定义在 `core` 包中：

```go
type Converter[Domain, Data any] interface {
    Serialize(domain Domain) Data
    Deserialize(data Data) Domain
}
```

---

#### ConcurrencyConflicts（并发校验工具）

在执行 UPDATE/DELETE 后检测乐观锁冲突。`CheckRowsAffected` 在影响行数不等于 1 时返回 `*ConcurrencyConflictError`。

```go
func (r *ProductRepository) Remove(ctx context.Context, product *domain.Product) error {
    count, err := r.db.ExecContext(ctx,
        "DELETE FROM product WHERE id=? AND version=?",
        product.ID, product.ConcurrencyVersion,
    )
    if err != nil {
        return err
    }
    return common.CheckRowsAffected(int(count), "remove product: stale version")
}
```

通过 `common.IsConcurrencyConflict(err)` 判断是否为并发冲突错误：

```go
if common.IsConcurrencyConflict(err) {
    // 提示用户数据已被修改，请刷新后重试
}
```

---

### 参考

* [整洁架构](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
* [Alibaba COLA](https://github.com/alibaba/COLA)
* [AxonFramework](https://github.com/AxonFramework/AxonFramework)
