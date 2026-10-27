package scripting

import "errors"

// 本文件原承载自研 LuaEngine 的哨兵错误与适配器。
// 切换到 go-scripts/lua 后，引擎错误由 go-scripts 提供（ErrLuaEngineNotInitialized 等），
// hook 引擎适配器移至 engine_runtime.go，故本文件仅保留包级注释占位。
//
// 如未来需要 pkg/scripting 自有的错误类型，可在此定义。

// ErrScriptVetoed 实体 before 钩子否决错误标记：业务层应以 400 语义透传，
// 而非 500（用户请求被规则拒绝）。由 app 层钩子桥包装后返回。
var ErrScriptVetoed = errors.New("script veto")

// IsScriptVetoed 判断错误是否为脚本否决。
func IsScriptVetoed(err error) bool {
	return errors.Is(err, ErrScriptVetoed)
}
