//go:build !no_docker
// +build !no_docker

package evaldo

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/refaktor/rye/env"
)

var Builtins_docker = map[string]*env.Builtin{

	//
	// ##### Docker ##### "Docker container management functions"
	//
	// Example:
	//  docker: docker-client
	//  docker .docker-client//Containers? |for { -> "id" |print }
	//  docker .docker-client//Logs? "container-id"
	//  docker .docker-client//Kill "container-id"
	//
	// Tests:
	// ; equal { docker-client |type? } 'native
	// Args:
	// * None
	// Returns:
	// * native Docker client object
	"docker-client": {
		Argsn: 0,
		Doc:   "Creates a new Docker client connection using environment variables.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			apiClient, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
			if err != nil {
				return MakeBuiltinError(ps, fmt.Sprintf("failed to create Docker client: %v", err), "docker-client")
			}
			return *env.NewNative(ps.Idx, apiClient, "docker-client")
		},
	},

	// Tests:
	// ; equal { docker-client .docker-client//Containers? |type? } 'block
	// Args:
	// * client: Docker client object
	// Returns:
	// * block of container information dicts
	"docker-client//Containers?": {
		Argsn: 1,
		Doc:   "Lists all Docker containers (including stopped ones).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dclient := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(dclient.Kind.Index) != "docker-client" {
					return MakeBuiltinError(ps, "expected a Docker client object", "docker-client//Containers?")
				}

				apiClient := dclient.Value.(*dockerclient.Client)
				containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{All: true})
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to list containers: %v", err), "docker-client//Containers?")
				}

				// Convert containers to Rye block of dicts
				result := make([]env.Object, 0, len(containers))
				for _, ctr := range containers {
					containerDict := env.NewDict(make(map[string]any))
					containerDict.Data["id"] = ctr.ID
					containerDict.Data["short-id"] = ctr.ID[:12]
					containerDict.Data["image"] = ctr.Image
					containerDict.Data["image-id"] = ctr.ImageID
					containerDict.Data["command"] = ctr.Command
					containerDict.Data["created"] = time.Unix(ctr.Created, 0).Format(time.RFC3339)
					containerDict.Data["state"] = ctr.State
					containerDict.Data["status"] = ctr.Status

					// Convert names to block
					names := make([]env.Object, 0, len(ctr.Names))
					for _, name := range ctr.Names {
						names = append(names, *env.NewString(strings.TrimPrefix(name, "/")))
					}
					containerDict.Data["names"] = *env.NewBlock(*env.NewTSeries(names))

					// Convert ports to block of dicts
					ports := make([]env.Object, 0, len(ctr.Ports))
					for _, port := range ctr.Ports {
						portDict := env.NewDict(make(map[string]any))
						portDict.Data["ip"] = port.IP
						portDict.Data["private-port"] = int64(port.PrivatePort)
						portDict.Data["public-port"] = int64(port.PublicPort)
						portDict.Data["type"] = port.Type
						ports = append(ports, *portDict)
					}
					containerDict.Data["ports"] = *env.NewBlock(*env.NewTSeries(ports))

					// Convert labels to dict
					labelsDict := env.NewDict(make(map[string]any))
					for k, v := range ctr.Labels {
						labelsDict.Data[k] = v
					}
					containerDict.Data["labels"] = *labelsDict

					result = append(result, *containerDict)
				}

				return *env.NewBlock(*env.NewTSeries(result))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "docker-client//Containers?")
			}
		},
	},

	// Tests:
	// ; equal { docker-client .docker-client//Running-containers? |type? } 'block
	// Args:
	// * client: Docker client object
	// Returns:
	// * block of running container information dicts
	"docker-client//Running-containers?": {
		Argsn: 1,
		Doc:   "Lists only running Docker containers.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dclient := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(dclient.Kind.Index) != "docker-client" {
					return MakeBuiltinError(ps, "expected a Docker client object", "docker-client//Running-containers?")
				}

				apiClient := dclient.Value.(*dockerclient.Client)
				containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{All: false})
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to list containers: %v", err), "docker-client//Running-containers?")
				}

				// Convert containers to Rye block of dicts
				result := make([]env.Object, 0, len(containers))
				for _, ctr := range containers {
					containerDict := env.NewDict(make(map[string]any))
					containerDict.Data["id"] = ctr.ID
					containerDict.Data["short-id"] = ctr.ID[:12]
					containerDict.Data["image"] = ctr.Image
					containerDict.Data["image-id"] = ctr.ImageID
					containerDict.Data["command"] = ctr.Command
					containerDict.Data["created"] = time.Unix(ctr.Created, 0).Format(time.RFC3339)
					containerDict.Data["state"] = ctr.State
					containerDict.Data["status"] = ctr.Status

					// Convert names to block
					names := make([]env.Object, 0, len(ctr.Names))
					for _, name := range ctr.Names {
						names = append(names, *env.NewString(strings.TrimPrefix(name, "/")))
					}
					containerDict.Data["names"] = *env.NewBlock(*env.NewTSeries(names))

					result = append(result, *containerDict)
				}

				return *env.NewBlock(*env.NewTSeries(result))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "docker-client//Running-containers?")
			}
		},
	},

	// Tests:
	// ; equal { docker-client .docker-client//Logs? "container-id" |type? } 'string
	// Args:
	// * client: Docker client object
	// * container-id: String ID of the container (full or short)
	// Returns:
	// * string containing container logs
	"docker-client//Logs?": {
		Argsn: 2,
		Doc:   "Gets logs from a Docker container.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dclient := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(dclient.Kind.Index) != "docker-client" {
					return MakeBuiltinError(ps, "expected a Docker client object", "docker-client//Logs?")
				}

				var containerID string
				switch id := arg1.(type) {
				case env.String:
					containerID = id.Value
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "docker-client//Logs?")
				}

				apiClient := dclient.Value.(*dockerclient.Client)
				options := container.LogsOptions{
					ShowStdout: true,
					ShowStderr: true,
					Timestamps: false,
					Follow:     false,
					Tail:       "all",
				}

				reader, err := apiClient.ContainerLogs(context.Background(), containerID, options)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get container logs: %v", err), "docker-client//Logs?")
				}
				defer reader.Close()

				// Read the logs
				logs, err := io.ReadAll(reader)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to read container logs: %v", err), "docker-client//Logs?")
				}

				// Docker multiplexes stdout/stderr with 8-byte header per message
				// Strip the headers for cleaner output
				cleanLogs := stripDockerLogHeaders(logs)

				return *env.NewString(cleanLogs)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "docker-client//Logs?")
			}
		},
	},

	// Tests:
	// ; equal { docker-client .docker-client//Logs\tail? "container-id" 100 |type? } 'string
	// Args:
	// * client: Docker client object
	// * container-id: String ID of the container (full or short)
	// * lines: Integer number of lines to return from the end
	// Returns:
	// * string containing container logs (last N lines)
	"docker-client//Logs\\tail?": {
		Argsn: 3,
		Doc:   "Gets last N lines of logs from a Docker container.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dclient := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(dclient.Kind.Index) != "docker-client" {
					return MakeBuiltinError(ps, "expected a Docker client object", "docker-client//Logs\\tail?")
				}

				var containerID string
				switch id := arg1.(type) {
				case env.String:
					containerID = id.Value
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "docker-client//Logs\\tail?")
				}

				var tailLines string
				switch lines := arg2.(type) {
				case env.Integer:
					tailLines = fmt.Sprintf("%d", lines.Value)
				default:
					return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "docker-client//Logs\\tail?")
				}

				apiClient := dclient.Value.(*dockerclient.Client)
				options := container.LogsOptions{
					ShowStdout: true,
					ShowStderr: true,
					Timestamps: false,
					Follow:     false,
					Tail:       tailLines,
				}

				reader, err := apiClient.ContainerLogs(context.Background(), containerID, options)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to get container logs: %v", err), "docker-client//Logs\\tail?")
				}
				defer reader.Close()

				logs, err := io.ReadAll(reader)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to read container logs: %v", err), "docker-client//Logs\\tail?")
				}

				cleanLogs := stripDockerLogHeaders(logs)
				return *env.NewString(cleanLogs)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "docker-client//Logs\\tail?")
			}
		},
	},

	// Tests:
	// ; equal { docker-client .docker-client//Kill "container-id" |type? } 'integer
	// Args:
	// * client: Docker client object
	// * container-id: String ID of the container to kill
	// Returns:
	// * Integer 1 on success
	"docker-client//Kill": {
		Argsn: 2,
		Doc:   "Kills a running Docker container with SIGKILL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dclient := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(dclient.Kind.Index) != "docker-client" {
					return MakeBuiltinError(ps, "expected a Docker client object", "docker-client//Kill")
				}

				var containerID string
				switch id := arg1.(type) {
				case env.String:
					containerID = id.Value
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "docker-client//Kill")
				}

				apiClient := dclient.Value.(*dockerclient.Client)
				err := apiClient.ContainerKill(context.Background(), containerID, "SIGKILL")
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to kill container: %v", err), "docker-client//Kill")
				}

				return *env.NewInteger(1)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "docker-client//Kill")
			}
		},
	},

	// Tests:
	// ; equal { docker-client .docker-client//Kill\signal "container-id" "SIGTERM" |type? } 'integer
	// Args:
	// * client: Docker client object
	// * container-id: String ID of the container
	// * signal: String signal to send (e.g., "SIGTERM", "SIGINT", "SIGKILL")
	// Returns:
	// * Integer 1 on success
	"docker-client//Kill\\signal": {
		Argsn: 3,
		Doc:   "Sends a signal to a Docker container.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dclient := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(dclient.Kind.Index) != "docker-client" {
					return MakeBuiltinError(ps, "expected a Docker client object", "docker-client//Kill\\signal")
				}

				var containerID string
				switch id := arg1.(type) {
				case env.String:
					containerID = id.Value
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "docker-client//Kill\\signal")
				}

				var signal string
				switch sig := arg2.(type) {
				case env.String:
					signal = sig.Value
				default:
					return MakeArgError(ps, 3, []env.Type{env.StringType}, "docker-client//Kill\\signal")
				}

				apiClient := dclient.Value.(*dockerclient.Client)
				err := apiClient.ContainerKill(context.Background(), containerID, signal)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to send signal to container: %v", err), "docker-client//Kill\\signal")
				}

				return *env.NewInteger(1)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "docker-client//Kill\\signal")
			}
		},
	},

	// Tests:
	// ; equal { docker-client .docker-client//Stop "container-id" |type? } 'integer
	// Args:
	// * client: Docker client object
	// * container-id: String ID of the container to stop
	// Returns:
	// * Integer 1 on success
	"docker-client//Stop": {
		Argsn: 2,
		Doc:   "Stops a running Docker container gracefully (with timeout).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dclient := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(dclient.Kind.Index) != "docker-client" {
					return MakeBuiltinError(ps, "expected a Docker client object", "docker-client//Stop")
				}

				var containerID string
				switch id := arg1.(type) {
				case env.String:
					containerID = id.Value
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "docker-client//Stop")
				}

				apiClient := dclient.Value.(*dockerclient.Client)
				timeout := 10 // Default timeout in seconds
				err := apiClient.ContainerStop(context.Background(), containerID, container.StopOptions{Timeout: &timeout})
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to stop container: %v", err), "docker-client//Stop")
				}

				return *env.NewInteger(1)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "docker-client//Stop")
			}
		},
	},

	// Tests:
	// ; equal { docker-client .docker-client//Start "container-id" |type? } 'integer
	// Args:
	// * client: Docker client object
	// * container-id: String ID of the container to start
	// Returns:
	// * Integer 1 on success
	"docker-client//Start": {
		Argsn: 2,
		Doc:   "Starts a stopped Docker container.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dclient := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(dclient.Kind.Index) != "docker-client" {
					return MakeBuiltinError(ps, "expected a Docker client object", "docker-client//Start")
				}

				var containerID string
				switch id := arg1.(type) {
				case env.String:
					containerID = id.Value
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "docker-client//Start")
				}

				apiClient := dclient.Value.(*dockerclient.Client)
				err := apiClient.ContainerStart(context.Background(), containerID, container.StartOptions{})
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to start container: %v", err), "docker-client//Start")
				}

				return *env.NewInteger(1)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "docker-client//Start")
			}
		},
	},

	// Tests:
	// ; equal { docker-client .docker-client//Restart "container-id" |type? } 'integer
	// Args:
	// * client: Docker client object
	// * container-id: String ID of the container to restart
	// Returns:
	// * Integer 1 on success
	"docker-client//Restart": {
		Argsn: 2,
		Doc:   "Restarts a Docker container.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dclient := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(dclient.Kind.Index) != "docker-client" {
					return MakeBuiltinError(ps, "expected a Docker client object", "docker-client//Restart")
				}

				var containerID string
				switch id := arg1.(type) {
				case env.String:
					containerID = id.Value
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "docker-client//Restart")
				}

				apiClient := dclient.Value.(*dockerclient.Client)
				timeout := 10
				err := apiClient.ContainerRestart(context.Background(), containerID, container.StopOptions{Timeout: &timeout})
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to restart container: %v", err), "docker-client//Restart")
				}

				return *env.NewInteger(1)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "docker-client//Restart")
			}
		},
	},

	// Tests:
	// ; equal { docker-client .docker-client//Inspect? "container-id" |type? } 'dict
	// Args:
	// * client: Docker client object
	// * container-id: String ID of the container to inspect
	// Returns:
	// * Dict containing detailed container information
	"docker-client//Inspect?": {
		Argsn: 2,
		Doc:   "Gets detailed information about a Docker container.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dclient := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(dclient.Kind.Index) != "docker-client" {
					return MakeBuiltinError(ps, "expected a Docker client object", "docker-client//Inspect?")
				}

				var containerID string
				switch id := arg1.(type) {
				case env.String:
					containerID = id.Value
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "docker-client//Inspect?")
				}

				apiClient := dclient.Value.(*dockerclient.Client)
				info, err := apiClient.ContainerInspect(context.Background(), containerID)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to inspect container: %v", err), "docker-client//Inspect?")
				}

				// Build detailed dict
				result := env.NewDict(make(map[string]any))
				result.Data["id"] = info.ID
				result.Data["name"] = strings.TrimPrefix(info.Name, "/")
				result.Data["created"] = info.Created
				result.Data["path"] = info.Path
				result.Data["image"] = info.Image
				result.Data["restart-count"] = int64(info.RestartCount)
				result.Data["platform"] = info.Platform

				// State information
				stateDict := env.NewDict(make(map[string]any))
				stateDict.Data["status"] = info.State.Status
				stateDict.Data["running"] = info.State.Running
				stateDict.Data["paused"] = info.State.Paused
				stateDict.Data["restarting"] = info.State.Restarting
				stateDict.Data["oom-killed"] = info.State.OOMKilled
				stateDict.Data["dead"] = info.State.Dead
				stateDict.Data["pid"] = int64(info.State.Pid)
				stateDict.Data["exit-code"] = int64(info.State.ExitCode)
				stateDict.Data["started-at"] = info.State.StartedAt
				stateDict.Data["finished-at"] = info.State.FinishedAt
				result.Data["state"] = *stateDict

				// Config information
				if info.Config != nil {
					configDict := env.NewDict(make(map[string]any))
					configDict.Data["hostname"] = info.Config.Hostname
					configDict.Data["domainname"] = info.Config.Domainname
					configDict.Data["user"] = info.Config.User
					configDict.Data["image"] = info.Config.Image
					configDict.Data["working-dir"] = info.Config.WorkingDir
					configDict.Data["tty"] = info.Config.Tty

					// Environment variables
					envVars := make([]env.Object, 0, len(info.Config.Env))
					for _, e := range info.Config.Env {
						envVars = append(envVars, *env.NewString(e))
					}
					configDict.Data["env"] = *env.NewBlock(*env.NewTSeries(envVars))

					// Command
					cmdVars := make([]env.Object, 0, len(info.Config.Cmd))
					for _, c := range info.Config.Cmd {
						cmdVars = append(cmdVars, *env.NewString(c))
					}
					configDict.Data["cmd"] = *env.NewBlock(*env.NewTSeries(cmdVars))

					result.Data["config"] = *configDict
				}

				// Network settings
				if info.NetworkSettings != nil {
					networkDict := env.NewDict(make(map[string]any))
					networkDict.Data["ip-address"] = info.NetworkSettings.IPAddress
					networkDict.Data["gateway"] = info.NetworkSettings.Gateway
					networkDict.Data["mac-address"] = info.NetworkSettings.MacAddress

					// Networks
					networksDict := env.NewDict(make(map[string]any))
					for name, network := range info.NetworkSettings.Networks {
						netDict := env.NewDict(make(map[string]any))
						netDict.Data["ip-address"] = network.IPAddress
						netDict.Data["gateway"] = network.Gateway
						netDict.Data["mac-address"] = network.MacAddress
						netDict.Data["network-id"] = network.NetworkID
						networksDict.Data[name] = *netDict
					}
					networkDict.Data["networks"] = *networksDict

					result.Data["network-settings"] = *networkDict
				}

				return *result
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "docker-client//Inspect?")
			}
		},
	},

	// Tests:
	// ; equal { docker-client .docker-client//Close |type? } 'integer
	// Args:
	// * client: Docker client object
	// Returns:
	// * Integer 1 on success
	"docker-client//Close": {
		Argsn: 1,
		Doc:   "Closes the Docker client connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dclient := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(dclient.Kind.Index) != "docker-client" {
					return MakeBuiltinError(ps, "expected a Docker client object", "docker-client//Close")
				}

				apiClient := dclient.Value.(*dockerclient.Client)
				err := apiClient.Close()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("failed to close Docker client: %v", err), "docker-client//Close")
				}

				return *env.NewInteger(1)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "docker-client//Close")
			}
		},
	},
}

// stripDockerLogHeaders removes the 8-byte header that Docker adds to multiplexed log output
func stripDockerLogHeaders(data []byte) string {
	var result strings.Builder
	i := 0
	for i < len(data) {
		// Each frame has an 8-byte header:
		// [0] = stream type (0=stdin, 1=stdout, 2=stderr)
		// [1-3] = reserved
		// [4-7] = size (big endian)
		if i+8 > len(data) {
			// Not enough data for header, write remaining as-is
			result.Write(data[i:])
			break
		}

		// Read the size from bytes 4-7 (big endian)
		size := int(data[i+4])<<24 | int(data[i+5])<<16 | int(data[i+6])<<8 | int(data[i+7])

		// Skip the header
		i += 8

		// Read the actual log content
		if i+size > len(data) {
			// Partial frame, write what we have
			result.Write(data[i:])
			break
		}

		result.Write(data[i : i+size])
		i += size
	}
	return result.String()
}
