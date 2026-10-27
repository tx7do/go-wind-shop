package scripting

import (
	"context"
	"testing"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"
)

// TestHotReload_CallbackIdempotent 脚本热更新重执行 hook.register 时，
// 同归属（同名脚本）旧回调必须被替换而非追加。
func TestHotReload_CallbackIdempotent(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ScriptDir = ""
	e := NewEngine(cfg, bLogger.NopLogger())
	defer e.Close()

	source := `
local hook = require "kratos_hook"
hook.register("reload.event", "v1 callback", function(ctx)
	ctx.set("version", 1)
	return true
end)
`
	if err := e.LoadScriptString(context.Background(), "reload_script", source); err != nil {
		t.Fatalf("first load failed: %v", err)
	}

	// 模拟热更新：同名脚本再次注册（新函数闭包）
	sourceV2 := `
local hook = require "kratos_hook"
hook.register("reload.event", "v2 callback", function(ctx)
	ctx.set("version", 2)
	return true
end)
`
	if err := e.LoadScriptString(context.Background(), "reload_script", sourceV2); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	e.callbacksMu.RLock()
	cbs := e.callbacks["reload.event"]
	e.callbacksMu.RUnlock()

	if len(cbs) != 1 {
		t.Fatalf("expected 1 callback after reload, got %d (duplicate registration not replaced)", len(cbs))
	}

	// 执行应命中新版本回调
	execCtx := NewContext("reload.event")
	if err := e.ExecuteHook(context.Background(), "reload.event", execCtx); err != nil {
		t.Fatalf("ExecuteHook failed: %v", err)
	}
	if got := execCtx.GetInt("version"); got != 2 {
		t.Errorf("expected version=2 after reload, got %d", got)
	}
}

// TestHotReload_RegistryUpsert 注册表 AddScript 同名脚本应替换而非报错。
func TestHotReload_RegistryUpsert(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ScriptDir = ""
	e := NewEngine(cfg, bLogger.NopLogger())
	defer e.Close()

	if err := e.RegisterHook("upsert.hook", "test"); err != nil {
		t.Fatalf("RegisterHook failed: %v", err)
	}

	s1 := &Script{Name: "s1", Hook: "upsert.hook", Source: "return 1", Enabled: true, Priority: 1}
	if err := e.AddScript("upsert.hook", s1); err != nil {
		t.Fatalf("AddScript v1 failed: %v", err)
	}
	s2 := &Script{Name: "s1", Hook: "upsert.hook", Source: "return 2", Enabled: true, Priority: 2}
	if err := e.AddScript("upsert.hook", s2); err != nil {
		t.Fatalf("AddScript v2 (same name) should upsert, got error: %v", err)
	}

	scripts := e.registry.GetScripts("upsert.hook")
	if len(scripts) != 1 {
		t.Fatalf("expected 1 script after upsert, got %d", len(scripts))
	}
	if scripts[0].Priority != 2 {
		t.Errorf("expected replaced script (priority 2), got priority %d", scripts[0].Priority)
	}
}

// TestExecute_MissingErrorNotSwallowed 脚本 execute() 运行时报错必须透出，不得被吞。
func TestExecute_MissingErrorNotSwallowed(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ScriptDir = ""
	e := NewEngine(cfg, bLogger.NopLogger())
	defer e.Close()

	script := &Script{
		Name: "boom_script",
		Source: `
function execute()
	local t = nil
	return t.field -- attempt to index a nil value
end
`,
		Enabled: true,
	}
	execCtx := NewContext("boom")
	err := e.Execute(context.Background(), script, execCtx)
	if err == nil {
		t.Fatal("expected runtime error from execute(), got nil (error swallowed)")
	}
}
