package main

import (
	"log"
	"net/http"
)

func main() {
	server := NewServer()

	http.HandleFunc("/connect", server.ConnectHandler)
	http.HandleFunc("/ping", server.PingHandler)
	http.HandleFunc("/getkey", server.GetKeyHandler)
	http.HandleFunc("/setkey", server.SetKeyHandler)
	http.HandleFunc("/strln", server.StrlnHandler)
	http.HandleFunc("/delkey", server.DelKeyHandler)
	http.HandleFunc("/append", server.AppendHandler)
	http.HandleFunc("/log", server.LogHandler)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
