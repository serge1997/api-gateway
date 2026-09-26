package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/serge1997/apigateway/internal/contracts"
)

type server struct {
	srv               *http.Server
	listenAddr        string
	timeout           time.Duration
	cfg               contracts.APIGateway
	router            contracts.Router
	readHeaderTimeout time.Duration
}

func New(opts ...Option) *server {
	s := &server{
		readHeaderTimeout: defaultReadHeaderTimout,
		listenAddr:        defaultAddr,
	}
	for _, opt := range opts {
		opt(s)
	}
	s.applyConfigValues()
	return s
}

func (s *server) setListenAddr() {
	if s.cfg.ListenAddr() == "" {
		s.listenAddr = defaultAddr
		return
	}
	s.listenAddr = s.cfg.ListenAddr()
}

func (s *server) setReadHeaderTimeout() {
	var z time.Duration
	if s.cfg.ReadHeaderTimeout() == z {
		s.readHeaderTimeout = defaultReadHeaderTimout
		return
	}
	s.readHeaderTimeout = s.cfg.ReadHeaderTimeout()
}

func (s *server) setHttpServer() *server {
	s.srv = newHttpServer(s)
	return s
}

func (s *server) Listen() error {
	initLog(s)
	return s.srv.ListenAndServe()
}

func (s *server) applyConfigValues() {
	s.setListenAddr()
	s.setReadHeaderTimeout()
	s.setHttpServer()
}

func (s *server) Close() error {
	return s.srv.Close()
}

func newHttpServer(s *server) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%s", s.listenAddr),
		ReadHeaderTimeout: s.readHeaderTimeout,
		Handler:           muxHandlers(s),
	}
}

func initLog(s *server) {
	fmt.Printf("Server start at :%s\n", s.listenAddr)
}

func muxHandlers(s *server) *http.ServeMux {
	mux := http.NewServeMux()
	for _, route := range s.router.Routes() {
		if route != nil {
			for _, method := range route.Methods() {
				path := fmt.Sprintf("%s %s", method, route.Path())
				mux.HandleFunc(path, route.Handler())
			}
		}
	}
	return mux
}
