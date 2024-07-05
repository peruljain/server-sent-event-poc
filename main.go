package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var logFile *os.File

func init() {
	var err error
	logFile, err = os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	rand.Seed(time.Now().UnixNano())
}

func main() {
	defer logFile.Close()

	go func() {
		logRandomText()
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, visit /events to get server-sent events.")
	})
	http.HandleFunc("/events", handleEvents)
	http.HandleFunc("/logs", logsHandler)

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

func logsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Create a channel to signal when the client closes the connection
	notify := r.Context().Done()

	// Tail the log file
	file, err := os.Open("server.log")
	if err != nil {
		http.Error(w, "Unable to open log file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Start at the end of the file
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Unable to get file stats", http.StatusInternalServerError)
		return
	}
	file.Seek(stat.Size(), 0)

	reader := bufio.NewReader(file)

	for {
		select {
		case <-notify:
			log.Println("Client closed connection")
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				log.Println("Error reading log file:", err)
				return
			}
			if len(line) > 0 {
				_, err := fmt.Fprintf(w, "data: %s\n\n", line)
				if err != nil {
					log.Println("Error writing to response:", err)
					return
				}
				w.(http.Flusher).Flush()
			}
			time.Sleep(1 * time.Second) // Sleep to avoid busy looping
		}
	}
}

func generateRandomText(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	text := make([]rune, length)
	for i := range text {
		text[i] = letters[rand.Intn(len(letters))]
	}
	return string(text)
}

func logRandomText() {
	for {
		text := generateRandomText(100) // Generate a random text of length 100
		log.Println("Random Text:", text)
		time.Sleep(5 * time.Second) // Log random text every 5 seconds
	}
}
