package runtime

import (
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
}

type Request struct {
	*http.Request
}

type Response struct {
	Status  int
	Body    string
	Headers map[string]string
}

type Server struct {
	Addr string
	Mux  *http.ServeMux
}

func NewServer(addr string) *Server {
	return &Server{
		Addr: addr,
		Mux:  http.NewServeMux(),
	}
}

func (s *Server) HandleFunc(pattern string, handler func(w ResponseWriter, r *Request)) {
	s.Mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		handler(ResponseWriter{w}, &Request{r})
	})
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.Addr, s.Mux)
}
