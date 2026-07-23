package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

var name, port string

type HealthResponse struct {
	Status string `json:"status"`
}

type RootResponse struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

func health(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := HealthResponse{
		Status: "ok",
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
	}
}

func root(w http.ResponseWriter, req *http.Request) {
	resp := RootResponse{
		Name: name,
		Port: port,
	}

	if req.URL.Path != "/" {
		http.NotFound(w, req)
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
}

func sleep(w http.ResponseWriter, r *http.Request) {
	time.Sleep(3 * time.Second)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("done sleeping"))
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
	}
}

func main() {
	name = os.Getenv("NAME")
	port = os.Getenv("PORT")

	if name == "" {
		name = "default"
	}

	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", root)
	http.HandleFunc("/health", health)
	http.HandleFunc("/sleep", sleep)

	svr := http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 4 * time.Second,
	}

	log.Printf("[%s] running on %s", name, port)

	log.Fatal(svr.ListenAndServe())
}
