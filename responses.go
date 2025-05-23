package main

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	INTERNAL_SERVER_MESSAGE = "Somthing went wrong"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeErrorResponse(w http.ResponseWriter, err error, code int, msg string) {
	log.Printf("error: %v", err)
	w.WriteHeader(code)
	re := errorResponse{
		Error: msg,
	}
	resp, err := json.Marshal(re)
	if err != nil {
		log.Printf("%v", err)
	}
	w.Write(resp)
}

func writeHtml(w http.ResponseWriter, status int, body string) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(status)
	w.Write([]byte(body))
}

func writeJson(w http.ResponseWriter, status int, body []byte) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}
