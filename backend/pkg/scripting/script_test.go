package scripting

// 本文件针对 script.go 的 Script 做单元测试（表驱动）：
//   - Hash：源码 SHA-256 十六进制摘要（标准测试向量）；
//   - Validate：名称/钩子/源码三项缺失的各自错误分支与合法通过分支；
//   - Clone：字段逐一复制且返回新实例。
//
// 注意：Clone 不复制 Language 字段（生产行为，按现状钉死）。

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestScriptHash Hash 的标准测试向量："abc" 与空串。
func TestScriptHash(t *testing.T) {
	// "abc" 的 SHA-256（公开标准向量）
	s := &Script{Source: "abc"}
	h := sha256.Sum256([]byte("abc"))
	require.Equal(t, hex.EncodeToString(h[:]), s.Hash())

	// 空源码
	empty := sha256.Sum256([]byte{})
	require.Equal(t, hex.EncodeToString(empty[:]), (&Script{}).Hash())
}

// TestScriptValidate Validate 的表驱动测试：任一必填字段缺失即报错，全填则通过。
func TestScriptValidate(t *testing.T) {
	cases := []struct {
		name    string
		script  Script
		wantErr string
	}{
		{"missing name", Script{Hook: "h", Source: "s"}, "script name is required"},
		{"missing hook", Script{Name: "n", Source: "s"}, "hook name is required"},
		{"missing source", Script{Name: "n", Hook: "h"}, "script source is required"},
		{"missing all", Script{}, "script name is required"},
		{"valid", Script{Name: "n", Hook: "h", Source: "s"}, ""},
	}
	for _, c := range cases {
		err := c.script.Validate()
		if c.wantErr == "" {
			require.NoError(t, err, "case %s", c.name)
		} else {
			require.EqualError(t, err, c.wantErr, "case %s", c.name)
		}
	}
}

// TestScriptClone Clone 复制全部受支持字段并返回新实例（指针不同）。
func TestScriptClone(t *testing.T) {
	now := time.Now()
	orig := &Script{
		ID:          7,
		Name:        "n",
		Hook:        "h",
		Source:      "s",
		Enabled:     true,
		Priority:    3,
		Description: "d",
		Version:     2,
		Author:      "a",
		Critical:    true,
		CreateTime:  now,
		UpdateTime:  now,
	}
	clone := orig.Clone()

	require.NotSame(t, orig, clone)
	require.Equal(t, orig.ID, clone.ID)
	require.Equal(t, orig.Name, clone.Name)
	require.Equal(t, orig.Hook, clone.Hook)
	require.Equal(t, orig.Source, clone.Source)
	require.Equal(t, orig.Enabled, clone.Enabled)
	require.Equal(t, orig.Priority, clone.Priority)
	require.Equal(t, orig.Description, clone.Description)
	require.Equal(t, orig.Version, clone.Version)
	require.Equal(t, orig.Author, clone.Author)
	require.Equal(t, orig.Critical, clone.Critical)
	require.Equal(t, orig.CreateTime, clone.CreateTime)
	require.Equal(t, orig.UpdateTime, clone.UpdateTime)
}
