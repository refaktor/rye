//go:build !no_mqtt
// +build !no_mqtt

package evaldo

import (
	"fmt"
	"strings"
	"time"

	"github.com/refaktor/rye/env"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var Builtins_mqtt = map[string]*env.Builtin{

	//
	// ##### MQTT Client Functions #####
	//

	// Tests:
	// example { Open mqtt://localhost:1883/my-client-id }
	// example { Open mqtt://user:pass@localhost:1883/my-client-id }
	// Args:
	// * uri: MQTT broker URI (format: mqtt://[user:pass@]host:port/client-id)
	// Returns:
	// * native MQTT client connection (type: "Rye-mqtt-client")
	// * error if connection fails
	"mqtt-schema//Open": {
		Argsn: 1,
		Doc:   "Opens a connection to an MQTT broker.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch uri := arg0.(type) {
			case env.Uri:
				// Parse the URI to extract connection details
				// Format: mqtt://[username:password@]host:port/client-id
				scheme := ps.Idx.GetWord(uri.Scheme.Index)
				if scheme != "mqtt" && scheme != "mqtts" {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "URI scheme must be 'mqtt' or 'mqtts'", "mqtt-schema//Open")
				}

				// Construct broker URL
				var brokerURL string
				if scheme == "mqtts" {
					brokerURL = "ssl://" + uri.Path
				} else {
					brokerURL = "tcp://" + uri.Path
				}

				// Extract client ID from path (after the host:port)
				// For now, use a default client ID if not provided in path
				clientID := "rye-mqtt-client"
				if len(uri.Path) > 0 {
					// If there's a slash in the path after host:port, use that as client ID
					parts := strings.Split(uri.Path, "/")
					if len(parts) > 1 && parts[1] != "" {
						clientID = parts[1]
						// Remove client ID from broker URL
						brokerURL = strings.Replace(brokerURL, "/"+clientID, "", 1)
					}
				}

				opts := mqtt.NewClientOptions()
				opts.AddBroker(brokerURL)
				opts.SetClientID(clientID)
				opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
					// Default message handler
				})

				// TODO: Extract username/password from URI if present
				// This would require parsing uri.Path more thoroughly

				client := mqtt.NewClient(opts)
				if token := client.Connect(); token.Wait() && token.Error() != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, fmt.Sprintf("Failed to connect to MQTT broker: %v", token.Error()), "mqtt-schema//Open")
				}

				return *env.NewNative(ps.Idx, client, "Rye-mqtt-client")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "mqtt-schema//Open")
			}
		},
	},

	// Tests:
	// example { client |Disconnect }
	// Args:
	// * client: MQTT client connection (type: "Rye-mqtt-client")
	// Returns:
	// * integer 1 for success
	// * error if disconnection fails
	"Rye-mqtt-client//Disconnect": {
		Argsn: 1,
		Doc:   "Disconnects from the MQTT broker.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				mqttClient := client.Value.(mqtt.Client)
				mqttClient.Disconnect(250) // 250ms timeout
				return *env.NewInteger(1)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mqtt-client//Disconnect")
			}
		},
	},

	// Tests:
	// example { client |Publish "sensors/temperature" "23.5" 0 false }
	// Args:
	// * client: MQTT client connection (type: "Rye-mqtt-client")
	// * topic: Topic to publish to (string)
	// * payload: Message payload (string)
	// * qos: Quality of Service level (integer 0, 1, or 2)
	// * retain: Whether message should be retained (boolean)
	// Returns:
	// * integer 1 for success
	// * error if publish fails
	"Rye-mqtt-client//Publish": {
		Argsn: 5,
		Doc:   "Publishes a message to an MQTT topic with specified QoS and retain flag.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch topic := arg1.(type) {
				case env.String:
					switch payload := arg2.(type) {
					case env.String:
						switch qos := arg3.(type) {
						case env.Integer:
							switch retain := arg4.(type) {
							case env.Integer: // Using integer for boolean (0 = false, 1 = true)
								mqttClient := client.Value.(mqtt.Client)
								retainBool := retain.Value != 0

								if qos.Value < 0 || qos.Value > 2 {
									ps.FailureFlag = true
									return MakeBuiltinError(ps, "QoS must be 0, 1, or 2", "Rye-mqtt-client//Publish")
								}

								token := mqttClient.Publish(topic.Value, byte(qos.Value), retainBool, payload.Value)
								if token.Wait() && token.Error() != nil {
									ps.FailureFlag = true
									return MakeBuiltinError(ps, fmt.Sprintf("Failed to publish message: %v", token.Error()), "Rye-mqtt-client//Publish")
								}

								return *env.NewInteger(1)
							default:
								ps.FailureFlag = true
								return MakeArgError(ps, 5, []env.Type{env.IntegerType}, "Rye-mqtt-client//Publish")
							}
						default:
							ps.FailureFlag = true
							return MakeArgError(ps, 4, []env.Type{env.IntegerType}, "Rye-mqtt-client//Publish")
						}
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Rye-mqtt-client//Publish")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-mqtt-client//Publish")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mqtt-client//Publish")
			}
		},
	},

	// Tests:
	// example { client |Publish-simple "sensors/temperature" "23.5" }
	// Args:
	// * client: MQTT client connection (type: "Rye-mqtt-client")
	// * topic: Topic to publish to (string)
	// * payload: Message payload (string)
	// Returns:
	// * integer 1 for success (uses QoS 0, no retain)
	// * error if publish fails
	"Rye-mqtt-client//Publish-simple": {
		Argsn: 3,
		Doc:   "Publishes a message to an MQTT topic with default settings (QoS 0, no retain).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch topic := arg1.(type) {
				case env.String:
					switch payload := arg2.(type) {
					case env.String:
						mqttClient := client.Value.(mqtt.Client)

						token := mqttClient.Publish(topic.Value, 0, false, payload.Value)
						if token.Wait() && token.Error() != nil {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, fmt.Sprintf("Failed to publish message: %v", token.Error()), "Rye-mqtt-client//Publish-simple")
						}

						return *env.NewInteger(1)
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Rye-mqtt-client//Publish-simple")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-mqtt-client//Publish-simple")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mqtt-client//Publish-simple")
			}
		},
	},

	// Tests:
	// example { client |Subscribe "sensors/+" 1 fn { msg } { print "Received: " + msg } }
	// Args:
	// * client: MQTT client connection (type: "Rye-mqtt-client")
	// * topic: Topic pattern to subscribe to (string, can include wildcards)
	// * qos: Quality of Service level (integer 0, 1, or 2)
	// * handler: Callback function to handle received messages
	// Returns:
	// * integer 1 for success
	// * error if subscription fails
	"Rye-mqtt-client//Subscribe": {
		Argsn: 4,
		Doc:   "Subscribes to an MQTT topic with a message handler function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch topic := arg1.(type) {
				case env.String:
					switch qos := arg2.(type) {
					case env.Integer:
						switch handler := arg3.(type) {
						case env.Function:
							mqttClient := client.Value.(mqtt.Client)

							if qos.Value < 0 || qos.Value > 2 {
								ps.FailureFlag = true
								return MakeBuiltinError(ps, "QoS must be 0, 1, or 2", "Rye-mqtt-client//Subscribe")
							}

							callback := func(client mqtt.Client, msg mqtt.Message) {
								// Create a new program state copy for the callback
								psCallback := *ps
								psCallback.FailureFlag = false
								psCallback.ErrorFlag = false
								psCallback.ReturnFlag = false

								// Create message object as a Dict containing topic and payload
								msgDict := make(map[string]any)
								msgDict["topic"] = *env.NewString(msg.Topic())
								// msgDict["payload"] = *env.NewString(string(msg.Payload()))
								msgDict["qos"] = *env.NewInteger(int64(msg.Qos()))
								msgDict["retained"] = *env.NewBoolean(msg.Retained())
								msgDict["duplicate"] = *env.NewBoolean(msg.Duplicate())
								msgDict["message-id"] = *env.NewInteger(int64(msg.MessageID()))

								CallFunctionArgs2(handler,
									&psCallback,
									*env.NewString(string(msg.Payload())),
									*env.NewDict(msgDict),
									nil)
							}

							token := mqttClient.Subscribe(topic.Value, byte(qos.Value), callback)
							if token.Wait() && token.Error() != nil {
								ps.FailureFlag = true
								return MakeBuiltinError(ps, fmt.Sprintf("Failed to subscribe: %v", token.Error()), "Rye-mqtt-client//Subscribe")
							}

							return *env.NewInteger(1)
						default:
							fmt.Println(arg3.Inspect(*ps.Idx))
							ps.FailureFlag = true
							return MakeArgError(ps, 4, []env.Type{env.FunctionType}, "Rye-mqtt-client//Subscribe")
						}
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "Rye-mqtt-client//Subscribe")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-mqtt-client//Subscribe")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mqtt-client//Subscribe")
			}
		},
	},

	// Tests:
	// example { client |Subscribe-simple "sensors/temperature" fn { msg } { print msg.payload } }
	// Args:
	// * client: MQTT client connection (type: "Rye-mqtt-client")
	// * topic: Topic pattern to subscribe to (string)
	// * handler: Callback function to handle received messages
	// Returns:
	// * integer 1 for success (uses QoS 0)
	// * error if subscription fails
	"Rye-mqtt-client//Subscribe-simple": {
		Argsn: 3,
		Doc:   "Subscribes to an MQTT topic with default QoS 0 and a message handler function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch topic := arg1.(type) {
				case env.String:
					switch handler := arg2.(type) {
					case env.Function:
						mqttClient := client.Value.(mqtt.Client)

						callback := func(client mqtt.Client, msg mqtt.Message) {
							// Create a new program state copy for the callback
							psCallback := *ps
							psCallback.FailureFlag = false
							psCallback.ErrorFlag = false
							psCallback.ReturnFlag = false

							// Create message object as a Dict
							msgDict := make(map[string]any)
							msgDict["topic"] = *env.NewString(msg.Topic())
							msgDict["qos"] = *env.NewInteger(int64(msg.Qos()))
							msgDict["retained"] = *env.NewBoolean(msg.Retained())

							CallFunctionArgs2(handler, &psCallback, *env.NewDict(msgDict), env.Void{}, nil)
						}

						token := mqttClient.Subscribe(topic.Value, 0, callback)
						if token.Wait() && token.Error() != nil {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, fmt.Sprintf("Failed to subscribe: %v", token.Error()), "Rye-mqtt-client//Subscribe-simple")
						}

						return *env.NewInteger(1)
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "Rye-mqtt-client//Subscribe-simple")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-mqtt-client//Subscribe-simple")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mqtt-client//Subscribe-simple")
			}
		},
	},

	// Tests:
	// example { client |Unsubscribe "sensors/temperature" }
	// Args:
	// * client: MQTT client connection (type: "Rye-mqtt-client")
	// * topic: Topic to unsubscribe from (string)
	// Returns:
	// * integer 1 for success
	// * error if unsubscription fails
	"Rye-mqtt-client//Unsubscribe": {
		Argsn: 2,
		Doc:   "Unsubscribes from an MQTT topic.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch topic := arg1.(type) {
				case env.String:
					mqttClient := client.Value.(mqtt.Client)

					token := mqttClient.Unsubscribe(topic.Value)
					if token.Wait() && token.Error() != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to unsubscribe: %v", token.Error()), "Rye-mqtt-client//Unsubscribe")
					}

					return *env.NewInteger(1)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-mqtt-client//Unsubscribe")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mqtt-client//Unsubscribe")
			}
		},
	},

	// Tests:
	// example { client |Connected? }
	// Args:
	// * client: MQTT client connection (type: "Rye-mqtt-client")
	// Returns:
	// * integer 1 if connected, 0 if not connected
	"Rye-mqtt-client//Is-connected": {
		Argsn: 1,
		Doc:   "Checks if the MQTT client is currently connected to the broker.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				mqttClient := client.Value.(mqtt.Client)
				if mqttClient.IsConnected() {
					return *env.NewInteger(1)
				} else {
					return *env.NewInteger(0)
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mqtt-client//Connected?")
			}
		},
	},

	// Tests:
	// example { mqtt-options |Set-clean-session 1 |Set-keep-alive 60 |Set-timeout 30 }
	// Returns:
	// * new MQTT client options object (type: "Rye-mqtt-options")
	"mqtt-options": {
		Argsn: 0,
		Doc:   "Creates a new MQTT client options object for advanced configuration.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			opts := mqtt.NewClientOptions()
			return *env.NewNative(ps.Idx, opts, "Rye-mqtt-options")
		},
	},

	// Tests:
	// example { opts |Set-broker "tcp://localhost:1883" }
	// Args:
	// * options: MQTT options object (type: "Rye-mqtt-options")
	// * broker: Broker URI string
	// Returns:
	// * the same options object (for method chaining)
	"Rye-mqtt-options//Set-broker": {
		Argsn: 2,
		Doc:   "Sets the MQTT broker address in the options.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch opts := arg0.(type) {
			case env.Native:
				switch broker := arg1.(type) {
				case env.String:
					options := opts.Value.(*mqtt.ClientOptions)
					options.AddBroker(broker.Value)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-mqtt-options//Set-broker")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mqtt-options//Set-broker")
			}
		},
	},

	// Tests:
	// example { opts |Set-client-id "my-unique-client" }
	// Args:
	// * options: MQTT options object (type: "Rye-mqtt-options")
	// * client-id: Client identifier string
	// Returns:
	// * the same options object (for method chaining)
	"Rye-mqtt-options//Set-client-id": {
		Argsn: 2,
		Doc:   "Sets the client ID in the MQTT options.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch opts := arg0.(type) {
			case env.Native:
				switch clientID := arg1.(type) {
				case env.String:
					options := opts.Value.(*mqtt.ClientOptions)
					options.SetClientID(clientID.Value)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-mqtt-options//Set-client-id")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mqtt-options//Set-client-id")
			}
		},
	},

	// Tests:
	// example { opts |Set-keep-alive 60 }
	// Args:
	// * options: MQTT options object (type: "Rye-mqtt-options")
	// * seconds: Keep alive interval in seconds (integer)
	// Returns:
	// * the same options object (for method chaining)
	"Rye-mqtt-options//Set-keep-alive": {
		Argsn: 2,
		Doc:   "Sets the keep alive interval in seconds.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch opts := arg0.(type) {
			case env.Native:
				switch keepAlive := arg1.(type) {
				case env.Integer:
					options := opts.Value.(*mqtt.ClientOptions)
					options.SetKeepAlive(time.Duration(keepAlive.Value) * time.Second)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "Rye-mqtt-options//Set-keep-alive")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mqtt-options//Set-keep-alive")
			}
		},
	},

	// Tests:
	// example { opts |Connect-with-options }
	// Args:
	// * options: MQTT options object (type: "Rye-mqtt-options")
	// Returns:
	// * native MQTT client connection (type: "Rye-mqtt-client")
	// * error if connection fails
	"Rye-mqtt-options//Connect-with-options": {
		Argsn: 1,
		Doc:   "Creates and connects an MQTT client using the configured options.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch opts := arg0.(type) {
			case env.Native:
				options := opts.Value.(*mqtt.ClientOptions)
				client := mqtt.NewClient(options)

				if token := client.Connect(); token.Wait() && token.Error() != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, fmt.Sprintf("Failed to connect to MQTT broker: %v", token.Error()), "Rye-mqtt-options//Connect-with-options")
				}

				return *env.NewNative(ps.Idx, client, "Rye-mqtt-client")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mqtt-options//Connect-with-options")
			}
		},
	},
}
