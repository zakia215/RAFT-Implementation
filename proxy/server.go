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
	Key   string `json:"key"`
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

	if err := s.connect(req.IP + ":" + req.Port); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": "Connected Successfully"})
}

func (s *Server) connect(address string) error {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

func (s *Server) PingHandler(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		http.Error(w, "Not connected", http.StatusBadRequest)
		return
	}

	args := common.ExecuteArgs{
		Command: "ping",
	}

	reply := common.ExecuteReply{}
	err := s.client.Call("Node.Execute", args, &reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	if reply.Response == "NOT LEADER" {
		if err := s.connect(reply.LeaderId); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.client.Call("Node.Execute", args, &reply)
		if err != nil {
			http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": reply.Response})
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

	args := common.ExecuteArgs{
		Command: "get",
		Key:     req.Key,
	}

	reply := common.ExecuteReply{}

	err := s.client.Call("Node.Execute", args, &reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	if reply.Response == "NOT LEADER" {
		if err := s.connect(reply.LeaderId); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.client.Call("Node.Execute", args, &reply)
		if err != nil {
			http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": reply.Response})
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

	args := common.ExecuteArgs{
		Command: "set",
		Key:     req.Key,
		Value:   req.Value,
	}

	reply := common.ExecuteReply{}

	err := s.client.Call("Node.Execute", args, &reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	if reply.Response == "NOT LEADER" {
		if err := s.connect(reply.LeaderId); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.client.Call("Node.Execute", args, &reply)
		if err != nil {
			http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": reply.Response})
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

	args := common.ExecuteArgs{
		Command: "strln",
		Key:     req.Key,
	}

	reply := common.ExecuteReply{}

	err := s.client.Call("Node.Execute", args, &reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	if reply.Response == "NOT LEADER" {
		if err := s.connect(reply.LeaderId); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.client.Call("Node.Execute", args, &reply)
		if err != nil {
			http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": reply.Response})
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

	args := common.ExecuteArgs{
		Command: "del",
		Key:     req.Key,
	}

	reply := common.ExecuteReply{}

	err := s.client.Call("Node.Execute", args, &reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	if reply.Response == "NOT LEADER" {
		if err := s.connect(reply.LeaderId); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.client.Call("Node.Execute", args, &reply)
		if err != nil {
			http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": reply.Response})
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

	args := common.ExecuteArgs{
		Command: "append",
		Key:     req.Key,
		Value:   req.Value,
	}

	reply := common.ExecuteReply{}

	err := s.client.Call("Node.Execute", args, &reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	if reply.Response == "NOT LEADER" {
		if err := s.connect(reply.LeaderId); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.client.Call("Node.Execute", args, &reply)
		if err != nil {
			http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"reply": reply.Response})
}

func (s *Server) LogHandler(w http.ResponseWriter, r *http.Request) {
	if s.client == nil {
		http.Error(w, "Not connected", http.StatusBadRequest)
		return
	}

	logEntries := new([]common.LogEntry)
	err := s.client.Call("Node.GetLog", struct{}{}, &logEntries)

	if err != nil {
		http.Error(w, fmt.Sprintf("RPC call error: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{"reply": *logEntries}
	json.NewEncoder(w).Encode(response)
}
