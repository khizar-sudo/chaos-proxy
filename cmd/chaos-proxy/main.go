package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from chaos-proxy!\n")
		fmt.Fprintf(w, "Received: %s %s\n", r.Method, r.URL.Path)
	})

	addr := ":8080"
	fmt.Printf("Starting chaos-proxy on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
