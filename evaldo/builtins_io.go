//go:build !b_no_io
// +build !b_no_io

package evaldo

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/refaktor/rye/env"

	"net/http"
	//	"net/http/cgi"

	"github.com/jlaffaye/ftp"
)

func __input(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch str := arg0.(type) {
	case env.String:
		fmt.Print("" + str.Value)
		var input string
		fmt.Scanln(&input)
		fmt.Print(input)
		/* reader := bufio.NewReader(os.Stdin)
		fmt.Print(str)
		inp, _ := reader.ReadString('\n')
		fmt.Println(inp) */
		return *env.NewString(input)
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.StringType}, "__input")
	}
}

func __create(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch s := arg0.(type) {
	case env.Uri:
		// path := strings.Split(s.Path, "://")
		file, err := os.Create(s.Path)
		if err != nil {
			ps.ReturnFlag = true
			ps.FailureFlag = true
			return MakeBuiltinError(ps, err.Error(), "__create")
		}
		return *env.NewNative(ps.Idx, file, "file")
	default:
		ps.ReturnFlag = true
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.UriType}, "__create")
	}
}

func __fs_read(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch f := arg0.(type) {
	case env.Uri:
		data, err := os.ReadFile(f.GetPath())
		if err != nil {
			return MakeBuiltinError(ps, err.Error(), "__fs_read")
		}
		return *env.NewString(string(data))
	default:
		return MakeArgError(ps, 1, []env.Type{env.UriType}, "__fs_read")
	}
	// Read file to byte slice
}

func __fs_read_bytes(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch f := arg0.(type) {
	case env.Uri:
		data, err := os.ReadFile(f.GetPath())
		if err != nil {
			return MakeBuiltinError(ps, err.Error(), "__fs_read_bytes")
		}
		return *env.NewNative(ps.Idx, data, "bytes")
	default:
		return MakeArgError(ps, 1, []env.Type{env.UriType}, "__fs_read_bytes")
	}
	// Read file to byte slice
}

func __fs_read_lines(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch f := arg0.(type) {
	case env.Uri:
		file, err := os.OpenFile(f.GetPath(), os.O_RDONLY, os.ModePerm)
		if err != nil {
			// log.Fatalf("open file error: %v", err)
			return MakeBuiltinError(ps, err.Error(), "__fs_read_lines")
		}
		defer file.Close()

		// var lines []env.Object
		lines := make([]env.Object, 0)
		sc := bufio.NewScanner(file)
		for sc.Scan() {
			lines = append(lines, *env.NewString(sc.Text())) // GET the line string
		}
		if err := sc.Err(); err != nil {
			log.Fatalf("scan file error: %v", err)
			return MakeBuiltinError(ps, err.Error(), "__fs_read_lines")
		}
		return *env.NewBlock(*env.NewTSeries(lines))
	default:
		return MakeArgError(ps, 1, []env.Type{env.UriType}, "__fs_read_lines")
	}
	// Read file to byte slice
}

func __stat(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch r := arg0.(type) {
	case env.Native:
		info, err := r.Value.(*os.File).Stat()
		if err != nil {
			ps.FailureFlag = true
			return MakeBuiltinError(ps, err.Error(), "__stat")
		}
		return *env.NewNative(ps.Idx, info, "file-info")
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__stat")
	}
}

func __https_s_get(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch f := arg0.(type) {
	case env.Uri:
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
		defer cancel()
		proto := ps.Idx.GetWord(f.GetProtocol().Index)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, proto+"://"+f.GetPath(), nil)
		if err != nil {
			ps.FailureFlag = true
			return *env.NewError(err.Error())
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ps.FailureFlag = true
			return *env.NewError(err.Error())
		}
		// Print the HTTP Status Code and Status Name
		//mt.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			return *env.NewString(string(body))
		} else {
			ps.FailureFlag = true
			errMsg := fmt.Sprintf("Status Code: %v, Body: %v", resp.StatusCode, string(body))
			return MakeBuiltinError(ps, errMsg, "__https_s_get")
		}
		// log.Printf("Data read: %s\n", data)
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__https_s_get")
	}
	// Read file to byte slice
}

func __http_s_post(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch f := arg0.(type) {
	case env.Uri:
		switch t := arg2.(type) {
		case env.Word:
			switch d := arg1.(type) {
			case env.String:
				var tt string
				tidx, terr := ps.Idx.GetIndex("json")
				tidx2, terr2 := ps.Idx.GetIndex("text")
				if terr && t.Index == tidx {
					//if t.Value == "json" {
					tt = "application/json"
				} else if terr2 && t.Index == tidx2 {
					tt = "text/plain"
				} else {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Wrong content type.", "__http_s_post")
				}
				// TODO -- add other cases
				// fmt.Println("BEFORE")

				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
				defer cancel()
				proto := ps.Idx.GetWord(f.GetProtocol().Index)
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, proto+"://"+f.GetPath(), bytes.NewBufferString(d.Value))
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}
				req.Header.Set("Content-Type", tt)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}
				// Print the HTTP Status Code and Status Name
				//mt.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))

				// resp, err := http.Post(f.GetProtocol()+"://"+f.GetPath(), tt, bytes.NewBufferString(d.Value))

				// Print the HTTP Status Code and Status Name
				// fmt.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					// fmt.Println("ERR")
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "__http_s_post")
				}

				if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
					return *env.NewString(string(body))
				} else {
					// fmt.Println("ERR33")
					ps.FailureFlag = true
					return env.NewError2(resp.StatusCode, string(body))
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "__http_s_post")
			}
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 3, []env.Type{env.WordType}, "__http_s_post")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.UriType}, "__http_s_post")
	}
	// Read file to byte slice
}

func __email_send(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch to_ := arg0.(type) {
	case env.Email:
		switch msg := arg1.(type) {
		case env.String:
			idx, _ := ps.Idx.GetIndex("user-profile")
			uctx_, _ := ps.Ctx.Get(idx)
			uctx := uctx_.(env.RyeCtx)
			fmt.Println(to_)
			fmt.Println(msg)
			fmt.Println(uctx)
			// TODO continue: uncomment and make it work
			/*
				from, _ := uctx.Get(ps.Idx.GetIndex("smtp-from"))
				password, _ := uctx.Get(ps.Idx.GetIndex("smtp-password"))
				server, _ := uctx.Get(ps.Idx.GetIndex("smtp-server"))
				port, _ := uctx.Get(ps.Idx.GetIndex("smtp-port"))
				// Receiver email address.
				// to := []string{
				//	to_.Value,
				//}
				// Message.
				// message := []byte(msg.Value)
				m := gomail.NewMessage()

				// Set E-Mail sender
				m.SetHeader("From", from)

				// Set E-Mail receivers
				m.SetHeader("To", to_.Address)

				// Set E-Mail subject
				m.SetHeader("Subject", msg.Value)

				// Set E-Mail body. You can set plain text or html with text/html
				m.SetBody("text/plain", msg.Value)

				// Settings for SMTP server
				d := gomail.NewDialer(server, port, from, password)

				// This is only needed when SSL/TLS certificate is not valid on server.
				// In production this should be set to false.
				//			d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

				// Now send E-Mail
				if err := d.DialAndSend(m); err != nil {
					ps.FailureFlag = true
					return env.NewError(err.Error())
				}
			*/
			return *env.NewInteger(1)
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "__email_send")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.EmailType}, "__email_send")
	}
	// Read file to byte slice
}

func __https_s__new_request(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch uri := arg0.(type) {
	case env.Uri:
		switch method := arg1.(type) {
		case env.Word:
			method1 := ps.Idx.GetWord(method.Index)
			if !(method1 == "GET" || method1 == "POST") {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "Wrong method.", "__https_s__new_request")
			}
			switch data := arg2.(type) {
			case env.String:
				data1 := strings.NewReader(data.Value)
				proto := ps.Idx.GetWord(uri.GetProtocol().Index)
				req, err := http.NewRequest(method1, proto+"://"+uri.GetPath(), data1)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "__https_s__new_request")
				}
				return *env.NewNative(ps.Idx, req, "https-request")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "__https_s__new_request")
			}
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.WordType}, "__https_s__new_request")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.UriType}, "__https_s__new_request")
	}
}

func __https_request__set_header(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch req := arg0.(type) {
	case env.Native:
		switch method := arg1.(type) {
		case env.Word:
			name := ps.Idx.GetWord(method.Index)
			switch data := arg2.(type) {
			case env.String:
				req.Value.(*http.Request).Header.Set(name, data.Value)
				return arg0
			default:
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "__https_request__set_header")
			}
		default:
			return MakeArgError(ps, 2, []env.Type{env.WordType}, "__https_request__set_header")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__https_request__set_header")
	}
}

func __https_request__set_basic_auth(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch req := arg0.(type) {
	case env.Native:
		switch username := arg1.(type) {
		case env.String:
			switch password := arg2.(type) {
			case env.String:
				req.Value.(*http.Request).SetBasicAuth(username.Value, password.Value)
				return arg0
			default:
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "__https_request__set_basic_auth")
			}
		default:
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "__https_request__set_basic_auth")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__https_request__set_basic_auth")
	}
}

func __https_request__do(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch req := arg0.(type) {
	case env.Native:
		client := &http.Client{}
		resp, err := client.Do(req.Value.(*http.Request))
		// defer resp.Body.Close() // TODO -- comment this and figure out goling bodyclose
		if err != nil {
			return MakeBuiltinError(ps, err.Error(), "__https_request__do")
		}
		return *env.NewNative(ps.Idx, resp, "https-response")
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__https_request__do")
	}
}

func __https_response__read_body(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch resp := arg0.(type) {
	case env.Native:
		data, err := io.ReadAll(resp.Value.(*http.Response).Body)
		if err != nil {
			return MakeBuiltinError(ps, err.Error(), "__https_response__read_body")
		}
		return *env.NewString(string(data))
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__https_response__read_body")
	}
}

var Builtins_io = map[string]*env.Builtin{

	"input": {
		Argsn: 1,
		Doc:   "Take input from a user.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __input(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	//
	// ##### IO ##### "IO related functions"
	//
	// Tests:
	// equal { open %data/file.txt |type? } 'native
	// equal { open %data/file.txt |kind? } 'file
	"file-schema//open": {
		Argsn: 1,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				file, err := os.Open(s.Path)
				if err != nil {
					return makeError(ps, err.Error())
				}
				return *env.NewNative(ps.Idx, file, "file")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-schema//open")
			}
		},
	},

	// Tests:
	// equal { open\append %data/file.txt |type? } 'native
	// equal { open\append %data/file.txt |kind? } 'writer
	"file-schema//open\\append": {
		Argsn: 1,
		Doc:   "Open file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				file, err := os.OpenFile(s.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "__openFile")
				}
				return *env.NewNative(ps.Idx, file, "writer")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "__openFile")
			}
		},
	},

	// Tests:
	// equal { create %data/created.txt |type? } 'native
	// equal { create %data/created.txt |kind? } 'file
	"file-schema//create": {
		Argsn: 1,
		Doc:   "Create file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				// path := strings.Split(s.Path, "://")
				file, err := os.Create(s.Path)
				if err != nil {
					ps.ReturnFlag = true
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "__create")
				}
				return *env.NewNative(ps.Idx, file, "file")
			default:
				ps.ReturnFlag = true
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "__create")
			}
		},
	},

	// Tests:
	// equal { file-ext? %data/file.txt } ".txt"
	// equal { file-ext? %data/file.temp.png } ".png"
	// equal { file-ext? "data/file.temp.png" } ".png"
	"file-ext?": {
		Argsn: 1,
		Doc:   "Get file extension.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				ext := filepath.Ext(s.Path)
				return *env.NewString(ext)
			case env.String:
				ext := filepath.Ext(s.Value)
				return *env.NewString(ext)
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "file-ext?")
			}
		},
	},

	// should this be generic method or not?
	// Tests:
	// equal { reader %data/file.txt |kind? } 'reader
	// equal { reader open %data/file.txt |kind? } 'reader
	// equal { reader "some string" |kind? } 'reader
	"reader": {
		Argsn: 1,
		Doc:   "Open new reader.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Uri:
				file, err := os.Open(s.Path)
				//trace3(path)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error opening file.", "__open_reader")
				}
				return *env.NewNative(ps.Idx, bufio.NewReader(file), "reader")
			case env.Native:
				file, ok := s.Value.(*os.File)
				if !ok {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error opening file.", "__open_reader")
				}
				return *env.NewNative(ps.Idx, bufio.NewReader(file), "reader")
			case env.String:
				return *env.NewNative(ps.Idx, bufio.NewReader(strings.NewReader(s.Value)), "reader")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "__open_reader")
			}

		},
	},

	"stdin": {
		Argsn: 0,
		Doc:   "Standard input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, os.Stdin, "reader")
		},
	},

	"stdout": {
		Argsn: 0,
		Doc:   "Standard output.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(env1.Idx, os.Stdout, "writer")
		},
	},

	// TODO: add scanner ScanString method ... look at: https://stackoverflow.com/questions/47479564/go-bufio-readstring-in-loop-is-infinite

	// Tests:
	// equal { reader "some\nstring" |read\string "\n" } "some\n"
	"reader//read\\string": {
		Argsn: 2,
		Doc:   "Read string from a reader up to the first character of the ending string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				switch ending := arg1.(type) {
				case env.String:
					// Writer , Reader
					reader, ok := r.Value.(*bufio.Reader)
					if !ok {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Not Reader", "__read\\string")
					}
					inp, err := reader.ReadString(ending.Value[0])
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "__read\\string")
					}
					return *env.NewString(inp)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "__read\\string")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__read\\string")
			}

		},
	},

	"reader//copy": {
		Argsn: 2,
		Doc:   "Copy from a reader to a writer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				switch w := arg1.(type) {
				case env.Native:
					// Writer , Reader
					_, err := io.Copy(w.Value.(io.Writer), r.Value.(io.Reader))
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "__copy")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "__copy")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__copy")
			}

		},
	},

	// We have duplication reader file TODO think about this ... is it worth
	// changing how kinds work, making them more complex? not sure yet
	"file//copy": {
		Argsn: 2,
		Doc:   "Copy Rye file to ouptut.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch r := arg0.(type) {
			case env.Native:
				switch w := arg1.(type) {
				case env.Native:
					// Writer , Reader
					_, err := io.Copy(w.Value.(io.Writer), r.Value.(io.Reader))
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "__copy")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "__copy")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__copy")
			}

		},
	},

	// Tests:
	// equal { stat open %data/file.txt |kind? } 'file-info
	"file//stat": {
		Argsn: 1,
		Doc:   "Get stat of a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __stat(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Tests:
	// equal { size? stat open %data/file.txt } 16
	"file-info//size?": {
		Argsn: 1,
		Doc:   "Get size of a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				size := s.Value.(os.FileInfo).Size()
				return *env.NewInteger(size)
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "file-info//size?")
			}
		},
	},

	// Tests:
	// equal { read-all open %data/file.txt } "hello text file\n"
	"file//read-all": {
		Argsn: 1,
		Doc:   "Read all file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				data, err := io.ReadAll(s.Value.(io.Reader))
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error reading file.", "__read_all")
				}
				return *env.NewString(string(data))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__read_all")
			}
		},
	},

	"file//seek\\end": {
		Argsn: 1,
		Doc:   "Write to a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				reader, ok := s.Value.(*os.File)
				if !ok {
					return MakeBuiltinError(ps, "Native not io.Reader", "file//seek\\end")
				}
				reader.Seek(0, os.SEEK_END)
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "file//seek\\end")
			}
		},
	},

	// Tests:
	// equal { close open %data/file.txt } ""
	"file//close": {
		Argsn: 1,
		Doc:   "Closes an open file or reader or writer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Native:
				err := s.Value.(*os.File).Close()
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "__close")
				}
				return *env.NewString("")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "__close")
			}

		},
	},

	// Tests:
	// equal { read %data/file.txt } "hello text file\n"
	"file-schema//read": {
		Argsn: 1,
		Doc:   "Read a file given the path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __fs_read(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Tests:
	// equal { read %data/file.txt } "hello text file\n"
	"file-schema//read\\bytes": {
		Argsn: 1,
		Doc:   "Read a specific number of bytes from a file path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __fs_read_bytes(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Tests:
	// equal { read %data/file.txt } "hello text file\n"
	"file-schema//read\\lines": {
		Argsn: 1,
		Doc:   "Read files into the block of lines.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __fs_read_lines(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Tests:
	// equal { write %data/write.txt "written\n" } "written\n"
	"file-schema//write": {
		Argsn: 2,
		Doc:   "Write to a file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				switch s := arg1.(type) {
				case env.String:
					err := os.WriteFile(f.GetPath(), []byte(s.Value), 0600)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "__fs_write")
					}
					return arg1
				case env.Native:
					err := os.WriteFile(f.GetPath(), s.Value.([]byte), 0600)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "__fs_write")
					}
					return arg1
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.NativeType}, "__fs_write")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "__fs_write")
			}

		},
	},

	"writer//write\\string": {
		Argsn: 2,
		Doc:   "Write string to a writer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg1.(type) {
			case env.String:
				switch ww := arg0.(type) {
				case env.Native:
					writer, ok := ww.Value.(*os.File)
					if !ok {
						return MakeBuiltinError(ps, "Native not io.File", "writer//write\\string")
					}
					_, err := writer.WriteString(s.Value)
					if err != nil {
						return MakeBuiltinError(ps, "Error at write: "+err.Error(), "writer//write\\string")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "writer//write\\string")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "writer//write\\string")
			}

		},
	},

	/*
		"file-schema//open": {
			Argsn: 1,
			Doc:   "Open a file, get a reader",
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch f := arg0.(type) {
				case env.Uri:
					file, err := os.Open(s.Path)
					//trace3(path)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Error opening file.", "file-schema//open")
					}
					return *env.NewNative(ps.Idx, bufio.NewReader(file), "file-schema//open")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "file-schema//open")
				}
			},
		}, */

	"https-schema//open": {
		Argsn: 1,
		Doc:   "Open a HTTPS GET request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch f := arg0.(type) {
			case env.Uri:
				// ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
				// defer cancel()
				proto := ps.Idx.GetWord(f.GetProtocol().Index)
				// req, err := http.NewRequestWithContext(ctx, http.MethodGet, proto+"://"+f.GetPath(), nil)
				req, err := http.NewRequest(http.MethodGet, proto+"://"+f.GetPath(), nil)
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}
				// Print the HTTP Status Code and Status Name
				//mt.Println("HTTP Response Status:", resp.StatusCode, http.StatusText(resp.StatusCode))
				// defer resp.Body.Close()
				// body, _ := io.ReadAll(resp.Body)

				if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
					return *env.NewNative(ps.Idx, resp.Body, "https-schema://open")
				} else {
					ps.FailureFlag = true
					errMsg := fmt.Sprintf("Status Code: %v, Body: %v", resp.StatusCode)
					return MakeBuiltinError(ps, errMsg, "https-schema://open")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "https-schema://open")
			}
		},
	},

	"https-schema//get": {
		Argsn: 1,
		Doc:   "Make a HTTPS GET request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __https_s_get(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"https-schema//post": {
		Argsn: 3,
		Doc:   "Make a HTTPS POST request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __http_s_post(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"http-schema//get": {
		Argsn: 1,
		Doc:   "Make a HTTP GET request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __https_s_get(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"http-schema//post": {
		Argsn: 3,
		Doc:   "Make a HTTP POST request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __http_s_post(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"https-schema//new-request": {
		Argsn: 3,
		Doc:   "Create a new HTTPS Request object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __https_s__new_request(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"https-request//set-header": {
		Argsn: 3,
		Doc:   "Set header to the HTTPS Request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __https_request__set_header(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"https-request//set-basic-auth": {
		Argsn: 3,
		Doc:   "Set Basic Auth to the HTTPS Request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __https_request__set_basic_auth(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"https-request//call": {
		Argsn: 1,
		Doc:   "Call a HTTPS Request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __https_request__do(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"https-response//read-body": {
		Argsn: 1,
		Doc:   "Read body of HTTPS response.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __https_response__read_body(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"email//send": {
		Argsn: 2,
		Doc:   "Send email.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __email_send(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	"ftp-schema//open": {
		Argsn: 1,
		Doc:   "Open connection to FTP Server",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Uri:
				conn, err := ftp.Dial(s.Path)
				if err != nil {
					fmt.Println("Error connecting to FTP server:", err)
					return MakeBuiltinError(ps, "Error connecting to FTP server: "+err.Error(), "ftp-schema//open")
				}
				//trace3(path)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Error opening file.", "ftp-schema//open")
				}
				return *env.NewNative(ps.Idx, conn, "ftp-connection")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "ftp-schema//open")
			}
		},
	},

	"ftp-connection//login": {
		Argsn: 3,
		Doc:   "Login to connection to FTP Server",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Native:
				username, ok := arg1.(env.String)
				if !ok {
					// TODO ARG ERROR
					return nil
				}
				pwd, ok := arg2.(env.String)
				if !ok {
					// TODO ARG ERROR
					return nil
				}
				err := s.Value.(*ftp.ServerConn).Login(username.Value, pwd.Value)
				if err != nil {
					// TODO
					fmt.Println("Error logging in:", err)
					return nil
				}
				return s
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "ftp-connection//login")
			}
		},
	},

	"ftp-connection//retrieve": {
		Argsn: 2,
		Doc:   "Retrieve file from connection to FTP Server",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			switch s := arg0.(type) {
			case env.Native:
				path, ok := arg1.(env.String)
				if !ok {
					// TODO ARG ERROR
				}
				resp, err := s.Value.(*ftp.ServerConn).Retr(path.Value)
				if err != nil {
					fmt.Println("Error retrieving:", err)
					return nil
				}
				return *env.NewNative(ps.Idx, resp, "reader")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType, env.StringType}, "ftp-connection//login")
			}
		},
	},
}
