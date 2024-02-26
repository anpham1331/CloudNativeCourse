// Package main implements a client for movieinfo service
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/anpham1331/CloudNativeCourse/lab5/movieapi"
	"google.golang.org/grpc"
)

const (
	address      = "localhost:50051"
	defaultTitle = "Pulp fiction"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := movieapi.NewMovieInfoClient(conn)

	// Contact the server and print out its response.
	title := defaultTitle
	if len(os.Args) > 1 {
		title = os.Args[1]
	}

	// Timeout if server doesn't respond
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Get Movie Info
	r, err := c.GetMovieInfo(ctx, &movieapi.MovieRequest{Title: title})
	if err != nil {
		log.Fatalf("could not get movie info: %v", err)
	}
	log.Printf("Movie Info for %s %d %s %v", title, r.GetYear(), r.GetDirector(), r.GetCast())

	// Set Movie Info
	newMovieData := &movieapi.MovieData{
		Title:    "New Movie",
		Year:     "2024",
		Director: "New Director",
		Cast:     []string{"Actor1", "Actor2"},
	}
	status, err := c.SetMovieInfo(ctx, newMovieData)
	if err != nil {
		log.Fatalf("could not set movie info: %v", err)
	}
	log.Printf("Set Movie Info status: %s", status.GetCode())
}
