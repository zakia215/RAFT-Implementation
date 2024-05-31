package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
)

type Server struct {
	client *rpc.Client
}

type ConnectRequest struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) ConnectHandler(w http.ResponseWriter, r *http.Request) {
	var req ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client, err := rpc.Dial("tcp", req.IP+":"+req.Port)
	if err != nil {
		http.Error(w, fmt.Sprintf("Dial error: %v", err), http.StatusInternalServerError)
		return
	}

	s.client = client
	fmt.Fprintln(w, "Connected successfully")
}

func (s *Server) PingHandler(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		http.Error(w, "Not connected", http.StatusBadRequest)
		return
	}

	reply := new(string)
	err := s.client.Call("RPCService.Ping", struct{}{}, reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": *reply})
}

func main() {
	server := NewServer()

	http.HandleFunc("/connect", server.ConnectHandler)
	http.HandleFunc("/ping", server.PingHandler)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
