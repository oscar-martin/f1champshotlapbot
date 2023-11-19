package main

import (
	"fmt"
	"net/http"
)

func CreateServers(ports []int) {
	// this function will create a server for each port in the ports slice
	// each server will be created in a goroutine
	for _, port := range ports {
		go createHttpServer(port)
	}
}

func createHttpServer(port int) {
	// this function will create a http server in the specified port
	// the server will serve a simple message
	// the server will be created in a goroutine
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	})

	mux.HandleFunc("/rest/watch/sessionInfo", handleSessionInfoJsonF(port))
	mux.HandleFunc("/rest/watch/standings", handleStandingsJson)
	mux.HandleFunc("/rest/watch/standings/history", handleStandingsHistoryJson)

	fmt.Printf("Starting server in port %d\n", port)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func handleSessionInfoJsonF(port int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		portStr := fmt.Sprintf("%d", port)
		lastChar := portStr[len(portStr)-1:]
		http.ServeFile(w, r, fmt.Sprintf("data/sessionInfo%s.json", lastChar))
	}
}

func handleStandingsJson(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "data/standings.json")
}

func handleStandingsHistoryJson(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "data/standings_history.json")
}
