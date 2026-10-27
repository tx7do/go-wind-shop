package api

import (
	"context"
	"fmt"
	"reflect"
	"sync/atomic"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"

	gsEngine "github.com/tx7do/go-scripts"
	"github.com/tx7do/go-scripts/hostmodule"

	"github.com/tx7do/go-utils/eventbus"
)

// eventbusCbCounter 用于生成唯一的事件回调全局名。
var eventbusCbCounter int64

// ModuleEventBus 构建语言无关的 eventbus 模块（JS 等基于 map[string]any 桥接的语言使用）。
//
// handler 参数是脚本注册时捕获的原始函数值（goja 为 goja.Callable）：
// 与 jsCallback 同款策略——订阅阶段（脚本执行中）不调用引擎 API（goja 重入死锁），
// 事件触发时先 RegisterGlobal 为唯一命名全局，再经引擎 CallFunction 调用。
// eng 为 nil 或 manager 为 nil 时返回不含函数的空模块。
func ModuleEventBus(eng gsEngine.Engine, manager *eventbus.Manager, logger *bLogger.Helper) hostmodule.ModuleDef {
	if eng == nil || manager == nil {
		return hostmodule.ModuleDef{Name: "eventbus", Funcs: map[string]any{}}
	}

	bg := context.Background()
	busOf := func(busName string) eventbus.EventBus {
		if busName == "" || busName == "global" {
			return manager.Global()
		}
		return manager.GetBus(busName)
	}

	// captureHandler 把脚本函数值包装为 eventbus.Handler：
	// 触发时把函数值注册为唯一命名全局，再以事件 map 为唯一参数调用。
	captureHandler := func(handler any) eventbus.Handler {
		name := fmt.Sprintf("__eventbus_cb_%d", atomic.AddInt64(&eventbusCbCounter, 1))
		return &jsEventBusHandler{
			engine:    eng,
			funcName:  name,
			funcValue: handler,
			logger:    logger,
		}
	}

	return hostmodule.ModuleDef{
		Name: "eventbus",
		Funcs: map[string]any{
			// publish(eventType, data) → 向 global 总线发布
			"publish": func(eventType string, data any) error {
				return manager.Publish(bg, "global", eventbus.NewEvent(eventType, data))
			},
			// publishTo(busName, eventType, data) → 向指定总线发布
			"publishTo": func(busName, eventType string, data any) error {
				return manager.Publish(bg, busName, eventbus.NewEvent(eventType, data))
			},
			// subscribe(eventType, handler) → global 总线同步订阅
			"subscribe": func(eventType string, handler any) error {
				return manager.Global().Subscribe(eventType, captureHandler(handler))
			},
			// subscribeAsync(eventType, handler) → global 总线异步订阅
			"subscribeAsync": func(eventType string, handler any) error {
				return manager.Global().SubscribeAsync(eventType, captureHandler(handler))
			},
			// subscribeOnce(eventType, handler) → global 总线一次性订阅
			"subscribeOnce": func(eventType string, handler any) error {
				return manager.Global().SubscribeOnce(eventType, captureHandler(handler))
			},
			// subscribeOnBus(busName, eventType, handler) → 指定总线同步订阅
			"subscribeOnBus": func(busName, eventType string, handler any) error {
				return busOf(busName).Subscribe(eventType, captureHandler(handler))
			},
		},
	}
}

// jsEventBusHandler 适配脚本函数值为 eventbus.Handler。
type jsEventBusHandler struct {
	engine    gsEngine.Engine
	funcName  string // 回调函数的全局名（__eventbus_cb_N）
	funcValue any    // 原始脚本函数值（订阅时捕获，触发时注册为全局）
	logger    *bLogger.Helper
}

var _ eventbus.Handler = (*jsEventBusHandler)(nil)

// Handle 触发脚本函数，传入事件对象（map 形式）。
//
// 调用语义与 Lua 版 LuaEventHandler 对齐：优先在当前 goroutine 直接调用脚本函数值
// （发布发生在脚本执行内时为同 goroutine 合法重入，不经引擎锁，避免 execMu 自死锁）；
// 函数值不可直接调用时（异步/跨 goroutine 场景）退回引擎 CallFunction（带锁串行）。
func (h *jsEventBusHandler) Handle(ctx context.Context, event *eventbus.Event) error {
	eventArg := map[string]any{
		"id":        event.ID,
		"type":      event.Type,
		"source":    event.Source,
		"priority":  event.Priority,
		"timestamp": event.Timestamp.Unix(),
		"data":      event.Data,
		"metadata":  event.Metadata,
	}

	if fn := reflect.ValueOf(h.funcValue); fn.IsValid() && fn.Kind() == reflect.Func {
		out := fn.Call([]reflect.Value{reflect.ValueOf(eventArg)})
		if len(out) >= 1 {
			if err, ok := out[len(out)-1].Interface().(error); ok && err != nil {
				if h.logger != nil {
					h.logger.Errorf(ctx, "eventbus handler error: %v", err)
				}
				return fmt.Errorf("eventbus handler error: %w", err)
			}
		}
		return nil
	}

	if h.engine == nil {
		return nil
	}

	// 调用阶段无正在进行的 ExecuteString 时可安全获取引擎锁
	if err := h.engine.RegisterGlobal(h.funcName, h.funcValue); err != nil {
		return fmt.Errorf("eventbus handler register: %w", err)
	}

	if _, err := h.engine.CallFunction(ctx, h.funcName, eventArg); err != nil {
		if h.logger != nil {
			h.logger.Errorf(ctx, "eventbus handler error: %v", err)
		}
		return fmt.Errorf("eventbus handler error: %w", err)
	}
	return nil
}
