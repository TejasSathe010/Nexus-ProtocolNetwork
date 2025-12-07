package gateway

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	ctl "github.com/tejassathe/Nexus-ProtocolNetwork/pkg/control"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
)

type contextKey string

const (
	ContextKeyRequestID contextKey = "request_id"
	ContextKeyTenantID  contextKey = "tenant_id"
	ContextKeyAPIKey    contextKey = "api_key"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-Id")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), ContextKeyRequestID, reqID)
		w.Header().Set("X-Request-Id", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RecoverMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered", "panic", rec)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func LoggingMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			reqID, _ := r.Context().Value(ContextKeyRequestID).(string)
			log.Info("http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", reqID,
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.statusCode = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := s.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, fmt.Errorf("http.Hijacker not supported")
}

func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (s *statusRecorder) Push(target string, opts *http.PushOptions) error {
	if p, ok := s.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func AuthMiddleware(log logger.Logger, store *ctl.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerTenant := r.Header.Get("X-Tenant-Id")
			apiKey := r.Header.Get("X-Api-Key")

			if apiKey == "" {
				http.Error(w, "missing X-Api-Key", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			tenant, _, err := store.GetTenantByAPIKey(ctx, apiKey)
			if err != nil {
				log.Error("auth lookup failed", "err", err)
				http.Error(w, "auth error", http.StatusInternalServerError)
				return
			}
			if tenant == nil {
				http.Error(w, "invalid api key", http.StatusUnauthorized)
				return
			}

			if headerTenant != "" && headerTenant != tenant.ID {
				http.Error(w, "tenant mismatch", http.StatusForbidden)
				return
			}

			ctx = context.WithValue(ctx, ContextKeyTenantID, tenant.ID)
			ctx = context.WithValue(ctx, ContextKeyAPIKey, apiKey)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
