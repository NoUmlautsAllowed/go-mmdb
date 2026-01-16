package mmdb

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//go:embed index.html
var indexHTML embed.FS

type Server struct {
	Client *Client
	tmpl   *template.Template
}

func NewServer(client *Client) (*Server, error) {
	tmpl, err := template.ParseFS(indexHTML, "index.html")
	if err != nil {
		return nil, err
	}

	return &Server{
		Client: client,
		tmpl:   tmpl,
	}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	path := r.URL.Path
	method := r.Method

	var status int = http.StatusOK
	defer func() {
		duration := time.Since(start).Seconds()
		HttpRequestDuration.WithLabelValues(path, method).Observe(duration)
		HttpRequestsTotal.WithLabelValues(path, method, strconv.Itoa(status)).Inc()
	}()

	if r.URL.Path != "/" {
		status = http.StatusNotFound
		http.NotFound(w, r)
		return
	}

	ipStr := r.URL.Query().Get("ip")
	var info IPInfo
	if ipStr != "" {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			status = http.StatusBadRequest
			http.Error(w, "Invalid IP address", http.StatusBadRequest)
			return
		}
		info = s.Client.IPInfo(ip)
	} else {
		info = s.Client.IPInfoFromRequest(r)
	}

	format := r.URL.Query().Get("format")
	acceptHeader := r.Header.Get("Accept")

	if format == "json" || strings.Contains(acceptHeader, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.Execute(w, info); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
