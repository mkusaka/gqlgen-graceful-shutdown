package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/mkusaka/gqlgen-graceful-shutdown/graph"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)
	serv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	serverChan := make(chan error, 1)
	go func() {
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
		if err := serv.ListenAndServe(); err != nil {
			serverChan <- err
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, os.Interrupt)

	select {
	case <-sig:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := serv.Shutdown(ctx); err != nil {
			fmt.Println(fmt.Errorf("server closed with err: %+v", err))
			os.Exit(1)
		}
		log.Println("server shutdown")
		log.Println("closing server")
	case <-serverChan:
		log.Println("server shutdown")
	}
}
