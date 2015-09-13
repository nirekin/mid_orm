package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func readFormInt(n string) func(r *http.Request) (int, error) {
	return func(r *http.Request) (int, error) {
		if i, err := strconv.Atoi(r.FormValue(n)); err == nil {
			return i, nil
		} else {
			fmt.Printf("err %v\n", err)
			return 0, err
		}
	}
}

func readFormString(n string) func(r *http.Request) string {
	return func(r *http.Request) string {
		return r.FormValue(n)
	}
}

func readIntUrl(r *http.Request) (int, error) {
	s := strings.Split(r.URL.Path, "/")
	if i, err := strconv.Atoi(s[len(s)-1]); err == nil {
		return i, nil
	} else {
		return 0, err
	}
}

func eqstring(s1 string, s2 string) (v bool) {
	if s1 == s2 {
		v = true
		return
	}
	v = false
	return
}

func encodeUTF(s string) string {
	// TODO check to implement a real encoding
	s = strings.Replace(s, "!", "U+0021", -1)
	s = strings.Replace(s, "\"", "U+0022", -1)
	s = strings.Replace(s, "#", "U+0023", -1)
	s = strings.Replace(s, "&", "U+0024", -1)
	s = strings.Replace(s, "%", "U+0025", -1)
	s = strings.Replace(s, "&", "U+0026", -1)
	s = strings.Replace(s, "'", "U+0027", -1)
	s = strings.Replace(s, "(", "U+0028", -1)
	s = strings.Replace(s, ")", "U+0029", -1)
	s = strings.Replace(s, "*", "U+002A", -1)
	s = strings.Replace(s, "+", "U+002B", -1)
	s = strings.Replace(s, ",", "U+002C", -1)
	s = strings.Replace(s, "-", "U+002D", -1)
	s = strings.Replace(s, "/", "U+002F", -1)
	return s
}
