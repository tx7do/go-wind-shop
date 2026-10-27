package scripting

// 本文件针对 runtime_lua.go / runtime_javascript.go 的适配器做单元测试：
//   - luaCallback / jsCallback 的 Source 标识串；
//   - LuaBinder / JSBinder 的 httpOpts 两分支（无 cfg fail-closed 空护栏、
//     有 cfg 透传编排器配置）；
//   - luaVMManagerAdapter.MarkVMDedicated（no-op）；
//   - contextToLuaTable：直连 LState 的表行为（get 命中/缺失、set、stop 带
//     与不带 reason 两分支）；
//   - registerContextFunctions 的空上下文分支（LoadScriptString 期间
//     holder 为空：__get_ctx 返回空表、__set_ctx/__stop 为 no-op，Lua/JS 双引擎）；
//   - 经引擎执行的非空上下文分支：ctx.get/ctx.set/ctx.stop 的往返与中止
//     （Lua 的 contextToLuaTable 与 JS 的 jsContextObject，含带/不带 reason）；
//   - SupportedTypes：已注册适配器包含 Lua 与 JavaScript 两类型。

import (
	"context"
	"runtime"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"

	"github.com/stretchr/testify/require"
	gsEngine "github.com/tx7do/go-scripts"

	"github.com/tx7do/go-scripts/hostmodule"
)

// settleJSWatcherRace 等待本测试遗留的 go-scripts/js 中断监视 goroutine
// （随每次带超时 ctx 的 ExecuteString/CallFunction 派生）调度完毕，并就地
// 回收本测试创建的引擎垃圾：避免遗留 goroutine / GC 债务堆积推迟后续测试
// 中同型监视 goroutine 的调度。该竞态的根因（旧版监视器在双就绪 select 下
// 随机选边 + 立即 cancel 制造双就绪窗口）已在上游根修并随
// go-scripts/javascript v0.0.9 升版本仓后彻底退役，本函数仅保留作测试间
// 调度/GC 噪声抑制。生产行为零改动。
func settleJSWatcherRace() {
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
}

// TestCallbackSourceStrings 回调 Source 标识串的固定格式。
func TestCallbackSourceStrings(t *testing.T) {
	require.Equal(t, "lua:hook=probe-lua", (&luaCallback{hookName: "probe-lua"}).Source())
	require.Equal(t, "js:hook=probe-js", (&jsCallback{hookName: "probe-js"}).Source())
}

// TestBinderHTTPOpts httpOpts 两分支：无 cfg 时 fail-closed 空护栏，
// 有 cfg 时透传编排器配置（Lua/JS 两侧一致）。
func TestBinderHTTPOpts(t *testing.T) {
	empty := hostmodule.HTTPOptions{}
	require.Equal(t, empty, (&LuaBinder{}).httpOpts())
	require.Equal(t, empty, (&JSBinder{}).httpOpts())

	cfg := &Config{
		HTTPOptions: hostmodule.HTTPOptions{
			AllowedDomains:  []string{"a.example.com", "b.example.com"},
			Timeout:         7,
			MaxResponseBody: 4096,
			DenyLoopback:    true,
		},
	}
	require.Equal(t, cfg.HTTPOptions, (&LuaBinder{cfg: cfg}).httpOpts())
	require.Equal(t, cfg.HTTPOptions, (&JSBinder{cfg: cfg}).httpOpts())
}

// TestMarkVMDedicatedNoop go-scripts 池自管 VM 生命周期：适配器为 no-op，
// 调用即覆盖。
func TestMarkVMDedicatedNoop(t *testing.T) {
	(luaVMManagerAdapter{}).MarkVMDedicated(nil)
}

// TestContextToLuaTableDirect 直连 LState 验证 contextToLuaTable 的表行为：
// get 缺失键返回 nil、命中键返回原值，set 写入 Go 侧上下文，
// stop 带 reason 与不带 reason（默认）分别落 StopReason。
func TestContextToLuaTableDirect(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	ctx := NewContext("ctxtable_direct")
	ctx.Set("present", "value")
	L.SetGlobal("ctx", contextToLuaTable(L, ctx))

	require.NoError(t, L.DoString(`
assert(ctx.get("missing") == nil, "missing key must be nil")
assert(ctx.get("present") == "value", "present key must round-trip")
ctx.set("probe_key", "probe_value")
assert(ctx.get("probe_key") == "probe_value", "set value must round-trip")
`))
	require.Equal(t, "probe_value", ctx.GetString("probe_key"))
	require.Empty(t, ctx.GetString("present_missing"))

	// stop 带 reason
	require.NoError(t, L.DoString(`ctx.stop("direct-reason")`))
	require.True(t, ctx.Stopped)
	require.Equal(t, "direct-reason", ctx.StopReason)

	// stop 不带 reason → 默认文案
	ctx2 := NewContext("ctxtable_direct2")
	L.SetGlobal("ctx2", contextToLuaTable(L, ctx2))
	require.NoError(t, L.DoString(`ctx2.stop()`))
	require.True(t, ctx2.Stopped)
	require.Equal(t, "stopped by script", ctx2.StopReason)
}

// TestRegisterContextFunctionsNilCtxBranches LoadScriptString 期间 holder 为空：
// __get_ctx 返回空表（无 get/set/stop 方法）、__set_ctx/__stop 为 no-op。
func TestRegisterContextFunctionsNilCtxBranches(t *testing.T) {
	t.Run("lua empty ctx object", func(t *testing.T) {
		e := newTestEngine(t)
		require.NoError(t, e.LoadScriptString(context.Background(), "nilctx_lua", `
local t = __get_ctx()
assert(type(t) == "table", "expected a table")
assert(t.get == nil, "expected empty table outside execution")
assert(t.set == nil, "expected empty table outside execution")
assert(t.stop == nil, "expected empty table outside execution")
__set_ctx("k", "v")
__stop("r")
`))
	})

	t.Run("js empty ctx object", func(t *testing.T) {
		e := newTestJSEngine(t)
		require.NoError(t, e.LoadScriptString(context.Background(), "nilctx_js", `
var t = __get_ctx()
if (Object.keys(t).length !== 0) { throw "expected empty ctx object outside execution" }
__set_ctx("k", 1)
__stop("r")
`))
		settleJSWatcherRace()
	})
}

// TestContextTableViaLuaHookExecution 经引擎执行（非空上下文）验证
// contextToLuaTable 的 get 命中/缺失与 set 落库。
func TestContextTableViaLuaHookExecution(t *testing.T) {
	e := newTestEngine(t)
	require.NoError(t, e.AddScript("ctxtable_hook", &Script{
		Name: "ctxtable_script", Hook: "ctxtable_hook", Enabled: true, Priority: 1,
		Source: `
function execute()
    local c = __get_ctx()
    local v = c.get("missing")
    assert(v == nil, "missing key must be nil")
    c.set("probe_key", "probe_value")
    assert(c.get("probe_key") == "probe_value", "set value must round-trip")
    return true
end
`,
	}))
	execCtx := NewContext("ctxtable_hook")
	require.NoError(t, e.ExecuteHook(context.Background(), "ctxtable_hook", execCtx))
	require.Equal(t, "probe_value", execCtx.GetString("probe_key"))
}

// TestContextStopViaLuaHookExecution 经引擎执行验证 contextToLuaTable 的
// stop 两分支（默认文案与自定义 reason），均置 Stopped 并使执行中止。
func TestContextStopViaLuaHookExecution(t *testing.T) {
	t.Run("default reason", func(t *testing.T) {
		e := newTestEngine(t)
		require.NoError(t, e.AddScript("ctxstop_default", &Script{
			Name: "ctxstop_default_script", Hook: "ctxstop_default", Enabled: true, Priority: 1,
			Source: `
function execute()
    local c = __get_ctx()
    c.stop()
    return true
end
`,
		}))
		execCtx := NewContext("ctxstop_default")
		err := e.ExecuteHook(context.Background(), "ctxstop_default", execCtx)
		require.ErrorContains(t, err, "execution stopped: stopped by script")
		require.True(t, execCtx.Stopped)
		require.Equal(t, "stopped by script", execCtx.StopReason)
	})

	t.Run("custom reason", func(t *testing.T) {
		e := newTestEngine(t)
		require.NoError(t, e.AddScript("ctxstop_custom", &Script{
			Name: "ctxstop_custom_script", Hook: "ctxstop_custom", Enabled: true, Priority: 1,
			Source: `
function execute()
    local c = __get_ctx()
    c.stop("lua custom reason")
    return true
end
`,
		}))
		execCtx := NewContext("ctxstop_custom")
		err := e.ExecuteHook(context.Background(), "ctxstop_custom", execCtx)
		require.ErrorContains(t, err, "execution stopped: lua custom reason")
		require.True(t, execCtx.Stopped)
		require.Equal(t, "lua custom reason", execCtx.StopReason)
	})
}

// TestJSContextObjectViaHookExecution JS 侧 jsContextObject 的
// get 命中/缺失与 set 落库（经引擎执行）。
func TestJSContextObjectViaHookExecution(t *testing.T) {
	e := newTestJSEngine(t)
	require.NoError(t, e.AddScript("jsctx_hook", &Script{
		Name: "jsctx_script", Hook: "jsctx_hook", Enabled: true, Priority: 1,
		Source: `
function execute() {
    var c = __get_ctx()
    if (c.get("missing") != null) { throw "missing key must be null" }
    c.set("probe_key", "probe_value")
    if (c.get("probe_key") !== "probe_value") { throw "set value must round-trip" }
    return true
}
`,
	}))
	execCtx := NewContext("jsctx_hook")
	require.NoError(t, e.ExecuteHook(context.Background(), "jsctx_hook", execCtx))
	require.Equal(t, "probe_value", execCtx.GetString("probe_key"))
	settleJSWatcherRace()
}

// TestJSContextStopViaHookExecution JS 侧 jsContextObject 的 stop 两分支
// （默认文案与自定义 reason），均置 Stopped 并使执行中止。
func TestJSContextStopViaHookExecution(t *testing.T) {
	t.Run("default reason", func(t *testing.T) {
		e := newTestJSEngine(t)
		require.NoError(t, e.AddScript("jsctxstop_default", &Script{
			Name: "jsctxstop_default_script", Hook: "jsctxstop_default", Enabled: true, Priority: 1,
			Source: `
function execute() {
    var c = __get_ctx()
    c.stop()
    return true
}
`,
		}))
		execCtx := NewContext("jsctxstop_default")
		err := e.ExecuteHook(context.Background(), "jsctxstop_default", execCtx)
		require.ErrorContains(t, err, "execution stopped: stopped by script")
		require.True(t, execCtx.Stopped)
		require.Equal(t, "stopped by script", execCtx.StopReason)
		settleJSWatcherRace()
	})

	t.Run("custom reason", func(t *testing.T) {
		e := newTestJSEngine(t)
		require.NoError(t, e.AddScript("jsctxstop_custom", &Script{
			Name: "jsctxstop_custom_script", Hook: "jsctxstop_custom", Enabled: true, Priority: 1,
			Source: `
function execute() {
    var c = __get_ctx()
    c.stop("js custom reason")
    return true
}
`,
		}))
		execCtx := NewContext("jsctxstop_custom")
		err := e.ExecuteHook(context.Background(), "jsctxstop_custom", execCtx)
		require.ErrorContains(t, err, "execution stopped: js custom reason")
		require.True(t, execCtx.Stopped)
		require.Equal(t, "js custom reason", execCtx.StopReason)
		settleJSWatcherRace()
	})
}

// TestSupportedTypes 已注册语言适配器覆盖 Lua 与 JavaScript 两类型。
func TestSupportedTypes(t *testing.T) {
	types := SupportedTypes()
	require.Contains(t, types, gsEngine.LuaType)
	require.Contains(t, types, gsEngine.JavaScriptType)
}
