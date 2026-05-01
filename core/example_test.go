package core_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/runssnail/ddd-go/common"
	"github.com/runssnail/ddd-go/core"
)

// ── Domain types ──────────────────────────────────────────────────────────────

type CreateProductCommand struct {
	common.BaseCommand[common.Result[string]]
	Name string
}

type ProductCreatedEvent struct {
	common.BaseEvent
	ProductID string
}

func (e *ProductCreatedEvent) Tag() string { return "product.created" }

// ── Validator ─────────────────────────────────────────────────────────────────

type createProductValidator struct{}

func (v *createProductValidator) SupportType() reflect.Type {
	return reflect.TypeOf((*CreateProductCommand)(nil))
}

func (v *createProductValidator) Validate(_ context.Context, cmd any) error {
	c := cmd.(*CreateProductCommand)
	if c.Name == "" {
		return common.NewBizError(common.ErrParams, "name is required")
	}
	return nil
}

// ── Handler ───────────────────────────────────────────────────────────────────

type createProductHandler struct {
	bus core.EventBus
}

func (h *createProductHandler) SupportCommand() reflect.Type {
	return reflect.TypeOf((*CreateProductCommand)(nil))
}

func (h *createProductHandler) Handle(ctx context.Context, cmd any) (any, error) {
	c := cmd.(*CreateProductCommand)
	h.bus.Publish(ctx, &ProductCreatedEvent{
		BaseEvent: common.NewBaseEvent(),
		ProductID: "id-" + c.Name,
	})
	return common.Success("id-" + c.Name), nil
}

// ── Event handler ─────────────────────────────────────────────────────────────

type productCreatedEventHandler struct {
	core.BaseEventHandler[*ProductCreatedEvent]
	receivedID string
}

func (h *productCreatedEventHandler) Handle(_ context.Context, event common.Event) error {
	h.receivedID = event.(*ProductCreatedEvent).ProductID
	return nil
}

// ── Interceptor ───────────────────────────────────────────────────────────────

type loggingInterceptor struct {
	core.CommandInterceptorBase
	beforeCalled bool
	afterCalled  bool
}

func (i *loggingInterceptor) BeforeHandle(_ context.Context, _ any) error {
	i.beforeCalled = true
	return nil
}

func (i *loggingInterceptor) AfterHandle(_ context.Context, _ any, _ any) error {
	i.afterCalled = true
	return nil
}

// ── funcEventHandler — event handler backed by a function ────────────────────

type funcEventHandler struct {
	core.BaseEventHandler[*ProductCreatedEvent]
	fn func(ctx context.Context, event common.Event) error
}

func (f *funcEventHandler) Handle(ctx context.Context, event common.Event) error {
	return f.fn(ctx, event)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCommandBusDispatch(t *testing.T) {
	eventBus := core.NewEventBus()

	evHandler := &productCreatedEventHandler{}
	evHandler.BaseEventHandler = core.NewBaseEventHandler[*ProductCreatedEvent]()
	eventBus.RegisterHandler(evHandler)

	interceptor := &loggingInterceptor{
		CommandInterceptorBase: core.NewGlobalInterceptorBase(1),
	}

	bus := core.NewCommandBus()
	bus.RegisterHandler(&createProductHandler{bus: eventBus})
	bus.RegisterInterceptor(interceptor)
	bus.RegisterValidator(&createProductValidator{})
	bus.Init()

	ctx := context.Background()
	result, err := core.Dispatch[common.Result[string]](ctx, bus, &CreateProductCommand{Name: "widget"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsSuccess() {
		t.Fatalf("expected success, got code=%d", result.Code)
	}
	if result.Data != "id-widget" {
		t.Fatalf("expected id-widget, got %q", result.Data)
	}
	if !interceptor.beforeCalled {
		t.Error("BeforeHandle was not called")
	}
	if !interceptor.afterCalled {
		t.Error("AfterHandle was not called")
	}
	if evHandler.receivedID != "id-widget" {
		t.Errorf("event handler got %q, want id-widget", evHandler.receivedID)
	}
}

func TestValidatorRejectsEmptyName(t *testing.T) {
	eventBus := core.NewEventBus()

	bus := core.NewCommandBus()
	bus.RegisterHandler(&createProductHandler{bus: eventBus})
	bus.RegisterValidator(&createProductValidator{})
	bus.Init()

	_, err := core.Dispatch[common.Result[string]](context.Background(), bus, &CreateProductCommand{Name: ""})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if _, ok := common.IsBizError(err); !ok {
		t.Fatalf("expected *BizError, got %T: %v", err, err)
	}
}

func TestHandlerFunc(t *testing.T) {
	bus := core.NewCommandBus()
	bus.RegisterHandler(core.NewHandlerFunc(func(_ context.Context, cmd *CreateProductCommand) (common.Result[string], error) {
		return common.Success("fn-" + cmd.Name), nil
	}))
	bus.Init()

	result, err := core.Dispatch[common.Result[string]](context.Background(), bus, &CreateProductCommand{Name: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Data != "fn-test" {
		t.Errorf("got %q, want fn-test", result.Data)
	}
}

func TestAsyncPublish(t *testing.T) {
	eventBus := core.NewEventBus()
	done := make(chan string, 1)

	fh := &funcEventHandler{
		fn: func(_ context.Context, event common.Event) error {
			done <- event.(*ProductCreatedEvent).ProductID
			return nil
		},
	}
	fh.BaseEventHandler = core.NewBaseEventHandler[*ProductCreatedEvent]()
	eventBus.RegisterHandler(fh)

	eventBus.AsyncPublish(context.Background(), &ProductCreatedEvent{
		BaseEvent: common.NewBaseEvent(),
		ProductID: "async-id",
	})

	select {
	case id := <-done:
		if id != "async-id" {
			t.Errorf("got %q, want async-id", id)
		}
	case <-time.After(time.Second):
		t.Fatal("async handler not called within 1s")
	}
}

func TestConcurrencyHelpers(t *testing.T) {
	if err := common.CheckRowsAffected(1, "ok"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	err := common.CheckRowsAffected(0, "stale data")
	if err == nil {
		t.Fatal("expected error for 0 rows")
	}
	if !common.IsConcurrencyConflict(err) {
		t.Errorf("expected ConcurrencyConflictError, got %T", err)
	}
}

func TestPagingUtils(t *testing.T) {
	if got := common.GetOffset(1, 20); got != 0 {
		t.Errorf("offset(1,20) = %d, want 0", got)
	}
	if got := common.GetOffset(2, 20); got != 20 {
		t.Errorf("offset(2,20) = %d, want 20", got)
	}
	if got := common.GetPages(100, 20); got != 5 {
		t.Errorf("pages(100,20) = %d, want 5", got)
	}
	if got := common.GetPages(101, 20); got != 6 {
		t.Errorf("pages(101,20) = %d, want 6", got)
	}
}
