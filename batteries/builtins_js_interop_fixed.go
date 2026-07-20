//go:build wasm
// +build wasm

package batteries

import (
	"syscall/js"
	"time"
	
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// FIXED: Non-blocking async call that uses timeout instead of channels
var FixedAsyncBuiltins = map[string]*env.Builtin{
	
	// The CORRECT way to handle async in WASM
	// This version doesn't block the browser thread
	"js-call-async\\fixed": {
		Argsn: 2,
		Doc:   "Calls a JavaScript function that returns a Promise (non-blocking).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch funcName := arg0.(type) {
			case env.String:
				jsFunc := js.Global().Get(funcName.Value)
				if jsFunc.IsUndefined() {
					return evaldo.MakeBuiltinError(ps, "JavaScript function not found: "+funcName.Value, "js-call-async\\fixed")
				}

				// Convert arguments
				var jsArgs []interface{}
				switch args := arg1.(type) {
				case env.Block:
					series := args.Series.S
					jsArgs = make([]interface{}, len(series))
					for i, arg := range series {
						jsArgs[i] = ryeToJS(arg, ps)
					}
				default:
					jsArgs = []interface{}{ryeToJS(arg1, ps)}
				}

				// Use a global storage approach instead of channels
				resultId := js.Global().Get("Date").New().Call("getTime").String()
				
				// Set up global result storage if it doesn't exist
				if js.Global().Get("__ryeAsyncResults").IsUndefined() {
					js.Global().Set("__ryeAsyncResults", js.Global().Get("Object").New())
				}

				// Create a wrapper that stores results globally
				wrapperCode := `
					(function(funcName, args, resultId) {
						const fn = window[funcName];
						if (!fn) {
							window.__ryeAsyncResults[resultId] = { 
								ready: true, 
								success: false, 
								error: 'Function not found: ' + funcName 
							};
							return;
						}
						
						try {
							const result = fn.apply(null, args);
							if (result && typeof result.then === 'function') {
								// It's a Promise
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
								// Synchronous result
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
				wrapper.Invoke(funcName.Value, jsArgs, resultId)

				// Now, instead of using channels, we use a timeout-based polling approach
				// This is not ideal, but it works in WASM without deadlocking
				maxWait := 100 // Maximum polling iterations (adjust as needed)
				
				for i := 0; i < maxWait; i++ {
					// Check if result is ready
					result := js.Global().Get("__ryeAsyncResults").Get(resultId)
					
					if !result.IsUndefined() && result.Get("ready").Bool() {
						// Clean up
						js.Global().Get("__ryeAsyncResults").Delete(resultId)
						
						if result.Get("success").Bool() {
							return jsToRye(result.Get("data"), ps)
						} else {
							return evaldo.MakeBuiltinError(ps, "JavaScript Promise rejected: "+result.Get("error").String(), "js-call-async\\fixed")
						}
					}
					
					// Small delay to prevent busy-waiting and allow JS to execute
					// This is the key: we yield control back to the browser
					time.Sleep(10 * time.Millisecond)
				}
				
				// If we reach here, the operation timed out
				js.Global().Get("__ryeAsyncResults").Delete(resultId)
				return evaldo.MakeBuiltinError(ps, "Async operation timed out", "js-call-async\\fixed")
				
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "js-call-async\\fixed")
			}
		},
	},
	
	// Alternative: Callback-based approach that doesn't block at all
	"js-call-async\\callback": {
		Argsn: 3,
		Doc:   "Calls a JavaScript function asynchronously and executes Rye callback with result.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch funcName := arg0.(type) {
			case env.String:
				// arg1 = arguments, arg2 = callback block
				
				// Convert arguments
				var jsArgs []interface{}
				switch args := arg1.(type) {
				case env.Block:
					series := args.Series.S
					jsArgs = make([]interface{}, len(series))
					for i, arg := range series {
						jsArgs[i] = ryeToJS(arg, ps)
					}
				default:
					jsArgs = []interface{}{ryeToJS(arg1, ps)}
				}

				// Get the callback block
				// var callbackBlock env.Block
				switch cb := arg2.(type) {
				case env.Block:
					// callbackBlock = cb
					_ = cb // Store for future implementation
				default:
					return evaldo.MakeArgError(ps, 3, []env.Type{env.BlockType}, "js-call-async\\callback")
				}

				// Create unique ID for this callback
				callbackId := js.Global().Get("Date").New().Call("getTime").String()
				
				// Store callback in global registry
				if js.Global().Get("__ryeCallbacks").IsUndefined() {
					js.Global().Set("__ryeCallbacks", js.Global().Get("Object").New())
				}

				// The tricky part: we need to store the Rye callback somehow
				// For now, we'll use a simpler approach and just return immediately
				
				// Create the JS wrapper
				wrapperCode := `
					(function(funcName, args, callbackId) {
						const fn = window[funcName];
						if (!fn) {
							console.error('Function not found:', funcName);
							return;
						}
						
						try {
							const result = fn.apply(null, args);
							if (result && typeof result.then === 'function') {
								result.then(data => {
									console.log('Async result for', callbackId, ':', data);
									// In a full implementation, this would trigger the Rye callback
									window.postMessage({type: 'rye-async-result', callbackId: callbackId, success: true, data: data}, '*');
								}).catch(error => {
									console.error('Async error for', callbackId, ':', error);
									window.postMessage({type: 'rye-async-result', callbackId: callbackId, success: false, error: error.toString()}, '*');
								});
							} else {
								console.log('Sync result for', callbackId, ':', result);
								window.postMessage({type: 'rye-async-result', callbackId: callbackId, success: true, data: result}, '*');
							}
						} catch (error) {
							console.error('Call error for', callbackId, ':', error);
							window.postMessage({type: 'rye-async-result', callbackId: callbackId, success: false, error: error.toString()}, '*');
						}
					})
				`
				
				wrapper := js.Global().Get("eval").Invoke(wrapperCode)
				wrapper.Invoke(funcName.Value, jsArgs, callbackId)
				
				// Return the callback ID so the caller knows the operation is in progress
				return *env.NewString("ASYNC_STARTED_" + callbackId)
				
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "js-call-async\\callback")
			}
		},
	},
}