package cors

import (
	"net/http"
	"strings"
)

type cors struct {
}

func New() cors {
	return cors{}
}

func (c cors) Handle(w http.ResponseWriter, r *http.Request, cors []string) bool {
	origins := strings.Join(cors, ",")
	w.Header().Set("Access-Control-Allow-Origin", origins)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-service-name")
	return r.Method == http.MethodOptions
}
