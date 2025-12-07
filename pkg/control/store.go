package control

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

type Tenant struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type APIKey struct {
	ID        string
	TenantID  string
	Secret    string
	Label     string
	CreatedAt time.Time
}

type Route struct {
	ID            string
	TenantID      string
	MatchType     string
	MatchValue    string
	TargetChannel string
	CreatedAt     time.Time
}

func (s *Store) CreateTenant(ctx context.Context, name string) (*Tenant, error) {
	id := uuid.NewString()
	now := time.Now().UTC()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO tenants (id, name, created_at) VALUES (?, ?, ?)`,
		id, name, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	return &Tenant{
		ID:        id,
		Name:      name,
		CreatedAt: now,
	}, nil
}

func (s *Store) ListTenants(ctx context.Context) ([]Tenant, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, created_at FROM tenants ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", err)
	}
	defer rows.Close()

	var out []Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan tenant: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) CreateAPIKey(ctx context.Context, tenantID, label string) (*APIKey, error) {
	id := uuid.NewString()
	secret := "sk_" + uuid.NewString()
	now := time.Now().UTC()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO api_keys (id, tenant_id, secret, label, created_at)
         VALUES (?, ?, ?, ?, ?)`,
		id, tenantID, secret, label, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}

	return &APIKey{
		ID:        id,
		TenantID:  tenantID,
		Secret:    secret,
		Label:     label,
		CreatedAt: now,
	}, nil
}

func (s *Store) GetTenantByAPIKey(ctx context.Context, secret string) (*Tenant, *APIKey, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT t.id, t.name, t.created_at,
                k.id, k.tenant_id, k.secret, k.label, k.created_at
           FROM api_keys k
           JOIN tenants t ON t.id = k.tenant_id
          WHERE k.secret = ?`,
		secret,
	)

	var t Tenant
	var k APIKey
	if err := row.Scan(
		&t.ID, &t.Name, &t.CreatedAt,
		&k.ID, &k.TenantID, &k.Secret, &k.Label, &k.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("get tenant by api key: %w", err)
	}
	return &t, &k, nil
}

func (s *Store) CreateRoute(ctx context.Context, tenantID, matchType, matchValue, targetChannel string) (*Route, error) {
	matchType = strings.ToUpper(matchType)
	if matchType != "EXACT" && matchType != "PREFIX" {
		return nil, fmt.Errorf("invalid match_type: %s", matchType)
	}

	id := uuid.NewString()
	now := time.Now().UTC()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO routes (id, tenant_id, match_type, match_value, target_channel, created_at)
         VALUES (?, ?, ?, ?, ?, ?)`,
		id, tenantID, matchType, matchValue, targetChannel, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create route: %w", err)
	}

	return &Route{
		ID:            id,
		TenantID:      tenantID,
		MatchType:     matchType,
		MatchValue:    matchValue,
		TargetChannel: targetChannel,
		CreatedAt:     now,
	}, nil
}

func (s *Store) ListRoutes(ctx context.Context, tenantID string) ([]Route, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, match_type, match_value, target_channel, created_at
           FROM routes
          WHERE tenant_id = ?
          ORDER BY created_at ASC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list routes: %w", err)
	}
	defer rows.Close()

	var out []Route
	for rows.Next() {
		var r Route
		if err := rows.Scan(&r.ID, &r.TenantID, &r.MatchType, &r.MatchValue, &r.TargetChannel, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan route: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) FindRoutesForEvent(ctx context.Context, tenantID, eventType string) ([]Route, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, match_type, match_value, target_channel, created_at
           FROM routes
          WHERE tenant_id = ?`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("find routes: %w", err)
	}
	defer rows.Close()

	var matched []Route
	for rows.Next() {
		var r Route
		if err := rows.Scan(&r.ID, &r.TenantID, &r.MatchType, &r.MatchValue, &r.TargetChannel, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan route: %w", err)
		}

		switch r.MatchType {
		case "EXACT":
			if eventType == r.MatchValue {
				matched = append(matched, r)
			}
		case "PREFIX":
			if strings.HasPrefix(eventType, r.MatchValue) {
				matched = append(matched, r)
			}
		}
	}
	return matched, rows.Err()
}
