package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	_ "time"
)

type myServer struct {
	port string
}

func main() {
	var s myServer
	s.port = ":8888"
	go func() {
		mux := http.NewServeMux()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("%s %s  ", r.Method, r.Header["Accept-Encoding"])
			w.Write([]byte("Hello"))
		})

		server := &http.Server{
			Addr:    s.port,
			Handler: mux,
		}
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				fmt.Printf("error running http server: %s\n", err)
			}
		}
	}()
	time.Sleep(200 * time.Millisecond)
	go func() {
		url := "http://localhost" + s.port
		requestUrl, err := http.Get(url)
		if err != nil {
			log.Fatal("cant get the request")
		}
		defer requestUrl.Body.Close()
		fmt.Printf("%d\n", requestUrl.StatusCode)
	}()
	time.Sleep(200 * time.Millisecond)
}
