package evaldo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"rye/env"
	"strings"

	"net/http"
	// "net/url"
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

func __fs_read(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch f := arg0.(type) {
	case env.Uri:

		data, err := ioutil.ReadFile(f.GetPath())
		if err != nil {
			env1.FailureFlag = true
			return *env.NewError(err.Error())
		}

		// log.Printf("Data read: %s\n", data)
		return env.String{string(data)}
	}
	env1.FailureFlag = true
	return *env.NewError("Failed")

	// Read file to byte slice
}

func __https_s_get(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch f := arg0.(type) {
	case env.Uri:

		resp, err := http.Get("https://" + f.GetPath())
		if err != nil {
			env1.FailureFlag = true
			return *env.NewError(err.Error())
		}

		// Print the HTTP Status Code and Status Name
		//mt.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			return env.String{string(body)}
		} else {
			env1.FailureFlag = true
			return env.NewError2(resp.StatusCode, string(body))
		}

		// log.Printf("Data read: %s\n", data)

	}
	env1.FailureFlag = true
	return *env.NewError("Failed")

	// Read file to byte slice
}

func __http_s_post(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch f := arg0.(type) {
	case env.Uri:

		switch t := arg1.(type) {
		case env.Tagword:
			switch d := arg2.(type) {
			case env.String:

				var tt string
				tidx, terr := env1.Idx.GetIndex("json")
				if terr && t.Index == tidx {
					//if t.Value == "json" {
					tt = "application/json"
				} else {
					env1.FailureFlag = true
					return *env.NewError("wrong content type")
				}
				// TODO -- add other cases

				resp, err := http.Post("https://"+f.GetPath(), tt, bytes.NewBufferString(d.Value))
				if err != nil {
					env1.FailureFlag = true
					return *env.NewError(err.Error())
				}

				// Print the HTTP Status Code and Status Name
				// fmt.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)

				if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
					return env.String{string(body)}
				} else {
					env1.FailureFlag = true
					return env.NewError2(resp.StatusCode, string(body))
				}
			}
		}
	}
	env1.FailureFlag = true
	return *env.NewError("Failed")

	// Read file to byte slice
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

	"file-schema//read": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __fs_read(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"https-schema//get": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __https_s_get(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"https-schema//post": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __http_s_post(env1, arg0, arg1, arg2, arg3, arg4)
		},
	},
}
