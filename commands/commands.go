package commands

import (
	"regexp"
	"strings"

	api_models "github.com/GLGDLY/mhy_botsdk/api_models"
	apis "github.com/GLGDLY/mhy_botsdk/apis"
	events "github.com/GLGDLY/mhy_botsdk/events"
	logger "github.com/GLGDLY/mhy_botsdk/logger"
	utils "github.com/GLGDLY/mhy_botsdk/utils"
)

type Preprocessor events.BotListenerSendMessage

type OnCommand struct {
	Command        []string       // 可触发事件的指令列表，与正则 Regex 互斥，优先使用此项
	Regex          string         // 可触发指令的正则表达式，与指令表 Command 互斥
	regex          *regexp.Regexp // internal use
	Listener       events.BotListenerSendMessage
	RequireAT      bool   // 是否要求必须@机器人才能触发指令
	RequireAdmin   bool   // 是否要求频道主或或管理才可触发指令
	AdminErrorMsg  string // 当RequireAT，而触发用户的权限不足时，如此项不为空，返回此消息并短路；否则不进行短路
	IsShortCircuit bool   // 如果触发指令成功是否短路不运行后续指令（将根据注册顺序排序指令的短路机制）
}

func (p *OnCommand) processCommand(data events.EventSendMessage, _logger logger.LoggerInterface, _api *apis.ApiBase) bool {
	at := "@" + data.Robot.Template.Name
	if p.RequireAT && !strings.Contains(data.GetContent(false), at) {
		return false
	}
	if p.RequireAdmin {
		res, http_code, err := _api.GetMember(data.Robot.VillaId, data.Data.FromUserId)
		if err != nil {
			_logger.Error("command listener {", utils.GetFunctionName(p.Listener), "} get member role info error: ", err)
			return false
		} else if http_code != 200 || res.Retcode != 0 {
			_logger.Error("command listener {", utils.GetFunctionName(p.Listener), "} get member role info error: ", res)
			return false
		}
		is_admin := false
		for _, v := range res.Data.Member.RoleList {
			if v.RoleType == "MEMBER_ROLE_TYPE_ADMIN" || v.RoleType == "MEMBER_ROLE_TYPE_OWNER" {
				is_admin = true
				break
			}
		}
		if !is_admin {
			if p.AdminErrorMsg != "" {
				msg, _ := api_models.NewMsg(api_models.MsgTypeText)
				err := msg.SetText(p.AdminErrorMsg)
				if err != nil {
					_logger.Error("command listener {", utils.GetFunctionName(p.Listener), "} error: ", err)
					return false
				}
				_, http, err := _api.SendMessageCustomize(data.Robot.VillaId, data.Data.RoomId, msg)
				if err != nil || http != 200 {
					_logger.Error("command listener {", utils.GetFunctionName(p.Listener), "} error on sending admin error msg: ", err, "(", http, ")")
					return false
				}
				return true
			}
			return false
		}
	}
	utils.Try(func() { p.Listener(data) }, func(err interface{}, tb string) {
		_logger.Error("command listener {", utils.GetFunctionName(p.Listener), "} error: ", err, "\n", tb)
	})
	return p.IsShortCircuit
}

// 内部检查当前消息是否符合触发条件
func (p *OnCommand) CheckCommand(data events.EventSendMessage, _logger logger.LoggerInterface, _api *apis.ApiBase) bool {
	msg := data.GetContent(false)
	if p.Command != nil {
		for _, v := range p.Command {
			if strings.Contains(msg, v) {
				if p.processCommand(data, _logger, _api) {
					return true
				}
			}
		}
	}
	if p.regex == nil && p.Regex != "" {
		p.regex = regexp.MustCompile(p.Regex)
	}
	if p.regex != nil {
		if p.regex.FindString(msg) != "" {
			if p.processCommand(data, _logger, _api) {
				return true
			}
		}
	}
	return false
}

func (p *OnCommand) Equals(_p OnCommand) bool {
	if p.Command != nil && _p.Command != nil {
		if len(p.Command) != len(_p.Command) {
			return false
		}
		p_map := make(map[string]bool)
		for _, v := range p.Command {
			p_map[v] = true
		}
		for _, v := range _p.Command {
			if _, ok := p_map[v]; !ok {
				return false
			}
		}
		return true
	}
	if p.Regex != "" && _p.Regex != "" {
		return p.Regex == _p.Regex
	}
	return false
}
