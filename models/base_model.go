package models

import "time"

type BotBase struct {
	ID     string
	Secret string
}

type CommandBase struct {
	Command        []string // 可触发事件的指令列表，与正则 Regex 互斥，优先使用此项
	Regex          string   // 可触发指令的正则表达式，与指令表 Command 互斥
	RequireAT      bool     // 是否要求必须@机器人才能触发指令
	IsShortCircuit bool     // 如果触发指令成功是否短路不运行后续指令（将根据注册顺序排序指令的短路机制）
}

type Scope uint8

const (
	ScopeGlobal Scope = 1 << 0
	ScopeVilla  Scope = 1 << 1
	ScopeRoom   Scope = 1 << 2
	ScopeUser   Scope = 1 << 3
)

type WaitForCommandRegister struct {
	Scope       Scope
	Command     CommandBase
	Data        interface{}
	Timeout     *time.Duration // 超时时间，如果为0则不超时，nil默认60秒
	Identify    *string        // 用于标识该注册的字符串，用于验证是否重复或取消注册
	AllowRepeat *bool          // 是否允许重复触发（仅在Identify不为nil时生效），nil默认true
}
