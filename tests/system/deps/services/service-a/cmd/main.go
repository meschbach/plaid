package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/sys/unix"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ops/liveness", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("ok"))
	})
	address := "127.0.0.1:9123"
	srv := &http.Server{
		Handler: r,
		Addr:    address,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	signalChannel := make(chan os.Signal, 1)
	go func() {
		for sig := range signalChannel {
			fmt.Printf("Received signal %d\n", sig)
			switch sig {
			case unix.SIGTERM:
				fmt.Printf("SIGTERM received, shutting down.")
				shutdownSystem := func() {
					ctx, done := context.WithTimeout(context.Background(), 30*time.Second)
					defer done()
					if err := srv.Shutdown(ctx); err != nil {
						fmt.Fprintf(os.Stderr, "[WWW] Shutdown failed: %e\n", err)
					}
				}
				shutdownSystem()
			}
		}
	}()
	signal.Notify(signalChannel, unix.SIGTERM)
	fmt.Println("Starting deps/service-a: listening at " + address)
	log.Fatal(srv.ListenAndServe())
}
