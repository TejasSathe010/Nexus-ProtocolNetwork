package control

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	ctl "github.com/tejassathe/Nexus-ProtocolNetwork/pkg/control"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
)

type Handler struct {
	log   logger.Logger
	store *ctl.Store
}

func NewHandler(log logger.Logger, store *ctl.Store) *Handler {
	return &Handler{
		log:   log,
		store: store,
	}
}

type createTenantRequest struct {
	Name string `json:"name"`
}

type tenantResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type apiKeyResponse struct {
	ID        string `json:"id"`
	Secret    string `json:"secret"`
	Label     string `json:"label"`
	CreatedAt string `json:"created_at"`
}

type createAPIKeyRequest struct {
	Label string `json:"label"`
}

type createRouteRequest struct {
	MatchType     string `json:"match_type"`
	MatchValue    string `json:"match_value"`
	TargetChannel string `json:"target_channel"`
}

type routeResponse struct {
	ID            string `json:"id"`
	MatchType     string `json:"match_type"`
	MatchValue    string `json:"match_value"`
	TargetChannel string `json:"target_channel"`
	CreatedAt     string `json:"created_at"`
}

func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req createTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	t, err := h.store.CreateTenant(ctx, req.Name)
	if err != nil {
		h.log.Error("create tenant failed", "err", err)
		http.Error(w, "create tenant failed", http.StatusInternalServerError)
		return
	}

	resp := tenantResponse{
		ID:        t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenants, err := h.store.ListTenants(ctx)
	if err != nil {
		h.log.Error("list tenants failed", "err", err)
		http.Error(w, "list tenants failed", http.StatusInternalServerError)
		return
	}

	out := make([]tenantResponse, 0, len(tenants))
	for _, t := range tenants {
		out = append(out, tenantResponse{
			ID:        t.ID,
			Name:      t.Name,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenant_id")
	if tenantID == "" {
		http.Error(w, "missing tenant_id", http.StatusBadRequest)
		return
	}

	var req createAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	key, err := h.store.CreateAPIKey(ctx, tenantID, req.Label)
	if err != nil {
		h.log.Error("create api key failed", "err", err)
		http.Error(w, "create api key failed", http.StatusInternalServerError)
		return
	}

	resp := apiKeyResponse{
		ID:        key.ID,
		Secret:    key.Secret,
		Label:     key.Label,
		CreatedAt: key.CreatedAt.Format(time.RFC3339),
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenant_id")
	if tenantID == "" {
		http.Error(w, "missing tenant_id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	routes, err := h.store.ListRoutes(ctx, tenantID)
	if err != nil {
		h.log.Error("list routes failed", "err", err)
		http.Error(w, "list routes failed", http.StatusInternalServerError)
		return
	}

	out := make([]routeResponse, 0, len(routes))
	for _, rt := range routes {
		out = append(out, routeResponse{
			ID:            rt.ID,
			MatchType:     rt.MatchType,
			MatchValue:    rt.MatchValue,
			TargetChannel: rt.TargetChannel,
			CreatedAt:     rt.CreatedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateRoute(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenant_id")
	if tenantID == "" {
		http.Error(w, "missing tenant_id", http.StatusBadRequest)
		return
	}

	var req createRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.MatchType == "" || req.MatchValue == "" || req.TargetChannel == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	rt, err := h.store.CreateRoute(ctx, tenantID, req.MatchType, req.MatchValue, req.TargetChannel)
	if err != nil {
		h.log.Error("create route failed", "err", err)
		http.Error(w, "create route failed", http.StatusInternalServerError)
		return
	}

	resp := routeResponse{
		ID:            rt.ID,
		MatchType:     rt.MatchType,
		MatchValue:    rt.MatchValue,
		TargetChannel: rt.TargetChannel,
		CreatedAt:     rt.CreatedAt.Format(time.RFC3339),
	}
	writeJSON(w, http.StatusCreated, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
