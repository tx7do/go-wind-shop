package scripting

// 本文件针对 context.go 的 Context 做单元测试（表驱动）：
//   - 类型化读取：GetString / GetInt（int/int32/int64/float64 各分支与截断、
//     非数值归零）/ GetBool / GetMap 的命中与未命中、类型不符分支；
//   - 键管理：Set/Get 往返、Has、Delete（含删除不存在键的 no-op）；
//   - Clone：数据深拷贝独立、指针字段共享、ID/StartTime 刷新、标量字段复制；
//   - 链式构造器：WithUser/WithRequest/WithLogger/WithContext 返回自身并落字段；
//   - Stop：Stopped/StopReason 落值、返回错误、带与不带你 logger 的两个分支；
//   - Duration：非负；
//   - ToMap：基础键恒在，user/request/stop_reason 仅在对应字段非空时出现。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	bLogger "github.com/tx7do/kratos-bootstrap/logger"
)

// TestContext_TypedGetters 类型化读取的表驱动测试：命中、类型不符、键缺失。
func TestContext_TypedGetters(t *testing.T) {
	ctx := NewContext("typed_getters")

	t.Run("get string", func(t *testing.T) {
		ctx.Set("s", "value")
		require.Equal(t, "value", ctx.GetString("s"))
		// 类型不符 / 键缺失 → 空串
		ctx.Set("n", 1)
		require.Empty(t, ctx.GetString("n"))
		require.Empty(t, ctx.GetString("missing"))
	})

	t.Run("get int", func(t *testing.T) {
		for _, c := range []struct {
			name string
			val  any
			want int
		}{
			{"int", int(5), 5},
			{"int32", int32(7), 7},
			{"int64", int64(9), 9},
			{"float whole", float64(2), 2},
			{"float fractional", float64(2.9), 2},
			{"string", "x", 0},
			{"bool", true, 0},
		} {
			ctx.Set("k", c.val)
			require.Equal(t, c.want, ctx.GetInt("k"), "case %s", c.name)
		}
		require.Equal(t, 0, ctx.GetInt("missing"))
	})

	t.Run("get bool", func(t *testing.T) {
		ctx.Set("b", true)
		require.True(t, ctx.GetBool("b"))
		ctx.Set("s", "x")
		require.False(t, ctx.GetBool("s"))
		require.False(t, ctx.GetBool("missing"))
	})

	t.Run("get map", func(t *testing.T) {
		m := map[string]any{"a": 1}
		ctx.Set("m", m)
		require.Equal(t, m, ctx.GetMap("m"))
		ctx.Set("s", "x")
		require.Nil(t, ctx.GetMap("s"))
		require.Nil(t, ctx.GetMap("missing"))
	})
}

// TestContext_KeyManagement Set/Get 往返、Has、Delete（含对不存在键的 no-op）。
func TestContext_KeyManagement(t *testing.T) {
	ctx := NewContext("key_mgmt")

	require.False(t, ctx.Has("k"))
	ctx.Set("k", "v")
	require.True(t, ctx.Has("k"))
	require.Equal(t, "v", ctx.Get("k"))

	ctx.Delete("k")
	require.False(t, ctx.Has("k"))
	require.Nil(t, ctx.Get("k"))

	// 删除不存在的键：no-op，不 panic
	ctx.Delete("never-existed")
}

// TestContext_Clone 克隆语义：数据深拷贝且独立、User/Request/Logger/Cancel
// 指针共享、ID 与 StartTime 刷新、HookName/Stopped/StopReason 复制。
func TestContext_Clone(t *testing.T) {
	orig := NewContext("clone_probe")
	orig.Set("a", "b")
	orig.User = &UserContext{ID: 1, Username: "u"}
	orig.Request = &HTTPContext{Method: "GET", Path: "/p"}
	orig.Logger = bLogger.NewHelper(bLogger.NopLogger())
	orig.Cancel = context.Background()
	orig.Stopped = true
	orig.StopReason = "r"

	clone := orig.Clone()

	// 数据相等但独立：改原对象不影响克隆
	require.Equal(t, "b", clone.Get("a"))
	orig.Set("a", "mutated")
	require.Equal(t, "b", clone.Get("a"))

	// 指针字段共享
	require.Same(t, orig.User, clone.User)
	require.Same(t, orig.Request, clone.Request)
	require.Same(t, orig.Logger, clone.Logger)
	require.Equal(t, orig.Cancel, clone.Cancel)

	// 标量复制；ID 与 StartTime 刷新（克隆时刻不早于原对象创建时刻）
	require.Equal(t, "clone_probe", clone.HookName)
	require.True(t, clone.Stopped)
	require.Equal(t, "r", clone.StopReason)
	require.NotEqual(t, orig.ID, clone.ID)
	require.False(t, clone.StartTime.Before(orig.StartTime))
}

// TestContext_ChainBuilders With* 链式构造器：返回自身并把字段落上。
func TestContext_ChainBuilders(t *testing.T) {
	ctx := NewContext("chain")
	logger := bLogger.NewHelper(bLogger.NopLogger())
	user := &UserContext{ID: 2}
	req := &HTTPContext{Method: "POST"}
	cctx := context.Background()

	require.Same(t, ctx, ctx.WithUser(user))
	require.Same(t, ctx, ctx.WithRequest(req))
	require.Same(t, ctx, ctx.WithLogger(logger))
	require.Same(t, ctx, ctx.WithContext(cctx))

	require.Same(t, user, ctx.User)
	require.Same(t, req, ctx.Request)
	require.Same(t, logger, ctx.Logger)
	require.Equal(t, cctx, ctx.Cancel)
}

// TestContext_Stop Stop 的两个分支（带与不带 logger）：
// 均置位 Stopped/StopReason 并返回错误。
func TestContext_Stop(t *testing.T) {
	withLogger := NewContext("stop_a")
	withLogger.Logger = bLogger.NewHelper(bLogger.NopLogger())
	err := withLogger.Stop("because-a")
	require.EqualError(t, err, "execution stopped: because-a")
	require.True(t, withLogger.Stopped)
	require.Equal(t, "because-a", withLogger.StopReason)

	withoutLogger := NewContext("stop_b")
	err2 := withoutLogger.Stop("because-b")
	require.EqualError(t, err2, "execution stopped: because-b")
	require.True(t, withoutLogger.Stopped)
	require.Equal(t, "because-b", withoutLogger.StopReason)
}

// TestContext_Duration Duration 返回自创建起流逝的时间（非负）。
func TestContext_Duration(t *testing.T) {
	ctx := NewContext("duration")
	require.GreaterOrEqual(t, ctx.Duration(), time.Duration(0))
}

// TestContext_ToMap ToMap 的字段分支：基础键恒在；
// user/request/stop_reason 仅在对应字段非空时出现。
func TestContext_ToMap(t *testing.T) {
	bare := NewContext("tomap_bare")
	m := bare.ToMap()
	require.Equal(t, "tomap_bare", m["hook_name"])
	require.Equal(t, bare.ID, m["id"])
	require.Equal(t, map[string]any{}, m["data"])
	require.Equal(t, false, m["stopped"])
	require.NotEmpty(t, m["duration"])
	_, hasUser := m["user"]
	require.False(t, hasUser)
	_, hasReq := m["request"]
	require.False(t, hasReq)
	_, hasStop := m["stop_reason"]
	require.False(t, hasStop)

	full := NewContext("tomap_full")
	full.Set("k", "v")
	full.User = &UserContext{ID: 3}
	full.Request = &HTTPContext{Method: "GET"}
	full.Stopped = true
	full.StopReason = "boom"
	m2 := full.ToMap()
	require.Equal(t, full.User, m2["user"])
	require.Equal(t, full.Request, m2["request"])
	require.Equal(t, "boom", m2["stop_reason"])
	require.Equal(t, true, m2["stopped"])
	require.Equal(t, map[string]any{"k": "v"}, m2["data"])
}
