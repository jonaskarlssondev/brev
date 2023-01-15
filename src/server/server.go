package brev

import (
	"net/http"
)

type server struct {
	router *http.ServeMux
	hub    *hub
}

func NewBrevServer() *server {
	s := &server{}

	s.hub = newHub()

	s.router = http.NewServeMux()

	s.setupRoutes()

	return s
}

func (s *server) setupRoutes() {
	s.router.HandleFunc("/", home)
	s.router.HandleFunc("/register", register(s))
	s.router.HandleFunc("/subscribe", subscribe(s.hub))
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Brev"))
}

func (s *server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
