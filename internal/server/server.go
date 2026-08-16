package server

import (
	"fmt"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(port string) *Server {
	// Creating my server at the given port
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	return &Server{
		httpServer: server,
	}
}

func (s *Server) Start() error{
	return s.httpServer.ListenAndServe()
}
