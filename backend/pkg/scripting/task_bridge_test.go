package scripting

import (
	"context"
	"testing"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"

	"go-wind-shop/pkg/scripting/api"
)

// TestTaskBridge_RegisterAndInvoke 闭环验证：脚本经 task.register_handler 注册处理器，
// InvokeHandler（Engine.ExecuteTaskHandler 的底层）合并参数默认值、校验必填并执行。
func TestTaskBridge_RegisterAndInvoke(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ScriptDir = ""
	e := NewEngine(cfg, bLogger.NopLogger())
	defer e.Close()

	// 清理型脚本：params 可带 older_than（默认 60），把处理结果写回任务注册表供断言
	source := `
local task = require "task"
task.register_handler("cleanup_demo", "demo cleanup handler", function(params)
	local older = params.older_than or 0
	__set_ctx("ran", true)
	__set_ctx("older_than", older)
	return true
end, {
	optional = { older_than = 60 },
	required = {}
})
`
	if err := e.LoadScriptString(context.Background(), "task_script", source); err != nil {
		t.Fatalf("register script failed: %v", err)
	}

	if _, ok := api.GetHandler("cleanup_demo"); !ok {
		t.Fatal("handler cleanup_demo not registered")
	}

	// 经上下文参数回收执行结果
	execCtx := NewContext("task_invoke")
	prev := e.execCtx.set(execCtx)

	params := map[string]any{"older_than": int64(120)}
	if err := api.InvokeHandler(context.Background(), "cleanup_demo", params); err != nil {
		t.Fatalf("InvokeHandler failed: %v", err)
	}
	e.execCtx.reset(prev)

	if execCtx.GetBool("ran") != true {
		t.Fatal("handler did not run (ctx.ran not set)")
	}
	if got := execCtx.GetInt("older_than"); got != 120 {
		t.Errorf("expected older_than=120 (params override default 60), got %d", got)
	}
}

// TestTaskBridge_MissingRequired InvokeHandler 必须校验必填参数。
func TestTaskBridge_MissingRequired(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ScriptDir = ""
	e := NewEngine(cfg, bLogger.NopLogger())
	defer e.Close()

	source := `
local task = require "task"
task.register_handler("strict_handler", "requires bucket", function(params)
	return true
end, {
	required = { "bucket" }
})
`
	if err := e.LoadScriptString(context.Background(), "strict_script", source); err != nil {
		t.Fatalf("register script failed: %v", err)
	}

	err := api.InvokeHandler(context.Background(), "strict_handler", map[string]any{})
	if err == nil {
		t.Fatal("expected missing-required-param error, got nil")
	}
}

// TestTaskBridge_GenerationPrune 代际清理：旧代际处理器在 Prune 后消失，新代际保留。
func TestTaskBridge_GenerationPrune(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ScriptDir = ""
	e := NewEngine(cfg, bLogger.NopLogger())
	defer e.Close()

	register := func(name string) {
		source := `
local task = require "task"
task.register_handler("` + name + `", "gen test", function(params) return true end)
`
		if err := e.LoadScriptString(context.Background(), name+"_script", source); err != nil {
			t.Fatalf("register %s failed: %v", name, err)
		}
	}

	register("gen_old")
	gen := api.NextTaskGeneration()
	register("gen_new")

	removed := api.PruneStaleTaskHandlers(gen)
	if removed < 1 {
		t.Fatalf("expected at least 1 pruned handler, got %d", removed)
	}
	if _, ok := api.GetHandler("gen_old"); ok {
		t.Error("stale handler gen_old should have been pruned")
	}
	if _, ok := api.GetHandler("gen_new"); !ok {
		t.Error("fresh handler gen_new should survive pruning")
	}
}
