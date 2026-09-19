package server

import "github.com/serge1997/apigateway/internal/contracts"

type Option func(*server)

func WithRouter(router contracts.Router) Option {
	return func(s *server) {
		s.router = router
	}
}

func WithApiGateway(cfg contracts.APIGateway) Option {
	return func(s *server) {
		s.cfg = cfg
	}
}
