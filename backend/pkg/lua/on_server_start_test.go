package lua

import (
	bLogger "github.com/tx7do/kratos-bootstrap/logger"
	"context"
	"testing"

)

func TestOnServerStartScript(t *testing.T) {
	// Create engine
	config := DefaultConfig()
	config.ScriptDir = "" // Don't auto-load
	engine := NewEngine(config, bLogger.GetLogger())
	defer engine.Close()

	// Load the on_server_start.lua script
	scriptPath := "../../app/admin/service/scripts/on_server_start.lua"
	err := engine.LoadScriptFile(context.Background(), scriptPath)
	if err != nil {
		t.Fatalf("Failed to load on_server_start.lua: %v", err)
	}

	// Execute the hook with test data
	ctx := &Context{
		ID:       "server_startup",
		HookName: "on_server_start",
		Data: map[string]interface{}{
			"config": map[string]interface{}{
				"server": "test-server",
				"port":   8080,
			},
			"service": map[string]interface{}{
				"name":    "go-wind-shop",
				"version": "1.0.0",
			},
		},
	}

	err = engine.ExecuteHook(context.Background(), "on_server_start", ctx)
	if err != nil {
		t.Fatalf("Failed to execute on_server_start hook: %v", err)
	}

	t.Log("✓ on_server_start script executed successfully")
	t.Log("✓ All logs from the script should be visible above")
}
