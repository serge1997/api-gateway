package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/serge1997/apigateway/internal/contracts"
)

type server struct {
	srv          *http.Server
	listenAddr   string
	readTimeout  time.Duration
	writeTimeout time.Duration
	timeout      time.Duration
	cfg          contracts.APIGateway
	router       contracts.Router
}

func New(opts ...Option) *server {
	s := &server{}
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

func (s *server) setReadTimeout() {
	var z time.Duration
	if s.cfg.ReadTimeout() == z {
		s.readTimeout = defaultReadWriteTimeout
		return
	}
	s.readTimeout = s.cfg.ReadTimeout()
}

func (s *server) setWriteTimeout() {
	var z time.Duration
	if s.cfg.ReadTimeout() == z {
		s.writeTimeout = defaultReadWriteTimeout
		return
	}
	s.readTimeout = s.cfg.WriteTimeout()
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
	s.setReadTimeout()
	s.setWriteTimeout()
	s.setHttpServer()
}

func (s *server) Close() error {
	return s.srv.Close()
}

func newHttpServer(s *server) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%s", s.listenAddr),
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		Handler:      muxHandlers(s),
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
