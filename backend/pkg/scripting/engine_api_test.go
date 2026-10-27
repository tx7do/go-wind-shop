package scripting

// 本文件针对 engine.go 的编排器 API 做单元/集成测试：
//   - HTTPAllowlistFromEnv：环境变量白名单解析的表驱动测试（未设置/空白/
//     单域名/多域名含空白项）；
//   - scriptExtForType：引擎类型 → 脚本扩展名映射表；
//   - extractScriptFields：反射提取兼容结构体字段的表驱动测试（完整结构、
//     缺 Name/Source、字段类型不符、非结构体、nil 指针等分支）；
//   - Engine.AddScript：兼容结构体分支与非法类型分支；
//   - Engine.RemoveScript：移除/脚本不存在/钩子不存在；
//   - Engine.HookPoints / ResetScriptRegistrations：挂载与回调计数、全量重置；
//   - Engine.ExecuteHook 分支：空钩子早退、禁用脚本跳过、脚本语法错误、
//     execute() 返回 false、execute() 抛错、回调执行失败；
//   - Engine.Execute：直调路径；
//   - Engine.ExecuteTaskHandler：处理器存在（经脚本注册）与不存在两个分支。
//
// 引擎构造统一走 newTestEngine / newTestJSEngine（默认配置、无脚本目录、
// 测试尾自动 Close）。

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	gsEngine "github.com/tx7do/go-scripts"
	"github.com/tx7do/go-scripts/hostmodule"
)

// findHookPoint 在 HookPoints 结果中按名称查找条目（未找到返回 nil）。
func findHookPoint(t *testing.T, e *Engine, name string) *HookPointInfo {
	t.Helper()
	for _, hp := range e.HookPoints() {
		if hp.Name == name {
			return &hp
		}
	}
	return nil
}

// fakeScriptCallback 测试用语言无关回调：Call 返回固定 (true, err)。
type fakeScriptCallback struct {
	err error
}

func (f *fakeScriptCallback) Call(_ context.Context, _ *Context) (any, error) {
	return true, f.err
}

func (f *fakeScriptCallback) Source() string { return "fake-script-callback" }

// compatFull 字段名与类型完全兼容 extractScriptFields 的结构体。
type compatFull struct {
	Name, Hook, Source, Description, Author string
	Enabled, Critical                       bool
	Priority                                int
}

// compatNoName 缺 Name 字段。
type compatNoName struct {
	Hook, Source string
}

// compatNoSource 缺 Source 字段。
type compatNoSource struct {
	Name, Hook string
}

// wrongTypes 字段类型不符（Name/Enabled/Priority 均非预期类型）。
type wrongTypes struct {
	Name    int
	Source  string
	Enabled string
}

// nested 含嵌套结构体、无兼容字段。
type nested struct {
	Inner *compatFull
}

// TestHTTPAllowlistFromEnv 白名单解析表驱动测试：
// 未设置/纯空白 → 空白名单（fail-closed）；单域名/多域名（含空白项）→ 去空白收集。
func TestHTTPAllowlistFromEnv(t *testing.T) {
	t.Run("unset env yields empty allowlist", func(t *testing.T) {
		orig, had := os.LookupEnv(hostmodule.EnvHTTPAllowedDomains)
		if err := os.Unsetenv(hostmodule.EnvHTTPAllowedDomains); err != nil {
			t.Fatalf("unset env: %v", err)
		}
		t.Cleanup(func() {
			if had {
				_ = os.Setenv(hostmodule.EnvHTTPAllowedDomains, orig)
			}
		})
		require.Empty(t, hostmodule.HTTPAllowlistFromEnv().AllowedDomains)
	})

	t.Run("blank value yields empty allowlist", func(t *testing.T) {
		t.Setenv(hostmodule.EnvHTTPAllowedDomains, "   ")
		require.Empty(t, hostmodule.HTTPAllowlistFromEnv().AllowedDomains)
	})

	t.Run("single domain", func(t *testing.T) {
		t.Setenv(hostmodule.EnvHTTPAllowedDomains, "a.example.com")
		require.Equal(t, []string{"a.example.com"}, hostmodule.HTTPAllowlistFromEnv().AllowedDomains)
	})

	t.Run("multiple domains with blank entries", func(t *testing.T) {
		t.Setenv(hostmodule.EnvHTTPAllowedDomains, " a.example.com , b.example.com ,, ")
		require.Equal(t, []string{"a.example.com", "b.example.com"}, hostmodule.HTTPAllowlistFromEnv().AllowedDomains)
	})
}

// TestScriptExtForType 引擎类型 → 扩展名映射表（未知类型回退 .lua）。
func TestScriptExtForType(t *testing.T) {
	require.Equal(t, ".lua", scriptExtForType(gsEngine.LuaType))
	require.Equal(t, ".js", scriptExtForType(gsEngine.JavaScriptType))
	require.Equal(t, ".lua", scriptExtForType(gsEngine.Type("unknown-engine-type")))
}

// TestExtractScriptFields 反射提取表驱动测试：完整结构（值与指针两种形态）
// 提取全部八字段；缺字段/类型不符/非结构体/nil 指针均拒绝。
func TestExtractScriptFields(t *testing.T) {
	full := &compatFull{
		Name: "n", Hook: "h", Source: "s", Description: "d", Author: "a",
		Enabled: true, Critical: true, Priority: 4,
	}

	t.Run("pointer to full struct", func(t *testing.T) {
		hs, ok := extractScriptFields(full)
		require.True(t, ok)
		require.Equal(t, "n", hs.Name)
		require.Equal(t, "h", hs.Hook)
		require.Equal(t, "s", hs.Source)
		require.Equal(t, "d", hs.Description)
		require.Equal(t, "a", hs.Author)
		require.True(t, hs.Enabled)
		require.True(t, hs.Critical)
		require.Equal(t, 4, hs.Priority)
	})

	t.Run("value struct", func(t *testing.T) {
		hs, ok := extractScriptFields(*full)
		require.True(t, ok)
		require.Equal(t, "n", hs.Name)
		require.Equal(t, "s", hs.Source)
	})

	t.Run("rejected inputs", func(t *testing.T) {
		cases := []struct {
			name   string
			script any
		}{
			{"missing name", &compatNoName{Hook: "h", Source: "s"}},
			{"missing source", &compatNoSource{Name: "n", Hook: "h"}},
			{"missing both", struct{}{}},
			{"wrong field types", &wrongTypes{Name: 1, Source: "s"}},
			{"nested without compatible fields", &nested{}},
			{"int", 42},
			{"string", "str"},
			{"map", map[string]int{}},
			{"nil pointer", (*compatFull)(nil)},
		}
		for _, c := range cases {
			hs, ok := extractScriptFields(c.script)
			require.False(t, ok, "case %s should be rejected", c.name)
			require.Nil(t, hs, "case %s should yield no script", c.name)
		}
	})
}

// TestEngine_AddScriptCompatAndInvalid AddScript 的兼容结构体分支与
// 非法类型分支（含提取失败的结构体）。
func TestEngine_AddScriptCompatAndInvalid(t *testing.T) {
	e := newTestEngine(t)

	require.NoError(t, e.RegisterHook("compat_probe", "compat probe hook"))
	require.NoError(t, e.AddScript("compat_probe", &compatFull{
		Name: "compat_script", Hook: "compat_probe", Source: "return 1",
		Description: "d", Author: "a", Enabled: true, Priority: 1,
	}))
	hp := findHookPoint(t, e, "compat_probe")
	require.NotNil(t, hp)
	require.Equal(t, "compat_probe", hp.Name)
	require.Equal(t, "compat probe hook", hp.Description)
	require.Equal(t, 1, hp.ScriptCount)
	require.Equal(t, 0, hp.CallbackCount)

	// 非结构体类型 → 无效类型错误
	require.ErrorContains(t, e.AddScript("compat_probe", 42), "invalid script type")
	// 结构体但缺必填字段 → 提取失败 → 无效类型错误
	require.ErrorContains(t, e.AddScript("compat_probe", &compatNoName{Hook: "h", Source: "s"}), "invalid script type")
}

// TestEngine_RemoveScript 移除脚本、脚本不存在、钩子不存在三个分支。
func TestEngine_RemoveScript(t *testing.T) {
	e := newTestEngine(t)

	require.NoError(t, e.AddScript("rm_probe", &Script{Name: "rm_script", Hook: "rm_probe", Source: "return 1", Enabled: true, Priority: 1}))
	require.NotNil(t, findHookPoint(t, e, "rm_probe"))
	require.Equal(t, 1, findHookPoint(t, e, "rm_probe").ScriptCount)

	require.NoError(t, e.RemoveScript("rm_probe", "rm_script"))
	require.Equal(t, 0, findHookPoint(t, e, "rm_probe").ScriptCount)

	require.ErrorContains(t, e.RemoveScript("rm_probe", "missing_script"), "script not found")
	require.ErrorContains(t, e.RemoveScript("no_such_hook", "x"), "hook not found")
}

// TestEngine_HookPointsAndReset HookPoints 计数（脚本 + 公开 RegisterCallback
// 注册的回调）与 ResetScriptRegistrations 全量清空。
func TestEngine_HookPointsAndReset(t *testing.T) {
	e := newTestEngine(t)

	require.NoError(t, e.RegisterHook("hp_probe", "hp probe hook"))
	require.NoError(t, e.AddScript("hp_probe", &Script{Name: "hp_script", Hook: "hp_probe", Source: "return 1", Enabled: true, Priority: 1}))
	e.RegisterCallback("hp_probe", &fakeScriptCallback{})

	hp := findHookPoint(t, e, "hp_probe")
	require.NotNil(t, hp)
	require.Equal(t, 1, hp.ScriptCount)
	require.Equal(t, 1, hp.CallbackCount)

	e.ResetScriptRegistrations()
	require.Empty(t, e.HookPoints())
	require.Empty(t, e.ListHooks())
}

// TestEngine_ExecuteHookBranches ExecuteHook 的分支行为。
func TestEngine_ExecuteHookBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("empty hook returns nil", func(t *testing.T) {
		e := newTestEngine(t)
		require.NoError(t, e.RegisterHook("eh_empty", "d"))
		require.NoError(t, e.ExecuteHook(ctx, "eh_empty", nil))
	})

	t.Run("disabled script skipped while enabled runs", func(t *testing.T) {
		e := newTestEngine(t)
		require.NoError(t, e.AddScript("eh_disabled", &Script{
			Name: "disabled_one", Hook: "eh_disabled", Enabled: false, Priority: 1,
			Source: "function execute() local c = __get_ctx() c.set(\"disabled_ran\", \"yes\") return true end",
		}))
		require.NoError(t, e.AddScript("eh_disabled", &Script{
			Name: "enabled_one", Hook: "eh_disabled", Enabled: true, Priority: 2,
			Source: "function execute() local c = __get_ctx() c.set(\"enabled_ran\", \"yes\") return true end",
		}))
		execCtx := NewContext("eh_disabled")
		require.NoError(t, e.ExecuteHook(ctx, "eh_disabled", execCtx))
		require.False(t, execCtx.Has("disabled_ran"), "disabled script must be skipped")
		require.Equal(t, "yes", execCtx.GetString("enabled_ran"), "enabled script must run")
	})

	t.Run("syntax error propagates", func(t *testing.T) {
		e := newTestEngine(t)
		require.NoError(t, e.AddScript("eh_syntax", &Script{
			Name: "syntax_bad", Hook: "eh_syntax", Enabled: true, Priority: 1,
			Source: "local x = (",
		}))
		err := e.ExecuteHook(ctx, "eh_syntax", nil)
		require.ErrorContains(t, err, "script 'syntax_bad' failed")
	})

	t.Run("execute returning false aborts", func(t *testing.T) {
		e := newTestEngine(t)
		require.NoError(t, e.AddScript("eh_false", &Script{
			Name: "false_script", Hook: "eh_false", Enabled: true, Priority: 1,
			Source: "function execute() return false end",
		}))
		err := e.ExecuteHook(ctx, "eh_false", nil)
		require.ErrorContains(t, err, "script returned false")
	})

	t.Run("execute raising error propagates", func(t *testing.T) {
		e := newTestEngine(t)
		require.NoError(t, e.AddScript("eh_raise", &Script{
			Name: "raising_script", Hook: "eh_raise", Enabled: true, Priority: 1,
			Source: "function execute() error(\"boom\") end",
		}))
		err := e.ExecuteHook(ctx, "eh_raise", nil)
		require.ErrorContains(t, err, "execute function error")
	})

	t.Run("failing callback aborts", func(t *testing.T) {
		e := newTestEngine(t)
		e.RegisterCallback("eh_cb", &fakeScriptCallback{err: context.DeadlineExceeded})
		err := e.ExecuteHook(ctx, "eh_cb", nil)
		require.ErrorContains(t, err, "callback 1 failed")
	})
}

// TestEngine_ExecuteDirect Execute 直调路径：脚本定义 execute() 返回 true，
// 直接经 Execute（而非 ExecuteHook）执行成功。
func TestEngine_ExecuteDirect(t *testing.T) {
	e := newTestEngine(t)
	err := e.Execute(context.Background(), &Script{
		Name: "direct_exec", Source: "function execute() return true end",
	}, NewContext("direct_exec"))
	require.NoError(t, err)
}

// TestEngine_ExecuteTaskHandler 任务桥执行入口：脚本注册的处理器经
// ExecuteTaskHandler 执行成功；未注册的处理器报不存在。
func TestEngine_ExecuteTaskHandler(t *testing.T) {
	e := newTestEngine(t)

	source := `
local task = require "task"
task.register_handler("etd_probe_handler", "probe handler", function(params)
	return true
end, {
	optional = {},
	required = {}
})
`
	require.NoError(t, e.LoadScriptString(context.Background(), "etd_reg", source))
	require.NoError(t, e.ExecuteTaskHandler(context.Background(), "etd_probe_handler", map[string]any{}))
	err := e.ExecuteTaskHandler(context.Background(), "etd_missing_handler", nil)
	require.ErrorContains(t, err, "task handler not found")
}
