package evaldo

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"rye/env"
	"strings"
	//"strconv"
)

func __input(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch str := arg0.(type) {
	case env.String:
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(str)
		inp, _ := reader.ReadString('\n')
		fmt.Println(inp)
		return env.String{inp}
	default:
		//env1.ReturnFlag = true
		env1.FailureFlag = true
		return env.NewError("arg 1 should be string")
	}
}

func __open(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch s := arg0.(type) {
	case env.Uri:
		path := strings.Split(s.Path, "://")
		file, err := os.Open(path[1])
		if err != nil {
			//env1.ReturnFlag = true
			env1.FailureFlag = true
			return *env.NewError(err.Error())
		}
		return *env.NewNative(env1.Idx, file, "rye-file")
	default:
		//env1.ReturnFlag = true
		env1.FailureFlag = true
		return *env.NewError("just accepting Uri-s")
	}
}

func __create(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch s := arg0.(type) {
	case env.Uri:
		path := strings.Split(s.Path, "://")
		file, err := os.Create(path[1])
		if err != nil {
			env1.ReturnFlag = true
			env1.FailureFlag = true
			return *env.NewError(err.Error())
		}
		return *env.NewNative(env1.Idx, file, "rye-file")
	default:
		env1.ReturnFlag = true
		env1.FailureFlag = true
		return *env.NewError("just accepting Uri-s")
	}
}

func __open_reader(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch s := arg0.(type) {
	case env.Uri:
		path := strings.Split(s.Path, "://")
		file, err := os.Open(path[1])
		//trace3(path)
		if err != nil {
			env1.FailureFlag = true
			return *env.NewError("Error opening file")
		}
		return *env.NewNative(env1.Idx, bufio.NewReader(file), "rye-reader")
	default:
		env1.FailureFlag = true
		return *env.NewError("just accepting Uri-s")
	}
}

func __read_all(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch s := arg0.(type) {
	case env.Native:
		data, err := ioutil.ReadAll(s.Value.(io.Reader))
		if err != nil {
			env1.FailureFlag = true
			return *env.NewError("Error reading file")
		}
		return env.String{string(data)}
	}
	env1.FailureFlag = true
	return *env.NewError("Failed")
}

func __close(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch s := arg0.(type) {
	case env.Native:
		err := s.Value.(*os.File).Close()
		if err != nil {
			env1.FailureFlag = true
			return *env.NewError(err.Error())
		}
		return env.String{""}
	}
	env1.FailureFlag = true
	return *env.NewError("Failed")
}

func __write(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch f := arg0.(type) {
	case env.Native:
		switch s := arg1.(type) {
		case env.String:

			bytesWritten, err := f.Value.(io.Writer).Write([]byte(s.Value))
			if err != nil {
				env1.FailureFlag = true
				return *env.NewError(err.Error())
			}
			return env.Integer{int64(bytesWritten)}
			//log.Printf("Wrote %d bytes.\n", bytesWritten)
		}
	}
	env1.FailureFlag = true
	return *env.NewError("Failed")
}

var Builtins_io = map[string]*env.Builtin{

	"input": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __input(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"file-schema//open": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __open(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"file-schema//create": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __create(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"file-schema//open-reader": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __open_reader(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"rye-file//read-all": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __read_all(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"rye-file//write": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __write(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"rye-file//close": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __close(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},
}
