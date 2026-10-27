package scripting

// 本文件针对 source_db.go 的 DBSource 做补充单元测试（表驱动）：
//   - signature 的四个分支：hasher 正常返回指纹、hasher 报错回退源码、
//     无 hasher 直接取源码、loader 报错返回空指纹（避免误触发）；
//   - InvalidateAll：缓存命中不回源、全量失效后回源；
//   - Load 在已取消的 context 上直接报 context.Canceled。
//
// loader/hasher 均为注入的内存回调，无需真实数据库。

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDBSource_SignatureBranches signature 的表驱动分支测试。
func TestDBSource_SignatureBranches(t *testing.T) {
	ctx := context.Background()
	loader := func(_ context.Context, _ string) (string, error) {
		return "loader-code", nil
	}

	t.Run("hasher result wins", func(t *testing.T) {
		hasher := func(context.Context, string) (string, error) { return "hasher-sig", nil }
		src := NewDBSource(loader, WithScriptHasher(hasher))
		require.Equal(t, "hasher-sig", src.signature(ctx, "k"))
	})

	t.Run("hasher error falls back to source", func(t *testing.T) {
		hasher := func(context.Context, string) (string, error) { return "", errors.New("hash boom") }
		src := NewDBSource(loader, WithScriptHasher(hasher))
		require.Equal(t, "loader-code", src.signature(ctx, "k"))
	})

	t.Run("no hasher uses source", func(t *testing.T) {
		src := NewDBSource(loader)
		require.Equal(t, "loader-code", src.signature(ctx, "k"))
	})

	t.Run("loader error yields empty signature", func(t *testing.T) {
		failing := func(context.Context, string) (string, error) { return "", errors.New("load boom") }
		src := NewDBSource(failing)
		require.Empty(t, src.signature(ctx, "k"))
	})
}

// TestDBSource_InvalidateAll 缓存命中不回源、InvalidateAll 全量失效后回源。
func TestDBSource_InvalidateAll(t *testing.T) {
	ctx := context.Background()
	calls := 0
	loader := func(_ context.Context, key string) (string, error) {
		calls++
		return fmt.Sprintf("v%d", calls), nil
	}
	src := NewDBSource(loader)

	v1, err := src.Load(ctx, "k")
	require.NoError(t, err)
	require.Equal(t, "v1", v1)
	require.Equal(t, 1, calls)

	v2, err := src.Load(ctx, "k")
	require.NoError(t, err)
	require.Equal(t, "v1", v2, "cache hit must not re-invoke the loader")
	require.Equal(t, 1, calls)

	src.InvalidateAll()

	v3, err := src.Load(ctx, "k")
	require.NoError(t, err)
	require.Equal(t, "v2", v3, "after InvalidateAll the loader must be re-invoked")
	require.Equal(t, 2, calls)
}

// TestDBSource_LoadLoaderError Load 在回源加载失败时包装并返回错误。
func TestDBSource_LoadLoaderError(t *testing.T) {
	failing := func(context.Context, string) (string, error) { return "", errors.New("load boom") }
	src := NewDBSource(failing)
	_, err := src.Load(context.Background(), "k")
	require.ErrorContains(t, err, "db source: load \"k\"")
}

// TestDBSource_LoadCanceledContext 已取消的 context 直接报 context.Canceled。
func TestDBSource_LoadCanceledContext(t *testing.T) {
	src := NewDBSource(func(context.Context, string) (string, error) { return "v", nil })
	cctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := src.Load(cctx, "k")
	require.ErrorIs(t, err, context.Canceled)
}
