//go:build wasm
// +build wasm

package evaldo

import (
	"syscall/js"

	// "encoding/json"

	"github.com/refaktor/rye/env"
)

// JavaScript interop functions for Rye WASM
var Builtins_js_interop = map[string]*env.Builtin{

	// Tests:
	// js-call "console.log" { "Hello" "from" "Rye" }
	// js-call "Math.max" { 10 20 5 30 }
	// Args:
	// * function-name: String name of the JavaScript function to call
	// * args: Block of arguments to pass to the function
	// Returns:
	// * result from JavaScript function call
	"js-call": {
		Argsn: 2,
		Doc:   "Calls a JavaScript function with multiple arguments from a block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch funcName := arg0.(type) {
			case env.String:
				jsFunc := js.Global().Get(funcName.Value)
				if jsFunc.IsUndefined() {
					return MakeBuiltinError(ps, "JavaScript function not found: "+funcName.Value, "js-call")
				}

				// Convert arguments block to JavaScript values
				var jsArgs []interface{}
				switch args := arg1.(type) {
				case env.Block:
					series := args.Series.S
					jsArgs = make([]interface{}, len(series))
					for i, arg := range series {
						jsArgs[i] = ryeToJS(arg, ps)
					}
				default:
					// If not a block, treat as single argument
					jsArgs = []interface{}{ryeToJS(arg1, ps)}
				}

				// Call the function with multiple arguments
				result := jsFunc.Invoke(jsArgs...)

				// Convert result back to Rye
				return jsToRye(result, ps)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "js-call")
			}
		},
	},

	// Tests:
	// js-get "https://api.example.com/users"
	// Args:
	// * url: String URL to fetch
	// Returns:
	// * JSON response data as Rye objects
	"___js-get": {
		Argsn: 1,
		Doc:   "Makes HTTP GET request using JavaScript fetch with authentication cookies.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch url := arg0.(type) {
			case env.String:
				jsFunc := js.Global().Get("jsGet")
				if jsFunc.IsUndefined() {
					return MakeBuiltinError(ps, "JavaScript function 'jsGet' not available", "js-get")
				}

				// Use the fixed async approach
				resultId := js.Global().Get("Date").New().Call("getTime").String()

				if js.Global().Get("__ryeAsyncResults").IsUndefined() {
					js.Global().Set("__ryeAsyncResults", js.Global().Get("Object").New())
				}

				// Create wrapper for jsGet
				wrapperCode := `
					(function(url, resultId) {
						try {
							const result = window.jsGet(url);
							if (result && typeof result.then === 'function') {
								result.then(data => {
									window.__ryeAsyncResults[resultId] = { 
										ready: true, 
										success: true, 
										data: data 
									};
								}).catch(error => {
									window.__ryeAsyncResults[resultId] = { 
										ready: true, 
										success: false, 
										error: error.toString() 
									};
								});
							} else {
								window.__ryeAsyncResults[resultId] = { 
									ready: true, 
									success: true, 
									data: result 
								};
							}
						} catch (error) {
							window.__ryeAsyncResults[resultId] = { 
								ready: true, 
								success: false, 
								error: error.toString() 
							};
						}
					})
				`

				wrapper := js.Global().Get("eval").Invoke(wrapperCode)
				wrapper.Invoke(url.Value, resultId)

				// Polling approach
				maxWait := 5000 // 5 seconds for HTTP requests
				startTime := js.Global().Get("Date").New().Call("getTime").Int()

				for {
					result := js.Global().Get("__ryeAsyncResults").Get(resultId)

					if !result.IsUndefined() && result.Get("ready").Bool() {
						js.Global().Get("__ryeAsyncResults").Delete(resultId)

						if result.Get("success").Bool() {
							return jsToRye(result.Get("data"), ps)
						} else {
							return MakeBuiltinError(ps, "HTTP request failed: "+result.Get("error").String(), "js-get")
						}
					}

					currentTime := js.Global().Get("Date").New().Call("getTime").Int()
					if currentTime-startTime > maxWait {
						break
					}
				}

				js.Global().Get("__ryeAsyncResults").Delete(resultId)
				return MakeBuiltinError(ps, "HTTP request timed out", "js-get")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "js-get")
			}
		},
	},

	// Tests:
	// js-post "https://api.example.com/users" { "name" "John" "email" "john@example.com" }
	// Args:
	// * url: String URL to post to
	// * data: Dict or Block of data to send as JSON
	// Returns:
	// * JSON response data as Rye objects
	"__js-post": {
		Argsn: 2,
		Doc:   "Makes HTTP POST request using JavaScript fetch with authentication cookies.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch url := arg0.(type) {
			case env.String:
				jsFunc := js.Global().Get("jsPost")
				if jsFunc.IsUndefined() {
					return MakeBuiltinError(ps, "JavaScript function 'jsPost' not available", "js-post")
				}

				// Convert data to JS value
				jsData := ryeToJS(arg1, ps)

				// Call jsPost function
				promise := jsFunc.Invoke(url.Value, jsData)

				// Handle Promise
				done := make(chan js.Value)
				errChan := make(chan js.Value)

				promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					if len(args) > 0 {
						done <- args[0]
					}
					return nil
				}))

				promise.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					if len(args) > 0 {
						errChan <- args[0]
					}
					return nil
				}))

				select {
				case result := <-done:
					return jsToRye(result, ps)
				case err := <-errChan:
					return MakeBuiltinError(ps, "HTTP POST failed: "+err.String(), "js-post")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "js-post")
			}
		},
	},

	// Tests:
	// js-authenticated-request "https://api.example.com/protected" "GET" void
	// Args:
	// * url: String URL to request
	// * method: String HTTP method (GET, POST, PUT, DELETE)
	// * data: Any data to send (void for GET requests)
	// Returns:
	// * JSON response data as Rye objects
	"js-authenticated-request": {
		Argsn: 3,
		Doc:   "Makes authenticated HTTP request using current user context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch url := arg0.(type) {
			case env.String:
				switch method := arg1.(type) {
				case env.String:
					jsFunc := js.Global().Get("jsAuthenticatedRequest")
					if jsFunc.IsUndefined() {
						return MakeBuiltinError(ps, "JavaScript function 'jsAuthenticatedRequest' not available", "js-authenticated-request")
					}

					// Convert data to JS value
					jsData := ryeToJS(arg2, ps)

					// Call jsAuthenticatedRequest function
					promise := jsFunc.Invoke(url.Value, method.Value, jsData)

					// Handle Promise
					done := make(chan js.Value)
					errChan := make(chan js.Value)

					promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
						if len(args) > 0 {
							done <- args[0]
						}
						return nil
					}))

					promise.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
						if len(args) > 0 {
							errChan <- args[0]
						}
						return nil
					}))

					select {
					case result := <-done:
						return jsToRye(result, ps)
					case err := <-errChan:
						return MakeBuiltinError(ps, "Authenticated request failed: "+err.String(), "js-authenticated-request")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "js-authenticated-request")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "js-authenticated-request")
			}
		},
	},

	// Tests:
	// js-to-csv [ { "name" "John" "age" 30 } { "name" "Jane" "age" 25 } ]
	// Args:
	// * data: Block or Dict to convert to CSV format
	// Returns:
	// * String containing CSV data
	"js-to-csv": {
		Argsn: 1,
		Doc:   "Converts data to CSV format using JavaScript.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			jsFunc := js.Global().Get("jsToCsv")
			if jsFunc.IsUndefined() {
				return MakeBuiltinError(ps, "JavaScript function 'jsToCsv' not available", "js-to-csv")
			}

			// Convert data to JS value
			jsData := ryeToJS(arg0, ps)

			// Call jsToCsv function
			result := jsFunc.Invoke(jsData)

			// Return as string
			return *env.NewString(result.String())
		},
	},

	// Tests:
	// js-to-xml { "users" [ { "name" "John" } { "name" "Jane" } ] }
	// Args:
	// * data: Dict or Block to convert to XML format
	// Returns:
	// * String containing XML data
	"js-to-xml": {
		Argsn: 1,
		Doc:   "Converts data to XML format using JavaScript.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			jsFunc := js.Global().Get("jsToXml")
			if jsFunc.IsUndefined() {
				return MakeBuiltinError(ps, "JavaScript function 'jsToXml' not available", "js-to-xml")
			}

			// Convert data to JS value
			jsData := ryeToJS(arg0, ps)

			// Call jsToXml function
			result := jsFunc.Invoke(jsData)

			// Return as string
			return *env.NewString(result.String())
		},
	},

	// Tests:
	// js-download "name,age\nJohn,30\nJane,25" "users.csv"
	// Args:
	// * content: String content to download
	// * filename: String name for the downloaded file
	// Returns:
	// * String confirmation message
	"js-download": {
		Argsn: 2,
		Doc:   "Downloads content as a file using JavaScript.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch content := arg0.(type) {
			case env.String:
				switch filename := arg1.(type) {
				case env.String:
					jsFunc := js.Global().Get("jsDownload")
					if jsFunc.IsUndefined() {
						return MakeBuiltinError(ps, "JavaScript function 'jsDownload' not available", "js-download")
					}

					// Call jsDownload function
					result := jsFunc.Invoke(content.Value, filename.Value)

					// Return confirmation message
					return *env.NewString(result.String())
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "js-download")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "js-download")
			}
		},
	},

	// Tests:
	// js-log "Debug message from Rye"
	// Args:
	// * message: String message to log to browser console
	// Returns:
	// * the message string
	"js-log": {
		Argsn: 1,
		Doc:   "Logs a message to the browser console.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch message := arg0.(type) {
			case env.String:
				js.Global().Get("console").Call("log", "Rye: "+message.Value)
				return message
			default:
				// Convert other types to string
				str := arg0.Inspect(*ps.Idx)
				js.Global().Get("console").Call("log", "Rye: "+str)
				return *env.NewString(str)
			}
		},
	},

	// Tests:
	// js-current-time
	// Args:
	// Returns:
	// * String containing current ISO timestamp
	"js-current-time": {
		Argsn: 0,
		Doc:   "Gets current timestamp from JavaScript Date.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			jsFunc := js.Global().Get("jsCurrentTime")
			if jsFunc.IsUndefined() {
				// Fallback to direct JS call
				result := js.Global().Get("Date").New().Call("toISOString")
				return *env.NewString(result.String())
			}

			result := jsFunc.Invoke()
			return *env.NewString(result.String())
		},
	},

	// Tests:
	// js-alert "Hello from Rye!"
	// Args:
	// * message: String message to display in browser alert dialog
	// Returns:
	// * the message string
	"js-alert": {
		Argsn: 1,
		Doc:   "Shows a browser alert dialog with the given message.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch message := arg0.(type) {
			case env.String:
				js.Global().Call("alert", message.Value)
				return message
			default:
				// Convert other types to string
				str := arg0.Inspect(*ps.Idx)
				js.Global().Call("alert", str)
				return *env.NewString(str)
			}
		},
	},

	// Tests:
	// js-prompt "Enter your name:"
	// js-prompt "Enter your age:" "25"
	// Args:
	// * message: String message to display in the prompt dialog
	// * default-value: Optional string default value for the input
	// Returns:
	// * String containing the user input, or void if cancelled
	"js-prompt": {
		Argsn: -1, // Variable arguments (1 or 2)
		Doc:   "Shows a browser prompt dialog and returns user input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if arg0 is nil (no arguments)
			if arg0 == nil {
				return MakeBuiltinError(ps, "js-prompt expects 1 or 2 arguments", "js-prompt")
			}

			switch message := arg0.(type) {
			case env.String:
				var result js.Value
				// Check if default value is provided (arg1 is not nil)
				if arg1 != nil {
					switch defaultVal := arg1.(type) {
					case env.String:
						result = js.Global().Call("prompt", message.Value, defaultVal.Value)
					default:
						// Convert other types to string for default value
						defaultStr := arg1.Inspect(*ps.Idx)
						result = js.Global().Call("prompt", message.Value, defaultStr)
					}
				} else {
					result = js.Global().Call("prompt", message.Value)
				}

				// Check if user cancelled (returns null)
				if result.IsNull() {
					return env.NewVoid()
				}
				return *env.NewString(result.String())
			default:
				// Convert other types to string for message
				str := arg0.Inspect(*ps.Idx)
				var result js.Value
				if arg1 != nil {
					defaultStr := arg1.Inspect(*ps.Idx)
					result = js.Global().Call("prompt", str, defaultStr)
				} else {
					result = js.Global().Call("prompt", str)
				}
				if result.IsNull() {
					return env.NewVoid()
				}
				return *env.NewString(result.String())
			}
		},
	},

	"js-call-callback": {
		Argsn: 3,
		Doc: "Calls a JS async function and executes a Rye callback with the result. " +
			"Callback receives the result or an error string prefixed with 'Error: '.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			funcName, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "js-call-callback")
			}
			callback, ok := arg2.(env.Function)
			if !ok {
				return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "js-call-callback")
			}

			// Build JS args
			var jsArgs []interface{}
			if block, ok := arg1.(env.Block); ok {
				series := block.Series.S
				jsArgs = make([]interface{}, len(series))
				for i, a := range series {
					jsArgs[i] = ryeToJS(a, ps)
				}
			} else {
				jsArgs = []interface{}{ryeToJS(arg1, ps)}
			}

			// js.FuncOf callbacks MUST be released after use to avoid leaks.
			// We use a shared release func so each callback releases both.
			var successCb, errorCb js.Func

			successCb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				defer successCb.Release()
				defer errorCb.Release()

				var result env.Object
				if len(args) > 0 {
					result = jsToRye(args[0], ps)
				} else {
					result = env.NewVoid()
				}
				CallFunctionWithArgs(callback, ps, ps.Ctx, result)
				return nil
			})

			errorCb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				defer successCb.Release()
				defer errorCb.Release()

				msg := "Unknown error"
				if len(args) > 0 {
					msg = args[0].String()
				}
				CallFunctionWithArgs(callback, ps, ps.Ctx, env.String{Value: "Error: " + msg})
				return nil
			})

			// Inline JS avoids eval; just use the already-retrieved jsFunc.
			// But we still need to handle the sync-vs-Promise duality in JS.
			handlerSrc := `(function(fn, args, onSuccess, onError) {
            try {
                var r = fn.apply(null, args);
                if (r && typeof r.then === 'function') {
                    r.then(onSuccess).catch(onError);
                } else {
                    onSuccess(r);
                }
            } catch(e) {
                onError(e.toString());
            }
        })`
			handler := js.Global().Get("eval").Invoke(handlerSrc)

			jsFunc := js.Global().Get(funcName.Value)
			if jsFunc.IsUndefined() {
				successCb.Release()
				errorCb.Release()
				return MakeBuiltinError(ps, "JS function not found: "+funcName.Value, "js-call-callback")
			}

			handler.Invoke(jsFunc, jsArgs, successCb, errorCb)
			return *env.NewString("async-initiated")
		},
	},
}

// Helper function to convert Rye values to JavaScript values
func ryeToJS(obj env.Object, ps *env.ProgramState) js.Value {
	switch val := obj.(type) {
	case env.String:
		return js.ValueOf(val.Value)
	case env.Integer:
		return js.ValueOf(float64(val.Value))
	case env.Decimal:
		return js.ValueOf(val.Value)
	case env.Dict:
		// Convert dictionary to JavaScript object
		jsObj := js.Global().Get("Object").New()
		for key, value := range val.Data {
			jsObj.Set(key, ryeToJS(value.(env.Object), ps))
		}
		return jsObj
	case env.Block:
		// Convert block to JavaScript array
		series := val.Series.S
		jsArray := js.Global().Get("Array").New(len(series))
		for i, item := range series {
			jsArray.SetIndex(i, ryeToJS(item, ps))
		}
		return jsArray
	case env.Void:
		return js.Null()
	default:
		// For unknown types, convert to string
		return js.ValueOf(val.Inspect(*ps.Idx))
	}
}

// Helper function to convert JavaScript values to Rye values
func jsToRye(jsVal js.Value, ps *env.ProgramState) env.Object {
	switch jsVal.Type() {
	case js.TypeString:
		return *env.NewString(jsVal.String())
	case js.TypeNumber:
		num := jsVal.Float()
		// Check if it's an integer
		if num == float64(int64(num)) {
			return *env.NewInteger(int64(num))
		}
		return *env.NewDecimal(num)
	case js.TypeBoolean:
		if jsVal.Bool() {
			return *env.NewInteger(1)
		}
		return *env.NewInteger(0)
	case js.TypeObject:
		if jsVal.IsNull() {
			return env.NewVoid()
		}

		// Check if it's an array
		if jsVal.Get("length").Type() == js.TypeNumber {
			// It's an array
			length := jsVal.Get("length").Int()
			series := make([]env.Object, length)
			for i := 0; i < length; i++ {
				series[i] = jsToRye(jsVal.Index(i), ps)
			}
			return *env.NewBlock(*env.NewTSeries(series))
		} else {
			// It's an object - convert to dictionary
			data := make(map[string]any)

			// Get object keys using JavaScript Object.keys()
			keys := js.Global().Get("Object").Call("keys", jsVal)
			keysLength := keys.Get("length").Int()

			for i := 0; i < keysLength; i++ {
				key := keys.Index(i).String()
				value := jsToRye(jsVal.Get(key), ps)
				data[key] = value
			}

			return *env.NewDict(data)
		}
	case js.TypeUndefined:
		return env.NewVoid()
	default:
		// For unknown types, convert to string
		return *env.NewString(jsVal.String())
	}
}
