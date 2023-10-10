//go:build b_telegram
// +build b_telegram

package evaldo

import (
	"rye/env"

	tgm "github.com/go-telegram-bot-api/telegram-bot-api"
)

/*

	new-telegram-bot "MyRyeBot" :bot
	|on-updates-do {
		-> 'Message :msg |if {
			print msg -> 'From -> 'UserName
			print concat " " msg -> Text

			new-telegram-message msg -> 'Chat -> 'ID  msg -> 'Text
			|send* bot
		}
	}

	new-telegram-bot 1
	telegra-bot//on-updates-do
	new-telegram-message
	telegram-bot//send

*/

func TelegramUpdateToRyeDict(update_ tgm.Update) env.Dict {
	update := make(map[string]interface{})
	if update_.Message != nil {
		msg := make(map[string]interface{})
		msg["Text"] = update_.Message.Text
		msg["MessageID"] = update_.Message.MessageID
		if update_.Message.From != nil {
			from := make(map[string]interface{})
			from["UserName"] = update_.Message.From.UserName
			from["FirstName"] = update_.Message.From.FirstName
			from["ID"] = update_.Message.From.ID
			from["LastName"] = update_.Message.From.LastName
			from["IsBot"] = update_.Message.From.IsBot

			msg["From"] = *env.NewDict(from)
		}
		if update_.Message.Chat != nil {
			chat := make(map[string]interface{})
			chat["ID"] = update_.Message.Chat.ID
			msg["Chat"] = *env.NewDict(chat)
		}
		update["Message"] = *env.NewDict(msg)
	}
	return *env.NewDict(update)
}

var Builtins_telegrambot = map[string]*env.Builtin{

	"new-telegram-bot": {
		Argsn: 1,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.String:
				bot, err := tgm.NewBotAPI(arg.Value)
				if err != nil {
					return makeError(ps, "Arg 1 should be Integer.")
				}
				return *env.NewNative(ps.Idx, bot, "telegram-bot")
			default:
				return makeError(ps, "Arg 1 should be Integer.")
			}
		},
	},

	"telegram-bot//on-update": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bot := arg0.(type) {
			case env.Native:
				switch code := arg1.(type) {
				case env.Block:
					u := tgm.NewUpdate(0)
					u.Timeout = 60

					updates, _ := bot.Value.(*tgm.BotAPI).GetUpdatesChan(u)
					ser := ps.Ser
					ps.Ser = code.Series

					for update := range updates {
						dict := TelegramUpdateToRyeDict(update)
						// fmt.Println(dict.Probe(*ps.Idx))
						ps = EvalBlockInj(ps, dict, true)
						if ps.ErrorFlag {
							return ps.Res
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return env.Integer{1}
				default:
					return makeError(ps, "Arg 1 should be Native.")
				}
			default:
				return makeError(ps, "Arg 2 should be Block.")
			}
		},
	},

	"telegram-message//send": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch msg := arg0.(type) {
			case env.Native:
				switch bot := arg1.(type) {
				case env.Native:
					bot.Value.(*tgm.BotAPI).Send(msg.Value.(tgm.MessageConfig))
					return arg0
				default:
					return makeError(ps, "Arg 1 should be Native.")
				}
			default:
				return makeError(ps, "Arg 1 should be Native.")
			}

		},
	},
	"new-telegram-message": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cid := arg0.(type) {
			case env.Integer:
				switch txt := arg1.(type) {
				case env.String:
					msg := tgm.NewMessage(cid.Value, txt.Value)
					return *env.NewNative(ps.Idx, msg, "telegram-message")
				default:
					return makeError(ps, "Arg 2 should String.")
				}
			default:
				return makeError(ps, "Arg 1 should be Integer.")
			}
		},
	},
}
