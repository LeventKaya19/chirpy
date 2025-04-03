package main

import "net/http"

func main() {
	httpServerMux := http.NewServeMux()
	httpServer := http.Server{
		Handler: httpServerMux,
		Addr:    ":8080",
	}
	httpServer.ListenAndServe()
}
