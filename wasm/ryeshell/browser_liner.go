package main

import (
	"fmt"
	"syscall/js"
)

var (
	jsCallback js.Value
)

func sendMessageToJS(message string) {
	jsCallback.Invoke(message)
}

func main() {
	c := make(chan string)

	js.Global().Set("sendKeypress", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			c <- args[0].String()
		}
		return nil
	}))

	// Get the JavaScript function to call back
	jsCallback = js.Global().Get("receiveMessageFromGo")

	for {
		key := <-c
		// Process the keypress and then send a message back to JavaScript
		response := key
		if key == "A" {
			response = "\x1B[1;3;31mA\x1B[0m"
		}
		sendMessageToJS(response)
	}
}

func main33() {
	c := make(chan struct{}, 0)
	js.Global().Set("RyeEvalString", js.FuncOf(RyeEvalString))
	<-c
}

func RyeEvalString(this js.Value, args []js.Value) any {
	fmt.Println(args)

	//util.PrintHeader()
	//defer profile.Start(profile.CPUProfile).Stop()

	return "Other"
}
