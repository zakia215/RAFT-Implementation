package main

import (
	"log"
	"net/http"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	server := NewServer()

	http.Handle("/connect", CORS(http.HandlerFunc(server.ConnectHandler)))
	http.Handle("/ping", CORS(http.HandlerFunc(server.PingHandler)))
	http.Handle("/getkey", CORS(http.HandlerFunc(server.GetKeyHandler)))
	http.Handle("/setkey", CORS(http.HandlerFunc(server.SetKeyHandler)))
	http.Handle("/strln", CORS(http.HandlerFunc(server.StrlnHandler)))
	http.Handle("/delkey", CORS(http.HandlerFunc(server.DelKeyHandler)))
	http.Handle("/append", CORS(http.HandlerFunc(server.AppendHandler)))
	http.Handle("/log", CORS(http.HandlerFunc(server.LogHandler)))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
