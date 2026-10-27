package scripting

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"
)

// TestExamplesLoad 守卫 examples/ 目录下的全部示例脚本：
// 每个示例必须能在干净引擎上完整加载（解析 + 顶层执行），
// 防止示例随引擎 API 演进腐化。
//
// 本引擎实例未配置 Redis / EventBusManager / OSS，因此这三个可选模块
// 以空表打桩注册（打桩方式与 runtime_lua.go 的模块注册完全同构）。
// 示例的顶层只做 require 与钩子注册，对可选模块的调用都在回调或
// 挂载脚本体内——加载阶段不会触达打桩表的成员。
func TestExamplesLoad(t *testing.T) {
	const examplesDir = "examples"

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("read examples dir: %v", err)
	}

	// 空表打桩：require 命中后返回空 table
	stubBuilder := func(L *lua.LState) int {
		L.Push(L.NewTable())
		return 1
	}

	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".lua" {
			continue
		}

		src, err := os.ReadFile(filepath.Join(examplesDir, entry.Name()))
		if err != nil {
			t.Errorf("%s: read: %v", entry.Name(), err)
			continue
		}

		cfg := DefaultConfig()
		cfg.ScriptDir = "" // 不启用目录自动加载，逐文件显式加载

		engine := NewEngine(cfg, bLogger.NopLogger())

		for _, name := range []string{"kratos_cache", "kratos_eventbus", "kratos_oss"} {
			if err := engine.ScriptEngine().RegisterModule(name, luaPreloadAdapter(stubBuilder)); err != nil {
				t.Fatalf("stub module %s: %v", name, err)
			}
		}

		if err := engine.LoadScriptString(context.Background(), entry.Name(), string(src)); err != nil {
			t.Errorf("%s: load: %v", entry.Name(), err)
		}

		_ = engine.Close()
		loaded++
	}

	if loaded == 0 {
		t.Fatal("no example scripts found in examples/")
	}
	t.Logf("✓ %d example scripts loaded cleanly", loaded)
}
