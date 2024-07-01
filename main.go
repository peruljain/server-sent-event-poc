package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, visit /events to get server-sent events.")
	})
	http.HandleFunc("/events", handleEvents)

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleEvents(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a channel to send messages
	messageChan := make(chan string)

	// Goroutine to send messages to the channel
	go func() {
		for {
			time.Sleep(2 * time.Second)
			messageChan <- fmt.Sprintf("Current time: %s", time.Now().Format(time.RFC1123))
		}
	}()

	// Stream messages from the channel to the client
	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher, ok := w.(http.Flusher)
			if ok {
				flusher.Flush()
			}
		case <-r.Context().Done():
			log.Println("Client disconnected")
			return
		}
	}
}
