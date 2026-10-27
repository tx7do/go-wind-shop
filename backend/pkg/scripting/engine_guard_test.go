package scripting

// 本文件针对 engine.go 的工厂与降级守卫做单元测试：
//   - SetEngineFactory：nil 工厂为 no-op；自定义工厂安装后 NewEngine 采用之，
//     并以其 Close 报错覆盖 Engine.Close 的引擎关闭失败分支（编排器仍返回 nil）；
//   - NewEngineWithFactory：nil 配置/nil 工厂回退默认值；工厂报错降级到
//     默认 Lua 引擎；
//   - 白盒空引擎（scriptEngine == nil）：Execute / LoadScript / WatchScript /
//     LoadScriptString 一律返回 "no script engine available"，SetSource /
//     rebind 走 nil 守卫早退，Close 走 nil 引擎分支；
//   - 白盒引擎带引擎无 binder：SetOSS 触发 rebind 的 binder-nil 守卫早退。
//
// fakeEngine 借助接口嵌入满足 gsEngine.Engine 全接口，仅实现本流程会触达的
// GetType/Init/Close，其余方法一旦被调用即 panic（测试即失败）。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	gsEngine "github.com/tx7do/go-scripts"
	bLogger "github.com/tx7do/kratos-bootstrap/logger"

	"github.com/tx7do/go-scripts/hook"
)

// fakeEngine 测试用假引擎：GetType 固定 LuaType（以便走 Lua binder 的装配
// 分支），Init 恒成功，Close 返回注入的错误。
type fakeEngine struct {
	gsEngine.Engine
	closeErr error
}

func (f *fakeEngine) GetType() gsEngine.Type       { return gsEngine.LuaType }
func (f *fakeEngine) Init(_ context.Context) error { return nil }
func (f *fakeEngine) Close() error                 { return f.closeErr }

// TestSetEngineFactory 工厂覆盖语义：nil 不替换默认工厂；自定义工厂生效且
// NewEngine 返回的编排器持有假引擎；假引擎 Close 报错仅记录、编排器 Close
// 返回 nil（覆盖引擎关闭失败分支）。
func TestSetEngineFactory(t *testing.T) {
	orig := defaultEngineFactory
	t.Cleanup(func() { defaultEngineFactory = orig })

	t.Run("nil factory is a no-op", func(t *testing.T) {
		SetEngineFactory(nil)
		e := NewEngine(&Config{ScriptDir: ""}, bLogger.NopLogger())
		defer func() { _ = e.Close() }()
		_, isFake := e.ScriptEngine().(*fakeEngine)
		require.False(t, isFake, "nil factory must not replace the default factory")
		require.Equal(t, gsEngine.LuaType, e.ScriptEngine().GetType())
	})

	t.Run("custom factory installed and used", func(t *testing.T) {
		fake := &fakeEngine{closeErr: errors.New("fake close failure")}
		SetEngineFactory(func(_ *Config, _ bLogger.Logger) (gsEngine.Engine, error) {
			return fake, nil
		})
		e := NewEngine(&Config{ScriptDir: ""}, bLogger.NopLogger())
		require.Same(t, fake, e.ScriptEngine())
		// 假引擎 Close 报错：编排器记录后返回 nil
		require.NoError(t, e.Close())
	})
}

// TestNewEngineWithFactoryDefaults nil 配置与 nil 工厂回退默认值；
// 工厂报错时降级到默认 Lua 引擎。
func TestNewEngineWithFactoryDefaults(t *testing.T) {
	t.Run("nil config and nil factory fall back to defaults", func(t *testing.T) {
		e := NewEngineWithFactory(nil, bLogger.NopLogger(), nil)
		defer func() { _ = e.Close() }()
		require.NotNil(t, e.ScriptEngine())
		require.Equal(t, gsEngine.LuaType, e.ScriptEngine().GetType())
		require.NotNil(t, e.config)
		require.Equal(t, DefaultConfig().PoolSize, e.config.PoolSize)
	})

	t.Run("factory error falls back to default engine", func(t *testing.T) {
		e := NewEngineWithFactory(&Config{ScriptDir: ""}, bLogger.NopLogger(), func(_ *Config, _ bLogger.Logger) (gsEngine.Engine, error) {
			return nil, errors.New("factory boom")
		})
		defer func() { _ = e.Close() }()
		require.NotNil(t, e.ScriptEngine())
		require.Equal(t, gsEngine.LuaType, e.ScriptEngine().GetType())
	})
}

// TestEngine_NilEngineGuards 白盒空引擎：全部加载/执行入口统一报
// "no script engine available"，SetSource 与 SetOSS 的 rebind 走 nil 守卫，
// Close 走 nil 引擎分支并正常清空回调表。
func TestEngine_NilEngineGuards(t *testing.T) {
	e := &Engine{
		config:    DefaultConfig(),
		logger:    bLogger.NewHelper(bLogger.NopLogger()),
		registry:  hook.NewRegistry(),
		callbacks: make(map[string][]ScriptCallback),
	}

	require.EqualError(t, e.Execute(context.Background(), &Script{Name: "x", Source: "return 1"}, nil), "no script engine available")
	require.EqualError(t, e.LoadScript(context.Background(), "k"), "no script engine available")
	require.EqualError(t, e.WatchScript(context.Background(), "k"), "no script engine available")
	require.EqualError(t, e.LoadScriptString(context.Background(), "n", "return 1"), "no script engine available")

	e.SetSource(nil) // nil 引擎：no-op
	e.SetOSS(nil)    // rebind 的 nil-scriptEngine 守卫早退

	require.NoError(t, e.Close())
	require.Empty(t, e.callbacks)
}

// TestEngine_RebindNilBinderGuard 引擎非 nil 而 binder 为 nil：
// SetOSS 触发的 rebind 走 binder-nil 守卫早退（不触碰引擎）。
func TestEngine_RebindNilBinderGuard(t *testing.T) {
	fake := &fakeEngine{}
	e := &Engine{
		config:       DefaultConfig(),
		logger:       bLogger.NewHelper(bLogger.NopLogger()),
		registry:     hook.NewRegistry(),
		callbacks:    make(map[string][]ScriptCallback),
		scriptEngine: fake,
	}

	e.SetOSS(nil) // rebind：binder == nil → 早退
	require.NoError(t, e.Close())
}
