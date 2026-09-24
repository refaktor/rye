//go:build !no_os && !b_wasm

package batteries

import (
	"os/exec"

	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

var Builtins_tmux = map[string]*env.Builtin{
	// TMUX related builtins loaded under tmux context
	"is-available": {
		Argsn: 0,
		Doc:   "Checks if tmux is available on the system.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			_, err := exec.LookPath("tmux")
			return *env.NewBoolean(err == nil)
		},
	},

	"session": {
		Argsn: 1,
		Doc:   "Creates a new tmux session with the given name using gotmux library.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			tmux, err := gotmux.DefaultTmux()
			if err != nil {
				return evaldo.MakeBuiltinError(ps, "Failed to connect to tmux: "+err.Error(), "tmux-new-session")
			}

			switch name := arg0.(type) {
			case env.String:
				session, err := tmux.NewSession(&gotmux.SessionOptions{Name: name.Value})
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "Failed to create session: "+err.Error(), "tmux-new-session")
				}
				return *env.NewNative(ps.Idx, session, "tmux-session")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "tmux-new-session")
			}
		},
	},

	"sessions?": {
		Argsn: 0,
		Doc:   "Lists all tmux sessions using gotmux library.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			tmux, err := gotmux.DefaultTmux()
			if err != nil {
				return evaldo.MakeBuiltinError(ps, "Failed to connect to tmux: "+err.Error(), "tmux-list-sessions")
			}
			sessions, err := tmux.ListSessions()
			if err != nil {
				return evaldo.MakeBuiltinError(ps, "Failed to list sessions: "+err.Error(), "tmux-list-sessions")
			}
			items := make([]env.Object, len(sessions))
			for i, session := range sessions {
				items[i] = *env.NewNative(ps.Idx, session, "tmux-session")
			}
			return *env.NewBlock(*env.NewTSeries(items))
		},
	},

	"session?": {
		Argsn: 1,
		Doc:   "Gets a tmux session by name using gotmux library.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			tmux, err := gotmux.DefaultTmux()
			if err != nil {
				return evaldo.MakeBuiltinError(ps, "Failed to connect to tmux: "+err.Error(), "tmux-get-session")
			}
			switch name := arg0.(type) {
			case env.String:
				session, err := tmux.GetSessionByName(name.Value)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, "Failed to get session: "+err.Error(), "tmux-get-session")
				}
				return *env.NewNative(ps.Idx, session, "tmux-session")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "tmux-get-session")
			}
		},
	},

	"tmux-session//Window": {
		Argsn: 1,
		Doc:   "Creates a new window in the given tmux session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sess := arg0.(type) {
			case env.Native:
				if session, ok := sess.Value.(*gotmux.Session); ok {
					window, err := session.New()
					if err != nil {
						return evaldo.MakeBuiltinError(ps, "Failed to create window: "+err.Error(), "tmux-new-window")
					}
					return *env.NewNative(ps.Idx, window, "tmux-window")
				}
				return evaldo.MakeBuiltinError(ps, "Expected tmux-session object", "tmux-new-window")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-new-window")
			}
		},
	},

	"tmux-session//window\\named": {
		Argsn: 2,
		Doc:   "Creates a new window in the given tmux session with a specific name.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sess := arg0.(type) {
			case env.Native:
				if session, ok := sess.Value.(*gotmux.Session); ok {
					switch name := arg1.(type) {
					case env.String:
						window, err := session.NewWindow(&gotmux.NewWindowOptions{WindowName: name.Value})
						if err != nil {
							return evaldo.MakeBuiltinError(ps, "Failed to create named window: "+err.Error(), "tmux-new-window-named")
						}
						return *env.NewNative(ps.Idx, window, "tmux-window")
					default:
						return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "tmux-new-window-named")
					}
				}
				return evaldo.MakeBuiltinError(ps, "Expected tmux-session object", "tmux-new-window-named")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-new-window-named")
			}
		},
	},

	"tmux-session//Windows?": {
		Argsn: 1,
		Doc:   "Lists all windows in the given tmux session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sess := arg0.(type) {
			case env.Native:
				if session, ok := sess.Value.(*gotmux.Session); ok {
					windows, err := session.ListWindows()
					if err != nil {
						return evaldo.MakeBuiltinError(ps, "Failed to list windows: "+err.Error(), "tmux-list-windows")
					}
					items := make([]env.Object, len(windows))
					for i, window := range windows {
						items[i] = *env.NewNative(ps.Idx, window, "tmux-window")
					}
					return *env.NewBlock(*env.NewTSeries(items))
				}
				return evaldo.MakeBuiltinError(ps, "Expected tmux-session object", "tmux-list-windows")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-list-windows")
			}
		},
	},

	"tmux-session//Pane?": {
		Argsn: 2,
		Doc:   "Gets a pane from a tmux window by index.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				if window, ok := win.Value.(*gotmux.Window); ok {
					switch idx := arg1.(type) {
					case env.Integer:
						pane, err := window.GetPaneByIndex(int(idx.Value))
						if err != nil {
							return evaldo.MakeBuiltinError(ps, "Failed to get pane: "+err.Error(), "tmux-get-pane")
						}
						return *env.NewNative(ps.Idx, pane, "tmux-pane")
					default:
						return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "tmux-get-pane")
					}
				}
				return evaldo.MakeBuiltinError(ps, "Expected tmux-window object", "tmux-get-pane")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-get-pane")
			}
		},
	},

	"tmux-pane//Split-pane": {
		Argsn: 1,
		Doc:   "Splits a tmux pane horizontally.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				if pane, ok := p.Value.(*gotmux.Pane); ok {
					err := pane.Split()
					if err != nil {
						return evaldo.MakeBuiltinError(ps, "Failed to split pane: "+err.Error(), "tmux-split-pane")
					}
					return arg0
				}
				return evaldo.MakeBuiltinError(ps, "Expected tmux-pane object", "tmux-split-pane")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-split-pane")
			}
		},
	},

	"tmux-pane//Send-keys": {
		Argsn: 2,
		Doc:   "Sends keys/command to a tmux pane.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch p := arg0.(type) {
			case env.Native:
				if pane, ok := p.Value.(*gotmux.Pane); ok {
					switch cmd := arg1.(type) {
					case env.String:
						err := pane.SendKeys(cmd.Value)
						if err != nil {
							return evaldo.MakeBuiltinError(ps, "Failed to send keys: "+err.Error(), "tmux-send-keys")
						}
						return arg0
					default:
						return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "tmux-send-keys")
					}
				}
				return evaldo.MakeBuiltinError(ps, "Expected tmux-pane object", "tmux-send-keys")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-send-keys")
			}
		},
	},

	"tmux-session//Kill": {
		Argsn: 1,
		Doc:   "Kills/destroys a tmux session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sess := arg0.(type) {
			case env.Native:
				if session, ok := sess.Value.(*gotmux.Session); ok {
					err := session.Kill()
					if err != nil {
						return evaldo.MakeBuiltinError(ps, "Failed to kill session: "+err.Error(), "tmux-kill-session")
					}
					return arg0
				}
				return evaldo.MakeBuiltinError(ps, "Expected tmux-session object", "tmux-kill-session")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "tmux-kill-session")
			}
		},
	},
}
