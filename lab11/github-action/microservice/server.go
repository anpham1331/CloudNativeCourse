// Hello world microservice - server.go
package microservice

import (
	"fmt"
	"net/http"
)

func NewServer(host string, port string) *http.Server {
	addr := fmt.Sprintf("%s:%s", host, port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!\n")
}
