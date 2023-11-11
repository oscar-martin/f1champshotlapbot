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

	mux.HandleFunc("/rest/watch/sessionInfo", handleSessionInfoJson)
	mux.HandleFunc("/rest/watch/standings", handleStandingsJson)
	mux.HandleFunc("/rest/watch/standings/history", handleStandingsHistoryJson)

	fmt.Printf("Starting server in port %d\n", port)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func handleSessionInfoJson(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "data/sessionInfo.json")
}

func handleStandingsJson(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "data/standings.json")
}

func handleStandingsHistoryJson(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "data/standings_history.json")
}
