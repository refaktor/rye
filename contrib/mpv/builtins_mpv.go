//go:build b_mpv
// +build b_mpv

package mpv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gen2brain/go-mpv"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// MpvPlayer wraps mpv instance with optional IPC support for audio level metering
type MpvPlayer struct {
	Mpv        *mpv.Mpv
	SocketPath string
}

// Helper to get mpv instance from player
func getMpv(player env.Native) (*MpvPlayer, bool) {
	p, ok := player.Value.(*MpvPlayer)
	return p, ok
}

// IPC helper to send command and get response
func (p *MpvPlayer) ipcCommand(cmd map[string]interface{}) (map[string]interface{}, error) {
	if p.SocketPath == "" {
		return nil, fmt.Errorf("IPC not enabled - use Init\\ipc")
	}

	// Check if socket exists
	if _, err := os.Stat(p.SocketPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("IPC socket not ready yet: %s", p.SocketPath)
	}

	conn, err := net.DialTimeout("unix", p.SocketPath, 500*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IPC: %v", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(500 * time.Millisecond))

	data, _ := json.Marshal(cmd)
	data = append(data, '\n')
	_, err = conn.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write to IPC: %v", err)
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read from IPC: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse IPC response: %v", err)
	}
	return result, nil
}

var Builtins_mpv = map[string]*env.Builtin{

	//
	// ##### MPV Player ##### "Functions for audio/video playback using mpv"
	//

	"mpv": {
		Argsn: 0,
		Doc:   "Creates a new mpv player instance. Must call init before use.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			m := mpv.New()
			player := &MpvPlayer{Mpv: m}
			return *env.NewNative(ps.Idx, player, "mpv-player")
		},
	},

	"mpv-player//Init": {
		Argsn: 1,
		Doc:   "Initializes the mpv player. Must be called before loading files.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Init")
				}
				if err := p.Mpv.Initialize(); err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Init")
				}
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Init")
			}
		},
	},

	"mpv-player//Init\\ipc": {
		Argsn: 1,
		Doc:   "Initializes mpv with IPC socket for audio level metering. Use with @label:lavfi filters.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Init\\ipc")
				}
				p.SocketPath = filepath.Join(os.TempDir(), fmt.Sprintf("rye-mpv-%d.sock", os.Getpid()))
				os.Remove(p.SocketPath)

				if err := p.Mpv.SetOptionString("input-ipc-server", p.SocketPath); err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Init\\ipc")
				}
				if err := p.Mpv.Initialize(); err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Init\\ipc")
				}
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Init\\ipc")
			}
		},
	},

	"mpv-player//Terminate": {
		Argsn: 1,
		Doc:   "Terminates the mpv player and releases resources.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Terminate")
				}
				if p.SocketPath != "" {
					os.Remove(p.SocketPath)
				}
				p.Mpv.TerminateDestroy()
				return *env.NewInteger(1)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Terminate")
			}
		},
	},

	//
	// ##### File Loading #####
	//

	"mpv-player//Load": {
		Argsn: 2,
		Doc:   "Loads a media file or stream URL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Load")
				}
				switch path := arg1.(type) {
				case env.String:
					if err := p.Mpv.Command([]string{"loadfile", path.Value}); err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Load")
					}
					return player
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "mpv-player//Load")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Load")
			}
		},
	},

	"mpv-player//Load\\append": {
		Argsn: 2,
		Doc:   "Appends a media file to the playlist.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Load\\append")
				}
				switch path := arg1.(type) {
				case env.String:
					if err := p.Mpv.Command([]string{"loadfile", path.Value, "append"}); err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Load\\append")
					}
					return player
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "mpv-player//Load\\append")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Load\\append")
			}
		},
	},

	//
	// ##### Playback Control #####
	//

	"mpv-player//Play": {
		Argsn: 1,
		Doc:   "Resumes playback.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Play")
				}
				p.Mpv.SetProperty("pause", mpv.FormatFlag, false)
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Play")
			}
		},
	},

	"mpv-player//Pause": {
		Argsn: 1,
		Doc:   "Pauses playback.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Pause")
				}
				p.Mpv.SetProperty("pause", mpv.FormatFlag, true)
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Pause")
			}
		},
	},

	"mpv-player//Toggle-pause": {
		Argsn: 1,
		Doc:   "Toggles play/pause.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Toggle-pause")
				}
				p.Mpv.Command([]string{"cycle", "pause"})
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Toggle-pause")
			}
		},
	},

	"mpv-player//Stop": {
		Argsn: 1,
		Doc:   "Stops playback.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Stop")
				}
				p.Mpv.Command([]string{"stop"})
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Stop")
			}
		},
	},

	"mpv-player//Seek": {
		Argsn: 2,
		Doc:   "Seeks relative to current position.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Seek")
				}
				var seconds float64
				switch s := arg1.(type) {
				case env.Decimal:
					seconds = s.Value
				case env.Integer:
					seconds = float64(s.Value)
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.DecimalType, env.IntegerType}, "mpv-player//Seek")
				}
				p.Mpv.Command([]string{"seek", fmt.Sprintf("%f", seconds), "relative"})
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Seek")
			}
		},
	},

	"mpv-player//Seek\\absolute": {
		Argsn: 2,
		Doc:   "Seeks to absolute position.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Seek\\absolute")
				}
				var seconds float64
				switch s := arg1.(type) {
				case env.Decimal:
					seconds = s.Value
				case env.Integer:
					seconds = float64(s.Value)
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.DecimalType, env.IntegerType}, "mpv-player//Seek\\absolute")
				}
				p.Mpv.Command([]string{"seek", fmt.Sprintf("%f", seconds), "absolute"})
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Seek\\absolute")
			}
		},
	},

	//
	// ##### Playlist #####
	//

	"mpv-player//Playlist-next": {
		Argsn: 1,
		Doc:   "Plays next item in playlist.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Playlist-next")
				}
				p.Mpv.Command([]string{"playlist-next"})
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Playlist-next")
			}
		},
	},

	"mpv-player//Playlist-prev": {
		Argsn: 1,
		Doc:   "Plays previous item in playlist.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Playlist-prev")
				}
				p.Mpv.Command([]string{"playlist-prev"})
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Playlist-prev")
			}
		},
	},

	"mpv-player//Playlist-clear": {
		Argsn: 1,
		Doc:   "Clears the playlist.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Playlist-clear")
				}
				p.Mpv.Command([]string{"playlist-clear"})
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Playlist-clear")
			}
		},
	},

	//
	// ##### Volume #####
	//

	"mpv-player//Volume?": {
		Argsn: 1,
		Doc:   "Gets current volume level.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Volume?")
				}
				vol, err := p.Mpv.GetProperty("volume", mpv.FormatDouble)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Volume?")
				}
				return *env.NewDecimal(vol.(float64))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Volume?")
			}
		},
	},

	"mpv-player//Set-volume": {
		Argsn: 2,
		Doc:   "Sets volume level (0-100+).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Set-volume")
				}
				var level float64
				switch v := arg1.(type) {
				case env.Decimal:
					level = v.Value
				case env.Integer:
					level = float64(v.Value)
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.DecimalType, env.IntegerType}, "mpv-player//Set-volume")
				}
				p.Mpv.SetProperty("volume", mpv.FormatDouble, level)
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Set-volume")
			}
		},
	},

	"mpv-player//Mute": {
		Argsn: 1,
		Doc:   "Mutes audio.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Mute")
				}
				p.Mpv.SetProperty("mute", mpv.FormatFlag, true)
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Mute")
			}
		},
	},

	"mpv-player//Unmute": {
		Argsn: 1,
		Doc:   "Unmutes audio.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Unmute")
				}
				p.Mpv.SetProperty("mute", mpv.FormatFlag, false)
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Unmute")
			}
		},
	},

	"mpv-player//Toggle-mute": {
		Argsn: 1,
		Doc:   "Toggles mute.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Toggle-mute")
				}
				p.Mpv.Command([]string{"cycle", "mute"})
				return player
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Toggle-mute")
			}
		},
	},

	"mpv-player//Muted?": {
		Argsn: 1,
		Doc:   "Returns 1 if muted, 0 otherwise.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Muted?")
				}
				muted, err := p.Mpv.GetProperty("mute", mpv.FormatFlag)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Muted?")
				}
				if muted.(bool) {
					return *env.NewInteger(1)
				}
				return *env.NewInteger(0)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Muted?")
			}
		},
	},

	//
	// ##### Status #####
	//

	"mpv-player//Position?": {
		Argsn: 1,
		Doc:   "Gets current playback position in seconds.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Position?")
				}
				pos, err := p.Mpv.GetProperty("time-pos", mpv.FormatDouble)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "not available", "mpv-player//Position?")
				}
				return *env.NewDecimal(pos.(float64))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Position?")
			}
		},
	},

	"mpv-player//Duration?": {
		Argsn: 1,
		Doc:   "Gets duration in seconds (not available for streams).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Duration?")
				}
				dur, err := p.Mpv.GetProperty("duration", mpv.FormatDouble)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "not available", "mpv-player//Duration?")
				}
				return *env.NewDecimal(dur.(float64))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Duration?")
			}
		},
	},

	"mpv-player//Paused?": {
		Argsn: 1,
		Doc:   "Returns 1 if paused, 0 otherwise.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Paused?")
				}
				paused, err := p.Mpv.GetProperty("pause", mpv.FormatFlag)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Paused?")
				}
				if paused.(bool) {
					return *env.NewInteger(1)
				}
				return *env.NewInteger(0)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Paused?")
			}
		},
	},

	"mpv-player//Idle?": {
		Argsn: 1,
		Doc:   "Returns 1 if idle (nothing playing), 0 otherwise.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Idle?")
				}
				idle, err := p.Mpv.GetProperty("idle-active", mpv.FormatFlag)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Idle?")
				}
				if idle.(bool) {
					return *env.NewInteger(1)
				}
				return *env.NewInteger(0)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Idle?")
			}
		},
	},

	"mpv-player//Media-title?": {
		Argsn: 1,
		Doc:   "Gets the media title.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Media-title?")
				}
				title, err := p.Mpv.GetProperty("media-title", mpv.FormatString)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "not available", "mpv-player//Media-title?")
				}
				return *env.NewString(title.(string))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Media-title?")
			}
		},
	},

	"mpv-player//Path?": {
		Argsn: 1,
		Doc:   "Gets the file path or URL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Path?")
				}
				path, err := p.Mpv.GetProperty("path", mpv.FormatString)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "not available", "mpv-player//Path?")
				}
				return *env.NewString(path.(string))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Path?")
			}
		},
	},

	"mpv-player//Icy-title?": {
		Argsn: 1,
		Doc:   "Gets ICY stream title (current song from internet radio).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Icy-title?")
				}
				title, err := p.Mpv.GetProperty("metadata/by-key/icy-title", mpv.FormatString)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "not available", "mpv-player//Icy-title?")
				}
				return *env.NewString(title.(string))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Icy-title?")
			}
		},
	},

	"mpv-player//Cache-duration?": {
		Argsn: 1,
		Doc:   "Gets buffered duration in seconds.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Cache-duration?")
				}
				cache, err := p.Mpv.GetProperty("demuxer-cache-duration", mpv.FormatDouble)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "not available", "mpv-player//Cache-duration?")
				}
				return *env.NewDecimal(cache.(float64))
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Cache-duration?")
			}
		},
	},

	"mpv-player//Audio-info?": {
		Argsn: 1,
		Doc:   "Gets audio info as dict (codec, channels, samplerate).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Audio-info?")
				}
				info := make(map[string]any)
				if codec, err := p.Mpv.GetProperty("audio-codec-name", mpv.FormatString); err == nil {
					info["codec"] = *env.NewString(codec.(string))
				}
				if ch, err := p.Mpv.GetProperty("audio-params/channel-count", mpv.FormatInt64); err == nil {
					info["channels"] = *env.NewInteger(ch.(int64))
				}
				if sr, err := p.Mpv.GetProperty("audio-params/samplerate", mpv.FormatInt64); err == nil {
					info["samplerate"] = *env.NewInteger(sr.(int64))
				}
				return *env.NewDict(info)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Audio-info?")
			}
		},
	},

	//
	// ##### Audio Level Metering (requires IPC) #####
	//

	"mpv-player//Ipc-path?": {
		Argsn: 1,
		Doc:   "Returns the IPC socket path (for debugging).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Ipc-path?")
				}
				if p.SocketPath == "" {
					return *env.NewString("(not enabled - use Init\\ipc)")
				}
				// Check if socket exists
				if _, err := os.Stat(p.SocketPath); os.IsNotExist(err) {
					return *env.NewString(p.SocketPath + " (NOT FOUND)")
				}
				return *env.NewString(p.SocketPath + " (exists)")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Ipc-path?")
			}
		},
	},

	"mpv-player//Set-audio-filters": {
		Argsn: 2,
		Doc:   "Sets audio filters. For level metering use: @label:lavfi=[astats=metadata=1:reset=1]",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Set-audio-filters")
				}
				switch filters := arg1.(type) {
				case env.String:
					if err := p.Mpv.SetPropertyString("af", filters.Value); err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Set-audio-filters")
					}
					return player
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "mpv-player//Set-audio-filters")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Set-audio-filters")
			}
		},
	},

	"mpv-player//Audio-levels?": {
		Argsn: 2,
		Doc:   "Gets audio levels via IPC. Arg is filter label. Returns dict with RMS/Peak levels.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Audio-levels?")
				}
				switch label := arg1.(type) {
				case env.String:
					result, err := p.ipcCommand(map[string]interface{}{
						"command": []interface{}{"get_property", "af-metadata/" + label.Value},
					})
					if err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Audio-levels?")
					}

					if result["error"] != "success" {
						errMsg := "metadata not available"
						if e, ok := result["error"].(string); ok {
							errMsg = e
						}
						return evaldo.MakeBuiltinError(ps, errMsg, "mpv-player//Audio-levels?")
					}

					data, ok := result["data"].(map[string]interface{})
					if !ok {
						return evaldo.MakeBuiltinError(ps, "invalid metadata format", "mpv-player//Audio-levels?")
					}

					levels := make(map[string]any)
					// Extract key audio levels
					keyMap := map[string]string{
						"lavfi.astats.1.RMS_level":        "rms-left",
						"lavfi.astats.2.RMS_level":        "rms-right",
						"lavfi.astats.Overall.RMS_level":  "rms",
						"lavfi.astats.1.Peak_level":       "peak-left",
						"lavfi.astats.2.Peak_level":       "peak-right",
						"lavfi.astats.Overall.Peak_level": "peak",
					}
					for srcKey, dstKey := range keyMap {
						if val, exists := data[srcKey]; exists {
							if strVal, ok := val.(string); ok {
								if f, err := strconv.ParseFloat(strVal, 64); err == nil {
									levels[dstKey] = *env.NewDecimal(f)
								}
							}
						}
					}
					return *env.NewDict(levels)
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "mpv-player//Audio-levels?")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Audio-levels?")
			}
		},
	},

	//
	// ##### Configuration #####
	//

	"mpv-player//Set-option": {
		Argsn: 3,
		Doc:   "Sets an mpv option (before init).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Set-option")
				}
				switch name := arg1.(type) {
				case env.String:
					switch value := arg2.(type) {
					case env.String:
						if err := p.Mpv.SetOptionString(name.Value, value.Value); err != nil {
							return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Set-option")
						}
						return player
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "mpv-player//Set-option")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "mpv-player//Set-option")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Set-option")
			}
		},
	},

	"mpv-player//Set-property": {
		Argsn: 3,
		Doc:   "Sets an mpv property (after init).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Set-property")
				}
				switch name := arg1.(type) {
				case env.String:
					var err error
					switch value := arg2.(type) {
					case env.String:
						err = p.Mpv.SetPropertyString(name.Value, value.Value)
					case env.Decimal:
						err = p.Mpv.SetProperty(name.Value, mpv.FormatDouble, value.Value)
					case env.Integer:
						err = p.Mpv.SetProperty(name.Value, mpv.FormatInt64, value.Value)
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType, env.DecimalType, env.IntegerType}, "mpv-player//Set-property")
					}
					if err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Set-property")
					}
					return player
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "mpv-player//Set-property")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Set-property")
			}
		},
	},

	"mpv-player//Get-property": {
		Argsn: 2,
		Doc:   "Gets an mpv property as string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Get-property")
				}
				switch name := arg1.(type) {
				case env.String:
					val, err := p.Mpv.GetProperty(name.Value, mpv.FormatString)
					if err != nil {
						return evaldo.MakeBuiltinError(ps, "not available", "mpv-player//Get-property")
					}
					return *env.NewString(val.(string))
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "mpv-player//Get-property")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Get-property")
			}
		},
	},

	"mpv-player//Command": {
		Argsn: 2,
		Doc:   "Executes a raw mpv command.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch player := arg0.(type) {
			case env.Native:
				p, ok := getMpv(player)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Invalid mpv player", "mpv-player//Command")
				}
				switch args := arg1.(type) {
				case env.Block:
					cmdArgs := make([]string, 0, args.Series.Len())
					for i := 0; i < args.Series.Len(); i++ {
						item := args.Series.Get(i)
						if str, ok := item.(env.String); ok {
							cmdArgs = append(cmdArgs, str.Value)
						} else {
							return evaldo.MakeBuiltinError(ps, "All arguments must be strings", "mpv-player//Command")
						}
					}
					if err := p.Mpv.Command(cmdArgs); err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "mpv-player//Command")
					}
					return player
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.BlockType}, "mpv-player//Command")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "mpv-player//Command")
			}
		},
	},
}
