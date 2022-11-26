// go build -ldflags "-s -w" -o index.cgi cgi.go

package main

import (
	"fmt"
	"net/http"
	"net/http/cgi"
)

func main() {
	if err := cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, "Method:", r.Method, "<br/>")
		fmt.Fprintln(w, "URL:", r.URL.String(), "<br/>")
		query := r.URL.Query()
		for k := range query {
			fmt.Fprintln(w, "Query", k+":", query.Get(k), "<br/>")
		}
		r.ParseForm()
		form := r.Form
		for k := range form {
			fmt.Fprintln(w, "Form", k+":", form.Get(k), "<br/>")
		}
		post := r.PostForm
		for k := range post {
			fmt.Fprintln(w, "PostForm", k+":", post.Get(k), "<br/>")
		}
		fmt.Fprintln(w, "RemoteAddr:", r.RemoteAddr, "<br/>")
		if referer := r.Referer(); len(referer) > 0 {
			fmt.Fprintln(w, "Referer:", referer, "<br/>")
		}
		if ua := r.UserAgent(); len(ua) > 0 {
			fmt.Fprintln(w, "UserAgent:", ua, "<br/>")
		}
		for _, cookie := range r.Cookies() {
			fmt.Fprintln(w, "Cookie", cookie.Name+":", cookie.Value, cookie.Path, cookie.Domain, cookie.RawExpires, "<br/>")
		}

		fmt.Fprintln(w, "<!DOCTYPE HTML><html><h1>This output is from a go program over ssl !</h1></html>")
	})); err != nil {
		fmt.Println(err)
	}
}
