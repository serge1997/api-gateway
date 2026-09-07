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

func New(cfg contracts.APIGateway, router contracts.Router) *server {
	s := &server{cfg: cfg, router: router}
	s.setListenAddr().
		setReadTimeout().
		setWriteTimeout()

	s.srv = newHttpServer(s)
	return s
}

func (s *server) setListenAddr() *server {
	if s.cfg.ListenAddr() == "" {
		s.listenAddr = defaultAddr
		return s
	}
	s.listenAddr = s.cfg.ListenAddr()
	return s
}

func (s *server) setReadTimeout() *server {
	if s.cfg.ReadTimeout().String() == "0s" {
		s.readTimeout = defaultReadWriteTimeout
		return s
	}
	s.readTimeout = s.cfg.ReadTimeout()
	return s
}

func (s *server) setWriteTimeout() *server {
	if s.cfg.ReadTimeout().String() == "0s" {
		s.writeTimeout = defaultReadWriteTimeout
		return s
	}
	s.readTimeout = s.cfg.WriteTimeout()
	return s
}

func (s *server) Listen() error {
	initLog(s)
	return s.srv.ListenAndServe()
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

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range s.router.Routes() {
		if route != nil {

		}
	}
}
