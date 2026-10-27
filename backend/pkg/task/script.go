package task

import "encoding/json"

// ScriptTaskDispatchType 是全部脚本任务共用的 asynq 分发类型。
//
// 设计约束：asynq 的 mux 拒绝 server Start 之后注册 handler（"server already started"），
// 而脚本处理器在运行期随脚本增删动态变化、无法在启动期枚举——因此订阅固定为一个
// 分发类型，真正的处理器名放在消息载荷 handler 字段里，由脚本运行时分发。
//
// sys_tasks 行的写法：type=PERIODIC，type_name="script_task"，cron_spec=cron，
// task_payload={"handler":"<处理器名>","params":{...}}（params 可省略）。
const ScriptTaskDispatchType = "script_task"

// ScriptTaskData 是脚本任务处理器的 asynq 消息载荷。
// sys_tasks 行的 task_payload JSON 会原样解码到本结构：
//
//	{"handler": "cleanup_tmp", "params": {"older_than": 86400}}
type ScriptTaskData struct {
	Handler string         `json:"handler"`
	Params  map[string]any `json:"params,omitempty"`
}

// ScriptTaskPayloadFromRaw 把 sys_tasks 的 task_payload 原始 JSON 解为任务载荷。
func ScriptTaskPayloadFromRaw(raw []byte) (*ScriptTaskData, error) {
	data := &ScriptTaskData{}
	if len(raw) == 0 {
		return data, nil
	}
	if err := json.Unmarshal(raw, data); err != nil {
		return nil, err
	}
	return data, nil
}
