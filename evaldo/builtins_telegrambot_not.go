// +build !b_telegram

package evaldo

import (
	"rye/env"
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

var Builtins_telegrambot = map[string]*env.Builtin{}
