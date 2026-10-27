package scripting

// 本文件针对 source_file.go 的 FileSource 做补充单元测试（表驱动）：
//   - Close：no-op，恒返回 nil；
//   - resolveKey：空键、绝对路径、无搜索路径、有搜索路径（取首个拼接结果）
//     四个分支（搜索路径存在时永远返回第一个路径的拼接结果，交由底层
//     Load 处理不存在情形，按生产行为钉死）。

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFileSource_CloseNoop Close 为 no-op。
func TestFileSource_CloseNoop(t *testing.T) {
	require.NoError(t, NewFileSource().Close())
}

// TestFileSource_ResolveKeyTable resolveKey 的表驱动测试。
func TestFileSource_ResolveKeyTable(t *testing.T) {
	abs := filepath.Join(t.TempDir(), "a.lua")
	cases := []struct {
		name string
		src  *FileSource
		key  string
		want string
	}{
		{"empty key returned as-is", NewFileSource(), "", ""},
		{"absolute key returned as-is", NewFileSource(), abs, abs},
		{"no search paths returns key as-is", NewFileSource(), "rel.lua", "rel.lua"},
		{"first search path wins", NewFileSource("p1", "p2"), "rel.lua", filepath.Clean(filepath.Join("p1", "rel.lua"))},
	}
	for _, c := range cases {
		require.Equal(t, c.want, c.src.resolveKey(c.key), "case %s", c.name)
	}
}
