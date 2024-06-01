package main

import (
	"common"
	"encoding/json"
	"fmt"
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

type GetKeyRequest struct {
	Key string `json:"key"`
}

type SetKeyRequest struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

type StrlnRequest = GetKeyRequest

type DelKeyRequest = GetKeyRequest

type AppendRequest = SetKeyRequest

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

func (s *Server) GetKeyHandler(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		http.Error(w, "Not connected", http.StatusBadRequest)
		return
	}

	var req GetKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reply := new(string)
	err := s.client.Call("RPCService.Get", req.Key, reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": *reply})
}

func (s *Server) SetKeyHandler(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		http.Error(w, "Not connected", http.StatusBadRequest)
		return
	}

	var req SetKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	args := []string{req.Key, req.Value}

	reply := new(string)
	err := s.client.Call("RPCService.Set", args, reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": *reply})
}

func (s *Server) StrlnHandler(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		http.Error(w, "Not connected", http.StatusBadRequest)
		return
	}

	var req StrlnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reply := new(string)
	err := s.client.Call("RPCService.Strln", req.Key, reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": *reply})
}

func (s *Server) DelKeyHandler(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		http.Error(w, "Not connected", http.StatusBadRequest)
		return
	}

	var req DelKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reply := new(string)
	err := s.client.Call("RPCService.Del", req.Key, reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": *reply})
}

func (s *Server) AppendHandler(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		http.Error(w, "Not connected", http.StatusBadRequest)
		return
	}

	var req AppendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	args := []string{req.Key, req.Value}

	reply := new(string)
	err := s.client.Call("RPCService.Append", args, reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": *reply})
}

func (s *Server) LogHandler(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		http.Error(w, "Not connected", http.StatusBadRequest)
		return
	}

	logEntries := new([]common.LogEntry)
	err := s.client.Call("RPCService.GetLog", struct{}{}, logEntries)

	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}


	response := map[string]interface{}{"reply": *logEntries}
	json.NewEncoder(w).Encode(response)

}