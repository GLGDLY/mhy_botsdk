package bot

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	events "github.com/GLGDLY/mhy_botsdk/events"
	models "github.com/GLGDLY/mhy_botsdk/models"
	utils "github.com/GLGDLY/mhy_botsdk/utils"
)

/* private */
type waitForCommandRegister struct {
	register models.WaitForCommandRegister
	regex    *regexp.Regexp
	villa_id string
	room_id  string
	uid      string
	channel  chan *events.EventSendMessage
	cancel   chan bool
}

func (_bot *bot) validateWaitForCommandScope(reg waitForCommandRegister, data events.EventSendMessage) bool {
	if reg.register.Scope&models.ScopeGlobal != 0 { // always true if global scope is enabled
		return true
	}
	if reg.register.Scope&models.ScopeVilla != 0 {
		if reg.villa_id != utils.String(data.Data.VillaId) {
			return false
		}
	}
	if reg.register.Scope&models.ScopeRoom != 0 {
		if reg.room_id != utils.String(data.Data.RoomId) {
			return false
		}
	}
	if reg.register.Scope&models.ScopeUser != 0 {
		if reg.uid != utils.String(data.Data.FromUserId) {
			return false
		}
	}
	return true // if all scope is satisfied, return true
}

// return true if the short circuit is needed
func (_bot *bot) checkWaifForCommand(data events.EventSendMessage) bool {
	msg := data.GetContent(false)
	at := "@" + data.Robot.Template.Name
	for _, reg := range _bot.wait_for_command_registers {
		if _bot.validateWaitForCommandScope(reg, data) {
			if reg.register.Command.Command != nil {
				for _, v := range reg.register.Command.Command {
					if strings.Contains(msg, v) && (!reg.register.Command.RequireAT || strings.Contains(msg, at)) {
						reg.channel <- &data
						if reg.register.Command.IsShortCircuit {
							return true
						}
					}
				}
			}
			if reg.regex == nil && reg.register.Command.Regex != "" {
				reg.regex = regexp.MustCompile(reg.register.Command.Regex)
			}
			if reg.regex != nil {
				fmt.Println(reg.regex, msg, reg.regex.FindString(msg))
				if reg.regex.FindString(msg) != "" && (!reg.register.Command.RequireAT || strings.Contains(msg, at)) {
					reg.channel <- &data
					if reg.register.Command.IsShortCircuit {
						return true
					}
				}
			}
		}
	}
	return false
}

/* public */

// 等待特定指令的触发，并回传触发该指令的消息事件（或超时错误）；
// 用于暂停处理当前消息链，等待特定指令的触发或超时再回复
func (_bot *bot) WaitForCommand(reg models.WaitForCommandRegister) (*events.EventSendMessage, error) {
	// manage default values for optional args
	if reg.Timeout == nil {
		var timeout time.Duration = 1 * time.Minute
		reg.Timeout = &timeout
	}
	if reg.AllowRepeat == nil {
		var allow_repeat bool = true
		reg.AllowRepeat = &allow_repeat
	}

	// create internal register struct
	_reg := waitForCommandRegister{
		register: reg,
		channel:  make(chan *events.EventSendMessage, 1),
		cancel:   make(chan bool, 1),
	}
	defer close(_reg.channel)

	// valid scope with data type and write in cooresponding scope validation data
	if reg.Data != nil && reflect.TypeOf(reg.Data).Kind() == reflect.Ptr {
		reg.Data = reflect.ValueOf(reg.Data).Elem().Interface()
	}
	switch reg.Data.(type) {
	case events.EventJoinVilla:
		if reg.Scope&models.ScopeRoom != 0 {
			return nil, errors.New("JoinVilla事件不可使用Room房间作用域")
		}
		_reg.villa_id = utils.String(reg.Data.(events.EventJoinVilla).Data.VillaId)
		_reg.uid = utils.String(reg.Data.(events.EventJoinVilla).Data.JoinUid)
	case events.EventSendMessage:
		_reg.villa_id = utils.String(reg.Data.(events.EventSendMessage).Data.VillaId)
		_reg.room_id = utils.String(reg.Data.(events.EventSendMessage).Data.RoomId)
		_reg.uid = utils.String(reg.Data.(events.EventSendMessage).Data.FromUserId)
	case events.EventCreateRobot:
		if reg.Scope&models.ScopeRoom != 0 {
			return nil, errors.New("CreateRobot事件不可使用Room房间作用域")
		}
		if reg.Scope&models.ScopeUser != 0 {
			return nil, errors.New("CreateRobot事件不可使用User用户作用域")
		}
		_reg.villa_id = utils.String(reg.Data.(events.EventCreateRobot).Data.VillaId)
	case events.EventDeleteRobot:
		if reg.Scope&models.ScopeRoom != 0 {
			return nil, errors.New("DeleteRobot事件不可使用Room房间作用域")
		}
		if reg.Scope&models.ScopeUser != 0 {
			return nil, errors.New("DeleteRobot事件不可使用User用户作用域")
		}
		_reg.villa_id = utils.String(reg.Data.(events.EventDeleteRobot).Data.VillaId)
	case events.EventAddQuickEmoticon:
		_reg.villa_id = utils.String(reg.Data.(events.EventAddQuickEmoticon).Data.VillaId)
		_reg.room_id = utils.String(reg.Data.(events.EventAddQuickEmoticon).Data.RoomId)
		_reg.uid = utils.String(reg.Data.(events.EventAddQuickEmoticon).Data.Uid)
	case events.EventAuditCallback:
		_reg.villa_id = utils.String(reg.Data.(events.EventAuditCallback).Data.VillaId)
		_reg.room_id = utils.String(reg.Data.(events.EventAuditCallback).Data.RoomId)
		_reg.uid = utils.String(reg.Data.(events.EventAuditCallback).Data.UserId)
	case nil:
		return nil, errors.New("缺少必要的参数：Data")
	default:
		return nil, fmt.Errorf("未知的数据类型: %v", reflect.TypeOf(reg.Data))
	}

	// register to bot
	if !(*reg.AllowRepeat) && reg.Identify != nil {
		for _, v := range _bot.wait_for_command_registers {
			if v.register.Identify != nil && *v.register.Identify == *reg.Identify {
				return nil, errors.New("重复的标识 (AllowRepeat: false)")
			}
		}
	}

	_bot.wait_for_command_registers = append(_bot.wait_for_command_registers, _reg)
	defer func() {
		for i, v := range _bot.wait_for_command_registers {
			if reflect.DeepEqual(v, _reg) {
				_bot.wait_for_command_registers = append(_bot.wait_for_command_registers[:i], _bot.wait_for_command_registers[i+1:]...)
				break
			}
		}
	}()

	// wait for command
	var timer <-chan time.Time
	if *reg.Timeout == 0 { // no timeout if timeout is 0
		timer := make(chan time.Time, 1)
		defer close(timer)
	} else {
		timer = time.After(*reg.Timeout)
	}
	select {
	case res := <-_reg.channel:
		return res, nil
	case <-timer:
		return nil, fmt.Errorf("timeout")
	case <-_reg.cancel:
		return nil, fmt.Errorf("cancel")
	}
}

// 取消等待特定指令的注册
func (_bot *bot) CancelWaitForCommand(identify string) error {
	for i, v := range _bot.wait_for_command_registers {
		if v.register.Identify != nil && *v.register.Identify == identify {
			_bot.wait_for_command_registers[i].cancel <- true
			return nil
		}
	}
	return errors.New("未找到对应的注册")
}
