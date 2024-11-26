//go:build !no_os && !b_wasm

package evaldo

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/refaktor/rye/env"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// In request we return a raw-map, because it's very inside loop call, this is sparse call, and we get tons of fields, so it would be best
// to turn them to normal Rye map (which is now Env / later Context or something like it), and they query it from Rye.

func FileExists(filePath string) int {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0 // fmt.Println("File does not exist")
		} else {
			return -1 // fmt.Println("Error checking file:", err)
		}
	} else {
		return 1
	}
}

var Builtins_os = map[string]*env.Builtin{

	"cwd?": {
		Argsn: 0,
		Doc:   "Returns current working directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			path, err := os.Getwd()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "cwd")
			}
			return *env.NewUri1(ps.Idx, "file://"+path)
		},
	},

	"cd": {
		Argsn: 1,
		Doc:   "Changes current directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:

				err := os.Chdir(path.GetPath())
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "cd")
				}
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "cd")
			}
		},
	},

	"env?": {
		Argsn: 1,
		Doc:   "Gets the environment variable.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch variable_name := arg0.(type) {
			case env.String:

				val, ok := os.LookupEnv(variable_name.Value)
				if !ok {
					return MakeBuiltinError(ps, "Variable couldn't be read", "env?")
				}
				return env.NewString(val)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "env?")
			}
		},
	},

	/*	"cd_": {
		Argsn: 1,
		Doc:   "Changes current directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				new := filepath.Join(filepath.Dir(ps.WorkingPath), path.GetPath())
				res := FileExists(new)
				if res == 1 {
					ps.WorkingPath = filepath.Join(filepath.Dir(ps.WorkingPath), path.GetPath())
					return arg0
				} else if res == 0 {
					return MakeBuiltinError(ps, "Path doesn't exist", "cd")
				} else {
					return MakeBuiltinError(ps, "Error determining if path exists", "cd")
				}
				// TODO -- check if it exists
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "cd")
			}
		},
	}, */

	"mkdir": {
		Argsn: 1,
		Doc:   "Creates a directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				newDir := filepath.Join(filepath.Dir(ps.WorkingPath), path.GetPath())
				err := os.Mkdir(newDir, 0755) // Create directory with permissions 0755
				if err != nil {
					return MakeBuiltinError(ps, "Error creating directory: "+err.Error(), "mkdir")
				} else {
					return arg0
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "mkdir")
			}
		},
	},

	"mktmp": {
		Argsn: 0,
		Doc:   "Creates a temporary directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			dir, err := os.MkdirTemp("", "rye-tmp-")
			if err != nil {
				return MakeBuiltinError(ps, "Error creating temporary directory: "+err.Error(), "mktmp")
			}
			return *env.NewUri1(ps.Idx, "file://"+dir)
		},
	},

	"mv": {
		Argsn: 2,
		Doc:   "Creates a directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Uri:
				switch path2 := arg1.(type) {
				case env.Uri:
					old := filepath.Join(filepath.Dir(ps.WorkingPath), path.GetPath())
					new := filepath.Join(filepath.Dir(ps.WorkingPath), path2.GetPath())
					err := os.Rename(old, new)
					if err != nil {
						fmt.Println("Error renaming file:", err)
						return MakeBuiltinError(ps, "Error renaming file: "+err.Error(), "mv")
					} else {
						return arg1
					}
				default:
					return MakeArgError(ps, 1, []env.Type{env.UriType}, "mv")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "mv")
			}
		},
	},

	/*	"cwd_": {
		Argsn: 0,
		Doc:   "Returns current working directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewUri1(ps.Idx, "file://"+ps.WorkingPath)
		},
	}, */

	"ls": {
		Argsn: 0,
		Doc:   "Returns current working directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			files, err := os.ReadDir(".")
			if err != nil {
				return MakeBuiltinError(ps, "Error reading directory:"+err.Error(), "ls")
			}

			items := make([]env.Object, len(files))

			for i, file := range files {
				// fmt.Println(file.Name()) // Print only file/directory names

				items[i] = *env.NewUri1(ps.Idx, "file://"+file.Name())
			}
			return *env.NewBlock(*env.NewTSeries(items))
		},
	},

	"host-info?": {
		Argsn: 0,
		Doc:   "Get information about the host system.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			v, err := host.Info()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "host-info?")
			}
			r := env.NewDict(make(map[string]any, 10))
			r.Data["hostname"] = *env.NewString(v.Hostname)
			r.Data["uptime"] = *env.NewInteger(int64(v.Uptime))
			r.Data["boot-time"] = *env.NewInteger(int64(v.BootTime))
			r.Data["procs"] = *env.NewInteger(int64(v.Procs))
			r.Data["os"] = *env.NewString(v.OS)
			r.Data["platform"] = *env.NewString(v.Platform)
			r.Data["platform-family"] = *env.NewString(v.PlatformFamily)
			r.Data["platform-version"] = *env.NewString(v.PlatformVersion)
			r.Data["kernel-version"] = *env.NewString(v.KernelVersion)
			r.Data["virtualization-system"] = *env.NewString(v.VirtualizationSystem)
			return *r
		},
	},
	"users?": {
		Argsn: 0,
		Doc:   "Get information about users as a spreadsheet.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			users, err := host.Users()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "users?")
			}
			fmt.Println(users)
			s := env.NewSpreadsheet([]string{"User", "Terminal", "Host", "Started"})
			for _, user := range users {
				vals := []any{
					*env.NewString(user.User),
					*env.NewString(user.Terminal),
					*env.NewString(user.Host),
					*env.NewInteger(int64(user.Started)),
				}
				s.AddRow(*env.NewSpreadsheetRow(vals, s))
			}
			return *s
		},
	},
	"load-avg?": {
		Argsn: 0,
		Doc:   "Get the load average as a dict representing load average over the last 1, 5, and 15 minutes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			v, err := load.Avg()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "load-avg?")
			}
			r := env.NewDict(make(map[string]any, 3))
			r.Data["1"] = *env.NewDecimal(v.Load1)
			r.Data["5"] = *env.NewDecimal(v.Load5)
			r.Data["15"] = *env.NewDecimal(v.Load15)
			return *r
		},
	},
	"virtual-memory?": {
		Argsn: 0,
		Doc:   "Get information about virtual memory usage.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			v, err := mem.VirtualMemory()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "virtual-memory?")
			}
			r := env.NewDict(make(map[string]any, 3))
			r.Data["total"] = *env.NewInteger(int64(v.Total))
			r.Data["free"] = *env.NewInteger(int64(v.Free))
			r.Data["used-percent"] = *env.NewDecimal(v.UsedPercent)
			return *r
		},
	},
	"disk-usage?": {
		Argsn: 0,
		Doc:   "Get disk usage information as a spreadsheet.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			partitions, err := disk.Partitions(true)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "disk-usage?")
			}
			s := env.NewSpreadsheet([]string{"Filesystem", "Size", "Used", "Available", "Capacity", "iused", "ifree", "%iused", "Mounted on"})
			for _, partition := range partitions {
				usage, err := disk.Usage(partition.Mountpoint)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "disk-usage?")
				}
				vals := []any{
					*env.NewString(partition.Device),
					*env.NewInteger(int64(usage.Total)),
					*env.NewInteger(int64(usage.Used)),
					*env.NewInteger(int64(usage.Free)),
					*env.NewDecimal(usage.UsedPercent),
					*env.NewInteger(int64(usage.InodesUsed)),
					*env.NewInteger(int64(usage.InodesFree)),
					*env.NewInteger(int64(usage.InodesUsedPercent)),
					*env.NewString(usage.Path),
				}
				s.AddRow(*env.NewSpreadsheetRow(vals, s))
			}
			return *s
		},
	},
	"pids?": {
		Argsn: 0,
		Doc:   "Get process pids as a block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			pids, err := process.Pids()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "pids?")
			}

			pids2 := make([]env.Object, len(pids))
			for i, p := range pids {
				pids2[i] = env.NewInteger(int64(p))
			}
			return *env.NewBlock(*env.NewTSeries(pids2))
		},
	},
	"processes?": {
		Argsn: 0,
		Doc:   "Get information about all processes as a spreadsheet.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			processes, err := process.Processes()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "processes?")
			}
			s := proccesSpreadsheetBase()
			for _, process := range processes {
				processSpreadsheetAdd(s, process)
			}
			return *s
		},
	},
	"process": {
		Argsn: 1,
		Doc:   "Get information about process with a given PID.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pid := arg0.(type) {
			case env.Integer:
				process, err := process.NewProcess(int32(pid.Value))
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "process")
				}
				s := proccesSpreadsheetBase()
				processSpreadsheetAdd(s, process)
				return s.Rows[0].ToDict()
			default:
				return *MakeArgError(ps, 1, []env.Type{env.IntegerType}, "process")
			}
		},
	},

	"lookup-address": {
		Argsn: 1,
		Doc:   "Get address of an IP.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ip := arg0.(type) {
			case env.String:
				names, err := net.LookupAddr(ip.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "ip-lookup")
				}

				items := make([]env.Object, len(names))

				for i, name := range names {
					items[i] = *env.NewString(name)
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return *MakeArgError(ps, 1, []env.Type{env.StringType}, "ip-lookup")
			}
		},
	},

	"lookup-ip": {
		Argsn: 1,
		Doc:   "Get IP of an address.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ip := arg0.(type) {
			case env.String:
				names, err := net.LookupIP(ip.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "ip-lookup")
				}

				items := make([]env.Object, len(names))

				for i, name := range names {
					items[i] = *env.NewString(name.String())
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return *MakeArgError(ps, 1, []env.Type{env.StringType}, "ip-lookup")
			}
		},
	},

	"write\\clipboard": {
		Argsn: 1,
		Doc:   "Writes value to OS clipboard",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.String:
				err := clipboard.WriteAll(val.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "write\\clipboard")
				}
				return arg0
			default:
				return *MakeArgError(ps, 1, []env.Type{env.StringType}, "write\\clipboard")
			}
		},
	},

	"read\\clipboard": {
		Argsn: 0,
		Doc:   "Reads value from OS clipboard",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			val, err := clipboard.ReadAll()
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "read\\clipboard")
			}
			return *env.NewString(val)
		},
	},
}

func proccesSpreadsheetBase() *env.Spreadsheet {
	return env.NewSpreadsheet([]string{
		"User",
		"PID",
		"Status",
		"%CPU",
		"%MEM",
		"VZS",
		"RSS",
		"Num Threads",
		"Num FDs",
		"Num Open Files",
		"Num Connections",
		"Started at",
		"CPU Time",
		"Command",
	})
}

func processSpreadsheetAdd(s *env.Spreadsheet, process *process.Process) {
	var status env.String
	stat, err := process.Status()
	if err == nil {
		status = *env.NewString(strings.Join(stat, " "))
	} else {
		status = *env.NewString("???")
	}

	var vzs env.Object
	var rss env.Object
	memInfo, err := process.MemoryInfo()
	if err == nil {
		vzs = *env.NewInteger(int64(memInfo.VMS))
		rss = *env.NewInteger(int64(memInfo.RSS))
	} else {
		vzs = *env.NewString("???")
		rss = *env.NewString("???")
	}

	var numOpenFiles env.Object
	openFiles, err := process.OpenFiles()
	if err == nil {
		numOpenFiles = *env.NewInteger(int64(len(openFiles)))
	} else {
		numOpenFiles = *env.NewString("???")
	}

	var numConnections env.Object
	connections, err := process.Connections()
	if err == nil {
		numConnections = *env.NewInteger(int64(len(connections)))
	} else {
		numConnections = *env.NewString("???")
	}

	var startedAt env.Object
	createTime, err := process.CreateTime()
	if err == nil {
		startedAt = *env.NewDate(time.UnixMilli(createTime))
	} else {
		startedAt = *env.NewString("???")
	}

	var cpuTime env.Object
	times, err := process.Times()
	if err == nil {
		dur := time.Duration(times.User+times.System) * time.Second
		cpuTime = *env.NewString(fmt.Sprintf("%02d:%02d.%02d", int(dur.Minutes()), int(dur.Seconds())%60, int(dur.Milliseconds())%1000))
	} else {
		cpuTime = *env.NewString("???")
	}

	vals := []any{
		maybeString(process.Username),
		process.Pid,
		status,
		maybeFloat64(process.CPUPercent),
		maybeFloat32(process.MemoryPercent),
		vzs,
		rss,
		maybeInt32(process.NumThreads),
		maybeInt32(process.NumFDs),
		numOpenFiles,
		numConnections,
		startedAt,
		cpuTime,
		maybeString(process.Cmdline),
	}
	s.AddRow(*env.NewSpreadsheetRow(vals, s))
}

func maybeString(f func() (string, error)) env.Object {
	s, err := f()
	if err != nil {
		return *env.NewString("???")
	}
	return *env.NewString(s)
}

func maybeFloat64(f func() (float64, error)) env.Object {
	s, err := f()
	if err != nil {
		return *env.NewString("???")
	}
	return *env.NewDecimal(s)
}

func maybeFloat32(f func() (float32, error)) env.Object {
	s, err := f()
	if err != nil {
		return *env.NewString("???")
	}
	return *env.NewDecimal(float64(s))
}

func maybeInt32(f func() (int32, error)) env.Object {
	s, err := f()
	if err != nil {
		return *env.NewString("???")
	}
	return *env.NewInteger(int64(s))
}
