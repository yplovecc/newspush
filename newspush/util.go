package main

import (
	"encoding/json"
	"net/http"
)

func mustEncode(w http.ResponseWriter, i interface{}) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-type", "application/json;charset=utf-8")
	e := json.NewEncoder(w)
	if err := e.Encode(i); err != nil {
		//panic(err)
		e.Encode(err.Error())
	}
}

func errorMessage(w http.ResponseWriter, err error) {
	if err != nil {
		mustEncode(w, struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{Status: "error", Message: err.Error()})
	}
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
