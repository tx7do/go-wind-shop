package scripting

// 本文件针对 errors.go 的 IsScriptVetoed 做单元测试（表驱动）：
// 直接哨兵错误与经 fmt.Errorf %w 包装的均判定为否决；
// 无关错误与 nil 判定为非否决。

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIsScriptVetoed 表驱动验证否决哨兵的识别。
func TestIsScriptVetoed(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"direct sentinel", ErrScriptVetoed, true},
		{"wrapped sentinel", fmt.Errorf("wrapped: %w", ErrScriptVetoed), true},
		{"double wrapped", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrScriptVetoed)), true},
		{"unrelated", errors.New("ordinary failure"), false},
		{"nil", nil, false},
	}
	for _, c := range cases {
		require.Equal(t, c.want, IsScriptVetoed(c.err), "case %s", c.name)
	}
}
