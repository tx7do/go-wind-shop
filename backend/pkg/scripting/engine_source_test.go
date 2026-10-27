package scripting

// 本文件针对 engine.go 的脚本源与业务依赖注入做集成测试：
//   - SetSource/LoadScript：绑定 FileSource 后按 key 加载（缺失 key 报错、
//     存在 key 成功）；
//   - WatchScript：绑定源实现 Watcher → 启动成功，ctx 取消后 watcher 退出；
//   - LoadScriptFile：文件不存在 → "failed to read script file"；
//   - LoadScriptsFromDir：按引擎类型扩展名过滤（Lua 引擎仅加载 .lua、
//     JS 引擎仅加载 .js，以 hook.register 的落点为观测面）；
//   - SetOSS：注入 OSS 客户端前后，oss 模块（Lua 侧 kratos_oss / JS 侧 oss）
//     的 require/typeof 可见性对比（注入前不可见、注入后可见）。

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	bLogger "github.com/tx7do/kratos-bootstrap/logger"

	conf "github.com/tx7do/kratos-bootstrap/api/gen/go/conf/v1"

	"go-wind-shop/pkg/oss"
)

// newOssFixtureClient 构造指向不可达端点的 OSS 客户端（模块注册无需真实连接）。
func newOssFixtureClient(t *testing.T) *oss.MinIOClient {
	t.Helper()
	client := oss.NewMinIoClient(&conf.Bootstrap{
		Oss: &conf.OSS{
			Minio: &conf.OSS_MinIO{
				Endpoint:     "test.invalid:9000",
				UploadHost:   "test.invalid:9000",
				DownloadHost: "test.invalid:9000",
				AccessKey:    "fixture",
				SecretKey:    "fixture",
			},
		},
	}, bLogger.NopLogger())
	require.NotNil(t, client)
	return client
}

// TestEngine_SourceLoadAndWatch 绑定 FileSource 后 LoadScript 的存在/缺失
// key 两分支，与 WatchScript 的启动/取消。
func TestEngine_SourceLoadAndWatch(t *testing.T) {
	e := newTestEngine(t)
	ctx := context.Background()

	dir := t.TempDir()
	key := filepath.Join(dir, "watched.lua")
	require.NoError(t, os.WriteFile(key, []byte("return 1"), 0o644))

	e.SetSource(NewFileSource())

	// 键不存在：底层源加载失败
	require.Error(t, e.LoadScript(ctx, filepath.Join(dir, "missing.lua")))
	// 键存在：加载成功
	require.NoError(t, e.LoadScript(ctx, key))

	// Watch：绑定源实现 Watcher → 启动成功；随后取消 ctx 使 watcher 退出
	wctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	require.NoError(t, e.WatchScript(wctx, key))
	cancel()
}

// TestEngine_LoadScriptFileMissing 文件不存在 → 包装为读取失败错误。
func TestEngine_LoadScriptFileMissing(t *testing.T) {
	e := newTestEngine(t)
	err := e.LoadScriptFile(context.Background(), filepath.Join(t.TempDir(), "nope.lua"))
	require.ErrorContains(t, err, "failed to read script file")
}

// TestEngine_LoadScriptsFromDirExtensionFilter 扩展名过滤：同一目录含
// probe.lua 与 probe.js，Lua 引擎只加载前者、JS 引擎只加载后者
// （以各自编排器上的 hook.register 回调落点为观测面）。
func TestEngine_LoadScriptsFromDirExtensionFilter(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "aaa.lua"), []byte(`
local hook = require "kratos_hook"
hook.register("dirload_probe_lua", "d", function() return true end)
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bbb.js"), []byte(`
hook.register("dirload_probe_js", "d", function () { return true })
`), 0o644))
	ctx := context.Background()

	t.Run("lua engine loads only .lua", func(t *testing.T) {
		e := newTestEngine(t)
		require.NoError(t, e.LoadScriptsFromDir(ctx, dir))
		luaEntry := findHookPoint(t, e, "dirload_probe_lua")
		require.NotNil(t, luaEntry, "lua engine should load the .lua script")
		require.Equal(t, 1, luaEntry.CallbackCount)
		require.Nil(t, findHookPoint(t, e, "dirload_probe_js"), "lua engine must not load the .js script")
	})

	t.Run("js engine loads only .js", func(t *testing.T) {
		e := newTestJSEngine(t)
		require.NoError(t, e.LoadScriptsFromDir(ctx, dir))
		jsEntry := findHookPoint(t, e, "dirload_probe_js")
		require.NotNil(t, jsEntry, "js engine should load the .js script")
		require.Equal(t, 1, jsEntry.CallbackCount)
		require.Nil(t, findHookPoint(t, e, "dirload_probe_lua"), "js engine must not load the .lua script")
		settleJSWatcherRace()
	})
}

// TestEngine_SetOSSEnablesOssModule SetOSS 前后 oss 模块可见性对比。
func TestEngine_SetOSSEnablesOssModule(t *testing.T) {
	client := newOssFixtureClient(t)
	ctx := context.Background()

	t.Run("lua module visibility toggles with SetOSS", func(t *testing.T) {
		e := newTestEngine(t)
		// 注入前：require "kratos_oss" 失败
		require.Error(t, e.LoadScriptString(ctx, "oss_before_lua", `local m = require "kratos_oss"`))
		// 注入后：模块可见且为表
		e.SetOSS(client)
		require.NoError(t, e.LoadScriptString(ctx, "oss_after_lua", `
local m = require "kratos_oss"
assert(type(m) == "table", "kratos_oss should be a table after SetOSS")
`))
	})

	t.Run("js module visibility toggles with SetOSS", func(t *testing.T) {
		// 说明：go-scripts/js 对每次带超时 ctx 的执行都派生一个 ctx.Done 监视
		// goroutine，退出时可能对本引擎 runtime 做一次陈旧 Interrupt——同一
		// 引擎上背靠背的两次执行会互相干扰（见 settleJSWatcherRace 注释）。
		// 因此前后对照各用一个只执行一次的引擎。
		eBefore := newTestJSEngine(t)
		// 注入前：全局 oss 未定义
		require.NoError(t, eBefore.LoadScriptString(ctx, "oss_before_js", `
	if (typeof oss !== "undefined") { throw "oss must be absent before SetOSS" }
	`))
		settleJSWatcherRace()

		eAfter := newTestJSEngine(t)
		// 注入后：全局 oss 已定义
		eAfter.SetOSS(client)
		require.NoError(t, eAfter.LoadScriptString(ctx, "oss_after_js", `
	if (typeof oss === "undefined") { throw "oss must be present after SetOSS" }
	`))
		settleJSWatcherRace()
	})
}
