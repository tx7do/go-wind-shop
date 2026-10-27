package scripting

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"
)

func TestEngine_LoadScriptFile(t *testing.T) {
	// Self-contained fixture: write a self-registering script to a temp dir,
	// then load it through the explicit single-file entry point.
	scriptContent := `
		local hook = require "kratos_hook"

		hook.register("loadfile.test", "LoadScriptFile test hook", function(ctx)
			ctx.set("loaded_from_file", true)
			return true
		end)
	`

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "loadfile_test.lua")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to write test script: %v", err)
	}

	// ScriptDir empty: no directory auto-load, only the explicit file load below.
	config := DefaultConfig()
	config.ScriptDir = ""

	engine := NewEngine(config, bLogger.NopLogger())
	defer engine.Close()

	if err := engine.LoadScriptFile(context.Background(), scriptPath); err != nil {
		t.Fatalf("Failed to load script file: %v", err)
	}

	// The loaded script should have self-registered its hook; verify by execution.
	ctx := &Context{
		ID:       "test-loadfile",
		HookName: "loadfile.test",
		Data:     make(map[string]interface{}),
	}

	if err := engine.ExecuteHook(context.Background(), "loadfile.test", ctx); err != nil {
		t.Fatalf("Failed to execute hook: %v", err)
	}

	if loaded, ok := ctx.Data["loaded_from_file"].(bool); !ok || !loaded {
		t.Error("Loaded script callback was not executed")
	}

	t.Log("✓ LoadScriptFile test passed")
}
