package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type myServer struct {
	port string
}

func main() {
	s := myServer{
		port: ":8888",
	}

	done := make(chan struct{})

	// Start server
	go func() {
		mux := http.NewServeMux()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("%s %v\n", r.Method, r.Header["Accept-Encoding"])
			w.Write([]byte("Hello"))
		})

		mux.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message":"hello!"}`))
		})

		server := &http.Server{
			Addr:    s.port,
			Handler: mux,
		}

		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}
	}()

	// Client
	go func() {
		url := fmt.Sprintf("http://localhost%s/message", s.port)

		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(body))

		close(done)
	}()

	<-done
}
