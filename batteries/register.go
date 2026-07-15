package batteries

import (
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// registerHooks wires battery functions into the hook variables in evaldo,
// so that base evaldo code can call battery functionality without an import cycle.
func registerHooks() {
	evaldo.BatteryValidateHook = func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) env.Object {
		return BuiValidate(ps, arg0, arg1)
	}
	evaldo.BatteryConvertHook = func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) env.Object {
		return BuiConvert(ps, arg0, arg1)
	}
	evaldo.BatteryMarkdownDisplayHook = func(source string) []interface{} {
		return convertMarkdownDisplayItems(markdownDisplayItems(source))
	}
	evaldo.BatteryEyrEvalBlockInsideHook = func(ps *env.ProgramState, inj env.Object, injnow bool) {
		Eyr_EvalBlockInside(ps, inj, injnow)
	}
	evaldo.BatteryEyrEvalBlockHook = func(ps *env.ProgramState, full bool) {
		Eyr_EvalBlock(ps, full)
	}
	evaldo.BatteryDialectMathHook = func(ps *env.ProgramState, arg0 env.Object) env.Object {
		return DialectMath(ps, arg0)
	}
	evaldo.BatteryEvalInPersistentCtxHook = func(ps *env.ProgramState, ctx evaldo.PersistentCtxInterface) {
		if pctx, ok := ctx.(*PersistentCtx); ok {
			EvalBlockInPersistentCtx(ps, pctx)
		}
	}
}

// RegisterBatteries registers all battery (non-base) builtins into the program state.
// Call this after evaldo.RegisterBuiltins to get the full standard set of Rye builtins.
func RegisterBatteries(ps *env.ProgramState) {
	registerHooks()
	evaldo.RegisterBuiltins2(Builtins_error_creation, ps, "error-creation")
	evaldo.RegisterBuiltins2(Builtins_error_inspection, ps, "error-inspection")
	evaldo.RegisterBuiltins2(Builtins_error_handling, ps, "error-handling")
	evaldo.RegisterBuiltins2(Builtins_match, ps, "match")
	evaldo.RegisterBuiltins2(Builtins_table, ps, "table")
	evaldo.RegisterBuiltins2(Builtins_vector, ps, "vector")
	evaldo.RegisterBuiltins2(Builtins_matrix, ps, "matrix")
	evaldo.RegisterBuiltins2(Builtins_io, ps, "io")
	evaldo.RegisterBuiltins2(Builtins_cmd, ps, "cmd")
	evaldo.RegisterBuiltins2(Builtins_regexp, ps, "regexp")
	evaldo.RegisterBuiltins2(Builtins_validation, ps, "validation")
	evaldo.RegisterBuiltins2(Builtins_cli, ps, "cli")
	evaldo.RegisterBuiltins2(Builtins_conversion, ps, "conversion")
	evaldo.RegisterBuiltins2(Builtins_web, ps, "web")
	evaldo.RegisterBuiltins2(Builtins_markdown, ps, "markdown")
	evaldo.RegisterBuiltins2(Builtins_sxml, ps, "sxml")
	evaldo.RegisterBuiltins2(Builtins_html, ps, "html")
	evaldo.RegisterBuiltins2(Builtins_json, ps, "json")
	evaldo.RegisterBuiltins2(Builtins_bson, ps, "bson")
	evaldo.RegisterBuiltins2(Builtins_stackless, ps, "stackless")
	evaldo.RegisterBuiltins2(Builtins_eyr, ps, "eyr")
	evaldo.RegisterBuiltins2(Builtins_goroutines, ps, "goroutines")
	evaldo.RegisterBuiltins2(Builtins_msgdispatcher, ps, "msgdispatcher")
	evaldo.RegisterBuiltins2(Builtins_http, ps, "http")
	evaldo.RegisterBuiltins2(Builtins_sqlite, ps, "sqlite")
	evaldo.RegisterBuiltins2(Builtins_psql, ps, "psql")
	evaldo.RegisterBuiltins2(Builtins_mysql, ps, "mysql")
	evaldo.RegisterBuiltins2(Builtins_email, ps, "email")
	evaldo.RegisterBuiltins2(Builtins_imap, ps, "imap")
	evaldo.RegisterBuiltins2(Builtins_structures, ps, "structs")
	evaldo.RegisterBuiltins2(Builtins_smtpd, ps, "smtpd")
	evaldo.RegisterBuiltins2(Builtins_mail, ps, "mail")
	evaldo.RegisterBuiltins2(Builtins_ssh, ps, "ssh")
	evaldo.RegisterBuiltins2(Builtins_bcrypt, ps, "bcrypt")
	evaldo.RegisterBuiltins2(Builtins_console, ps, "console")
	evaldo.RegisterBuiltinsInContext(Builtins_crypto, ps, "crypto")
	evaldo.RegisterBuiltinsInContext(Builtins_encoding, ps, "encoding")
	evaldo.RegisterBuiltinsInContext(Builtins_math, ps, "math")
	evaldo.RegisterBuiltinsInContext(Builtins_os, ps, "os")
	evaldo.RegisterBuiltinsInContext(Builtins_pipes, ps, "pipes")
	evaldo.RegisterBuiltinsInContext(Builtins_term, ps, "term")
	evaldo.RegisterBuiltinsInContext(Builtins_termstr, ps, "termstr")
	evaldo.RegisterBuiltinsInContext(Builtins_tui, ps, "tui")
	evaldo.RegisterBuiltinsInContext(Builtins_telegrambot, ps, "telegram")
	evaldo.RegisterBuiltinsInContext(Builtins_mcp, ps, "mcp")
	evaldo.RegisterBuiltins2(Builtins_mqtt, ps, "mqtt")
	evaldo.RegisterBuiltins2(Builtins_chitosocket, ps, "chitosocket")
	evaldo.RegisterBuiltins2(builtins_trees, ps, "trees")
	// evaldo.RegisterBuiltinsInContext(Builtins_git, ps, "git")
	// temporarily removed: evaldo.RegisterBuiltinsInContext(Builtins_docker, ps, "docker")
	evaldo.RegisterBuiltinsInContext(Builtins_prometheus, ps, "prometheus")
	evaldo.RegisterBuiltinsInContext(Builtins_echarts, ps, "echarts")
	evaldo.RegisterBuiltinsInContext(Builtins_flui, ps, "flui")
	evaldo.RegisterBuiltins2(Builtins_js_interop, ps, "jsinterop")
	// evaldo.RegisterBuiltinsInContext(Builtins_flui_v2, ps, "flui2")
	// ## Archived / contrib modules (not included in batteries):
	// evaldo.RegisterBuiltins2(Builtins_gtk, ps, "gtk")
	// evaldo.RegisterBuiltins2(Builtins_nats, ps, "nats")
	// evaldo.RegisterBuiltins2(Builtins_qframe, ps, "qframe")
	// evaldo.RegisterBuiltins2(Builtins_raylib, ps, "raylib")
}
