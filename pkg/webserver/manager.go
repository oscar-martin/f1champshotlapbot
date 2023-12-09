package webserver

import (
	"context"
	"f1champshotlapsbot/pkg/resources"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var addr = ":8080"
var upgrader = websocket.Upgrader{} // use default options

type Manager struct {
	r                *mux.Router
	serverIdToRouter map[string]*mux.Router
}

func NewManager() *Manager {
	m := &Manager{
		r:                mux.NewRouter(),
		serverIdToRouter: make(map[string]*mux.Router),
	}

	m.rootHandlers()
	return m
}

func (m *Manager) router() *mux.Router {
	return m.r
}

func (m *Manager) GetRouter(serverId, serverIdPrefix string) *mux.Router {
	r := m.r.NewRoute().PathPrefix(serverIdPrefix)
	sr := r.Subrouter()

	m.serverIdToRouter[serverId] = sr

	return sr
}

func (m *Manager) rootHandlers() {
	fs := http.FileServer(http.Dir(resources.ResourcesDir))
	resStr := "/resources/"

	m.r.PathPrefix(resStr).Handler(http.StripPrefix(resStr, fs))
}

func (m *Manager) Debug() {
	_ = m.router().Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})
}

func (m *Manager) Serve() {
	if os.Getenv("WEBSERVER_ADDRESS") != "" {
		addr = os.Getenv("WEBSERVER_ADDRESS")
	}
	srv := &http.Server{
		Addr: addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      m.router(), // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Printf("webserver listening on %s\n", addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("webserver shutting down")
}
