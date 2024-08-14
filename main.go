package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

//go:embed template.html
var templateFS embed.FS

type Request struct {
	RequestLine string `json:"request_line"`
	Headers     string `json:"headers"`
	Body        string `json:"body"`
}

type Server struct {
	requests    []Request
	maxRequests int
	mu          sync.Mutex
	clients     map[*websocket.Conn]bool
	tmpl        *template.Template
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	host := flag.String("host", "localhost", "Host to listen on")
	port := flag.Int("port", 8000, "Port to listen on")
	maxRequests := flag.Int("max", 10, "Maximum number of requests to keep in buffer")
	certFile := flag.String("cert", "", "Path to SSL certificate file")
	keyFile := flag.String("key", "", "Path to SSL key file")
	flag.Parse()

	tmpl, err := template.ParseFS(templateFS, "template.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	server := &Server{
		requests:    make([]Request, 0, *maxRequests),
		maxRequests: *maxRequests,
		clients:     make(map[*websocket.Conn]bool),
		tmpl:        tmpl,
	}

	http.HandleFunc("/logs", server.handleLogs)
	http.HandleFunc("/logs/ws", server.handleWebSocket)
	http.HandleFunc("/logs/clear", server.handleClearLogs)
	http.HandleFunc("/bin/", server.handleRequest) // Changed from "/bin" to "/bin/"

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("Starting server on %s", addr)

	if *certFile != "" && *keyFile != "" {
		log.Fatal(http.ListenAndServeTLS(addr, *certFile, *keyFile, nil))
	} else {
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.tmpl.Execute(w, map[string]interface{}{
		"requests": s.requests,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/logs") {
		http.NotFound(w, r)
		return
	}

	// Check if the request path starts with "/bin/"
	if !strings.HasPrefix(r.URL.Path, "/bin/") {
		http.NotFound(w, r)
		return
	}

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	headers := ""
	for name, values := range r.Header {
		for _, value := range values {
			headers += fmt.Sprintf("%s: %s\n", name, value)
		}
	}

	request := Request{
		RequestLine: fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, r.Proto),
		Headers:     headers,
		Body:        string(body),
	}

	s.mu.Lock()
	if len(s.requests) >= s.maxRequests {
		s.requests = s.requests[1:]
	}
	s.requests = append(s.requests, request)
	s.mu.Unlock()

	s.broadcastRequests()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request logged"))
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	s.sendRequestsToClient(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			break
		}
	}
}

func (s *Server) handleClearLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	s.requests = make([]Request, 0, s.maxRequests)
	s.mu.Unlock()

	s.broadcastRequests()

	w.WriteHeader(http.StatusOK)
}

func (s *Server) broadcastRequests() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		s.sendRequestsToClient(client)
	}
}

func (s *Server) sendRequestsToClient(client *websocket.Conn) {
	data, err := json.Marshal(s.requests)
	if err != nil {
		log.Println("Error marshaling requests:", err)
		return
	}

	err = client.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Println("Error sending message:", err)
		delete(s.clients, client)
	}
}
