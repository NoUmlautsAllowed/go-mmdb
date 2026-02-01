package mmdb

import (
	"context"
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
	Client    *Client
	tmpl      *template.Template
	AuthToken string
}

type tmplData struct {
	IPInfo
	Authorized bool
	AuthToken  string
}

func NewServer(client *Client, authToken string) (*Server, error) {
	tmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"formatEpoch": func(epoch uint) string {
			return time.Unix(int64(epoch), 0).Format("2006-01-02 15:04:05 UTC")
		},
	}).ParseFS(indexHTML, "index.html")
	if err != nil {
		return nil, err
	}

	return &Server{
		Client:    client,
		tmpl:      tmpl,
		AuthToken: authToken,
	}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)

	if s.AuthToken != "" {
		return s.authMiddleware(mux)
	}

	return mux
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("auth")
		if token == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if token != s.AuthToken {
			format := r.URL.Query().Get("format")
			acceptHeader := r.Header.Get("Accept")

			if format == "html" || (format == "" && !strings.Contains(acceptHeader, "application/json") && !strings.Contains(acceptHeader, "text/plain")) {
				// For HTML, we still want to show the page but with an unauthorized banner
				r = r.WithContext(context.WithValue(r.Context(), "unauthorized", true))
				next.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}

		next.ServeHTTP(w, r)
	})
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

	unauthorized, _ := r.Context().Value("unauthorized").(bool)

	ipStr := r.URL.Query().Get("ip")
	var info IPInfo
	if unauthorized {
		status = http.StatusUnauthorized
	} else if ipStr != "" {
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

	switch {
	case format == "text" || strings.Contains(acceptHeader, "text/plain"):
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(status)
		w.Write([]byte(info.IP.String()))
	case format == "json" || strings.Contains(acceptHeader, "application/json"):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(info)
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		tkn := ""
		if !unauthorized {
			tkn = s.AuthToken
		}
		td := tmplData{info, !unauthorized, tkn}
		if err := s.tmpl.Execute(w, td); err != nil {
			log.Printf("Template execution error: %v", err)
		}
	}
}
