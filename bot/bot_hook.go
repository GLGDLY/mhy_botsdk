package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	events "github.com/GLGDLY/mhy_botsdk/events"
	utils "github.com/GLGDLY/mhy_botsdk/utils"
	"github.com/gin-gonic/gin"
)

func processEvent(_bot *Bot, event events.Event) {
	event_id := event.Event.Id
	if _bot.filter_manager.needFilter(event_id) {
		_bot.Logger.Debugf("filter repeat event: %v+\n", event)
		return
	} else {
		_bot.filter_manager.add(event_id)
	}

	if _bot.use_default_logger {
		_bot.Logger.Debugf("receive event: %v+\n", event)
	}
	event_type := event.Event.Type
switch_label:
	switch event_type {
	case events.JoinVilla:
		event := events.Event2EventJoinVilla(event)
		for _, listener := range _bot.listeners_join_villa {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	case events.SendMessage:
		event := events.Event2EventSendMessage(event, _bot.Api)
		if _bot.is_filter_self_msg && event.Data.Content.User.Id == _bot.Base.ID {
			break switch_label
		}
		// 1. run preprocessors
		for _, _preprocessor := range _bot.preprocessors {
			utils.Try(func() { _preprocessor(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("preprocessor {", utils.GetFunctionName(_preprocessor), "} error: ", err, "\n", tb)
			})
		}
		// 2. run wait_for command registers
		if _bot.checkWaifForCommand(event) {
			break switch_label
		}
		// 3. run plugins

		// 3_1. run plugins preprocessors
		for _, p := range _bot.plugins {
			if p.IsEnable {
				for _, _preprocessor := range p.Preprocessors {
					utils.Try(func() { _preprocessor(event, _bot.abstract_bot) }, func(err interface{}, tb string) {
						_bot.Logger.Error("preprocessor {", utils.GetFunctionName(_preprocessor), "} error: ", err, "\n", tb)
					})
				}
			}
		}

		// 3_2. run plugins commands
		for _, p := range _bot.plugins {
			if p.IsEnable {
				_is_short_circuit := false
				for _, _command := range p.OnCommand {
					if _command.CheckCommand(event, _bot.abstract_bot) {
						_is_short_circuit = true
						break // short circuit for plugin's internal commands
					}
				}
				if _is_short_circuit && _bot.is_plugins_short_circuit_affect_main {
					break switch_label // short circuit for all commands
				}
			}
		}

		// 4. run on commands
		for _, _command := range _bot.on_commands {
			if _command.CheckCommand(event, _bot.Logger, _bot.Api) {
				break switch_label // short circuit
			}
		}
		// 5. run normal listeners
		for _, listener := range _bot.listeners_send_message {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	case events.CreateRobot:
		event := events.Event2EventCreateRobot(event)
		for _, listener := range _bot.listeners_create_robot {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	case events.DeleteRobot:
		event := events.Event2EventDeleteRobot(event)
		for _, listener := range _bot.listeners_delete_robot {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	case events.AddQuickEmoticon:
		event := events.Event2EventAddQuickEmoticon(event)
		for _, listener := range _bot.listeners_add_quick_emoticon {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	case events.AuditCallback:
		event := events.Event2EventAuditCallback(event)
		for _, listener := range _bot.listeners_audit_callback {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	default:
		_bot.Logger.Warnf("unknown event type: %v\n", event_type)
	}
}

func hook(_bots []*Bot, c *gin.Context) {
	// send raw request to listeners on bot of current path and port
	for _, _bot := range _bots {
		if !_bot.is_running {
			continue
		}
		for _, listener := range _bot.listeners_raw_request {
			go utils.Try(func() { listener(c) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	}

	// handle event
	if c.Request.Method != "POST" || c.Request.Header.Get("Content-Type") != "application/json" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "bad request",
			"retcode": -1,
		})
		return
	}

	defer func() { // ensure response
		c.JSON(http.StatusOK, gin.H{
			"message": "",
			"retcode": 0,
		})
	}()

	// read body content
	raw_body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println("new event read body error: ", err)
		return
	}

	// decode event
	var event events.Event
	err = json.Unmarshal(raw_body, &event)
	if err != nil {
		fmt.Println("decode event error (" + err.Error() + "): " + string(raw_body))
		return
	}
	if event.Type == "hb" { // handle heartbeat packet if using websocket protocol(based on reverse proxy)
		return
	}

	// find bot ctx by id to allow multiple bot running on same port &|| path
	_id := event.Event.Robot.Template.Id
	_bot_ctx, ok := bot_context_manager[_id] // find bot ctx by id to allow multiple bot running on same port &|| path
	if !ok {
		// fmt.Println("bot id: ", _id, " not found in ctx: ", bot_context_manager)
		return
	}
	_bot := _bot_ctx.bot

	// check if bot is running
	if !_bot.is_running {
		return
	}

	// handle reverse proxy
	for _, http_proxy := range _bot.reverse_proxy_http_msg_chan {
		http_proxy <- raw_body
	}
	for _, ws_proxy := range _bot.reverse_proxy_ws_msg_chan {
		ws_proxy <- raw_body
	}

	// detailed event processing
	if _bot.is_verify_msg_signature {
		sign := c.Request.Header.Get("x-rpc-bot_sign")
		verify, err := pubKeyVerify(sign, string(raw_body), _bot.Base.Secret, _bot.Base.PubKey)
		if (!verify) || (err != nil) {
			_bot.Logger.Debug("new event verify error, rejected: ", err)
			return
		}
	}
	go processEvent(_bot, event) // use goroutine to avoid blocking (especially handle wait_for)

}
