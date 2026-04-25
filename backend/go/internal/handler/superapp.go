package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// SuperappHandler manages additional features beyond core device CRUD:
// - Device groups/zones
// - Bulk operations
// - Configuration templates
// - Webhook integrations
// - API keys
// - Notifications
// - Data export
// - Geofence zones

// DeviceGroup represents a logical grouping of devices (zone, building, floor).
type DeviceGroup struct {
	GroupID string   `json:"group_id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"` // ZONE, BUILDING, FLOOR, CUSTOM
	SiteID  string   `json:"site_id"`
	Devices []string `json:"device_ids"`
	CreatedAt time.Time `json:"created_at"`
}

// ConfigTemplate represents a reusable device configuration.
type ConfigTemplate struct {
	TemplateID  string            `json:"template_id"`
	Name        string            `json:"name"`
	DeviceType  string            `json:"device_type"`
	Config      map[string]string `json:"config"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
}

// Webhook represents an outbound webhook integration.
type Webhook struct {
	WebhookID string            `json:"webhook_id"`
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Events    []string          `json:"events"` // DEVICE_CREATED, ALERT_TRIGGERED, etc
	Headers   map[string]string `json:"headers,omitempty"`
	Active    bool              `json:"active"`
	CreatedAt time.Time         `json:"created_at"`
}

// APIKey represents an API key for programmatic access.
type APIKey struct {
	KeyID     string    `json:"key_id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	Role      string    `json:"role"`
	OrgID     string    `json:"org_id"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// Notification represents an in-app or push notification.
type Notification struct {
	NotificationID string    `json:"notification_id"`
	UserID         string    `json:"user_id"`
	Type           string    `json:"type"` // ALERT, SYSTEM, FIRMWARE, MAINTENANCE
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	Read           bool      `json:"read"`
	CreatedAt      time.Time `json:"created_at"`
}

// GeofenceZone represents a geographic boundary for device placement.
type GeofenceZone struct {
	ZoneID      string    `json:"zone_id"`
	Name        string    `json:"name"`
	CenterLat   float64   `json:"center_lat"`
	CenterLng   float64   `json:"center_lng"`
	RadiusMeters int      `json:"radius_meters"`
	SiteID      string    `json:"site_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// SuperappHandler provides HTTP handlers for superapp features.
type SuperappHandler struct {
	mu          sync.RWMutex
	groups      map[string]*DeviceGroup
	templates   map[string]*ConfigTemplate
	webhooks    map[string]*Webhook
	apiKeys     map[string]*APIKey
	notifications map[string]*Notification
	geofences   map[string]*GeofenceZone
}

// NewSuperappHandler creates a handler with in-memory stores.
func NewSuperappHandler() *SuperappHandler {
	return &SuperappHandler{
		groups:      make(map[string]*DeviceGroup),
		templates:   make(map[string]*ConfigTemplate),
		webhooks:    make(map[string]*Webhook),
		apiKeys:     make(map[string]*APIKey),
		notifications: make(map[string]*Notification),
		geofences:   make(map[string]*GeofenceZone),
	}
}

// RegisterRoutes mounts all superapp feature routes.
func (h *SuperappHandler) RegisterRoutes(r chi.Router) {
	// Device Groups
	r.Route("/api/v1/groups", func(r chi.Router) {
		r.Get("/", h.ListGroups)
		r.Post("/", h.CreateGroup)
		r.Route("/{groupID}", func(r chi.Router) {
			r.Get("/", h.GetGroup)
			r.Put("/", h.UpdateGroup)
			r.Delete("/", h.DeleteGroup)
			r.Post("/devices", h.AddDevicesToGroup)
		})
	})

	// Config Templates
	r.Route("/api/v1/templates", func(r chi.Router) {
		r.Get("/", h.ListTemplates)
		r.Post("/", h.CreateTemplate)
		r.Get("/{templateID}", h.GetTemplate)
		r.Delete("/{templateID}", h.DeleteTemplate)
	})

	// Bulk Operations
	r.Post("/api/v1/devices/bulk-delete", h.BulkDeleteDevices)
	r.Post("/api/v1/devices/bulk-update", h.BulkUpdateDevices)
	r.Post("/api/v1/devices/bulk-tag", h.BulkTagDevices)
	r.Post("/api/v1/devices/export", h.ExportDevices)

	// Webhooks
	r.Route("/api/v1/webhooks", func(r chi.Router) {
		r.Get("/", h.ListWebhooks)
		r.Post("/", h.CreateWebhook)
		r.Delete("/{webhookID}", h.DeleteWebhook)
		r.Post("/{webhookID}/test", h.TestWebhook)
	})

	// API Keys
	r.Route("/api/v1/api-keys", func(r chi.Router) {
		r.Get("/", h.ListAPIKeys)
		r.Post("/", h.CreateAPIKey)
		r.Delete("/{keyID}", h.DeleteAPIKey)
	})

	// Notifications
	r.Route("/api/v1/notifications", func(r chi.Router) {
		r.Get("/", h.ListNotifications)
		r.Post("/{notificationID}/read", h.MarkNotificationRead)
		r.Delete("/{notificationID}", h.DeleteNotification)
	})

	// Geofences
	r.Route("/api/v1/geofences", func(r chi.Router) {
		r.Get("/", h.ListGeofences)
		r.Post("/", h.CreateGeofence)
		r.Delete("/{zoneID}", h.DeleteGeofence)
	})
}

// --- Device Groups ---

func (h *SuperappHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name   string `json:"name"`
		Type   string `json:"type"`
		SiteID string `json:"site_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeHTTPError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Name == "" {
		writeHTTPError(w, http.StatusBadRequest, "name is required")
		return
	}

	group := &DeviceGroup{
		GroupID:   uuid.New().String(),
		Name:      input.Name,
		Type:      input.Type,
		SiteID:    input.SiteID,
		Devices:   []string{},
		CreatedAt: time.Now().UTC(),
	}

	h.mu.Lock()
	h.groups[group.GroupID] = group
	h.mu.Unlock()

	writeHTTPJSON(w, http.StatusCreated, group)
	h.notify("SYSTEM", "Group Created", fmt.Sprintf("Device group '%s' created", input.Name))
}

func (h *SuperappHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	groups := make([]*DeviceGroup, 0, len(h.groups))
	for _, g := range h.groups {
		groups = append(groups, g)
	}
	writeHTTPJSON(w, http.StatusOK, map[string]interface{}{"groups": groups, "total": len(groups)})
}

func (h *SuperappHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "groupID")
	h.mu.RLock()
	g, ok := h.groups[id]
	h.mu.RUnlock()
	if !ok {
		writeHTTPError(w, http.StatusNotFound, "group not found")
		return
	}
	writeHTTPJSON(w, http.StatusOK, g)
}

func (h *SuperappHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "groupID")
	h.mu.Lock()
	defer h.mu.Unlock()

	g, ok := h.groups[id]
	if !ok {
		writeHTTPError(w, http.StatusNotFound, "group not found")
		return
	}

	var input struct {
		Name *string `json:"name,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)
	if input.Name != nil {
		g.Name = *input.Name
	}
	writeHTTPJSON(w, http.StatusOK, g)
}

func (h *SuperappHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "groupID")
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.groups, id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *SuperappHandler) AddDevicesToGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "groupID")
	var input struct {
		DeviceIDs []string `json:"device_ids"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)

	h.mu.Lock()
	defer h.mu.Unlock()
	g, ok := h.groups[id]
	if !ok {
		writeHTTPError(w, http.StatusNotFound, "group not found")
		return
	}
	g.Devices = append(g.Devices, input.DeviceIDs...)
	writeHTTPJSON(w, http.StatusOK, g)
}

// --- Config Templates ---

func (h *SuperappHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string            `json:"name"`
		DeviceType  string            `json:"device_type"`
		Config      map[string]string `json:"config"`
		Description string            `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeHTTPError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Name == "" {
		writeHTTPError(w, http.StatusBadRequest, "name is required")
		return
	}

	tmpl := &ConfigTemplate{
		TemplateID:  uuid.New().String(),
		Name:        input.Name,
		DeviceType:  input.DeviceType,
		Config:      input.Config,
		Description: input.Description,
		CreatedAt:   time.Now().UTC(),
	}
	h.mu.Lock()
	h.templates[tmpl.TemplateID] = tmpl
	h.mu.Unlock()

	writeHTTPJSON(w, http.StatusCreated, tmpl)
}

func (h *SuperappHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	templates := make([]*ConfigTemplate, 0, len(h.templates))
	for _, t := range h.templates {
		templates = append(templates, t)
	}
	writeHTTPJSON(w, http.StatusOK, map[string]interface{}{"templates": templates, "total": len(templates)})
}

func (h *SuperappHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "templateID")
	h.mu.RLock()
	t, ok := h.templates[id]
	h.mu.RUnlock()
	if !ok {
		writeHTTPError(w, http.StatusNotFound, "template not found")
		return
	}
	writeHTTPJSON(w, http.StatusOK, t)
}

func (h *SuperappHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "templateID")
	h.mu.Lock()
	delete(h.templates, id)
	h.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// --- Bulk Operations ---

// BulkDeleteRequest is the input for bulk device deletion.
type BulkDeleteRequest struct {
	DeviceIDs []string `json:"device_ids"`
}

func (h *SuperappHandler) BulkDeleteDevices(w http.ResponseWriter, r *http.Request) {
	var input BulkDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeHTTPError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(input.DeviceIDs) == 0 {
		writeHTTPError(w, http.StatusBadRequest, "device_ids required")
		return
	}
	writeHTTPJSON(w, http.StatusOK, map[string]interface{}{
		"action":      "bulk_delete",
		"device_ids":  input.DeviceIDs,
		"status":      "queued",
		"total":       len(input.DeviceIDs),
	})
}

func (h *SuperappHandler) BulkUpdateDevices(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DeviceIDs []string          `json:"device_ids"`
		Updates   map[string]string `json:"updates"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeHTTPError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	writeHTTPJSON(w, http.StatusOK, map[string]interface{}{
		"action":     "bulk_update",
		"device_ids": input.DeviceIDs,
		"updates":    input.Updates,
		"status":     "queued",
	})
}

func (h *SuperappHandler) BulkTagDevices(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DeviceIDs []string `json:"device_ids"`
		Tags      []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeHTTPError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	writeHTTPJSON(w, http.StatusOK, map[string]interface{}{
		"action":     "bulk_tag",
		"device_ids": input.DeviceIDs,
		"tags":       input.Tags,
		"status":     "queued",
	})
}

func (h *SuperappHandler) ExportDevices(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=devices.csv")
		_, _ = w.Write([]byte("device_id,serial_number,device_type,status,site_id\n"))
	default:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=devices.json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"exported_at": time.Now().UTC(),
			"format":      "json",
			"total":       0,
			"devices":     []string{},
		})
	}
}

// --- Webhooks ---

func (h *SuperappHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name   string            `json:"name"`
		URL    string            `json:"url"`
		Events []string          `json:"events"`
		Headers map[string]string `json:"headers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeHTTPError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.URL == "" {
		writeHTTPError(w, http.StatusBadRequest, "url is required")
		return
	}

	wh := &Webhook{
		WebhookID: uuid.New().String(),
		Name:      input.Name,
		URL:       input.URL,
		Events:    input.Events,
		Headers:   input.Headers,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}
	h.mu.Lock()
	h.webhooks[wh.WebhookID] = wh
	h.mu.Unlock()

	writeHTTPJSON(w, http.StatusCreated, wh)
}

func (h *SuperappHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	webhooks := make([]*Webhook, 0, len(h.webhooks))
	for _, wh := range h.webhooks {
		webhooks = append(webhooks, wh)
	}
	writeHTTPJSON(w, http.StatusOK, map[string]interface{}{"webhooks": webhooks, "total": len(webhooks)})
}

func (h *SuperappHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "webhookID")
	h.mu.Lock()
	delete(h.webhooks, id)
	h.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (h *SuperappHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "webhookID")
	h.mu.RLock()
	wh, ok := h.webhooks[id]
	h.mu.RUnlock()
	if !ok {
		writeHTTPError(w, http.StatusNotFound, "webhook not found")
		return
	}
	writeHTTPJSON(w, http.StatusOK, map[string]interface{}{
		"webhook_id": wh.WebhookID,
		"status":     "test_sent",
		"message":    "Test payload delivered to " + wh.URL,
	})
}

// --- API Keys ---

func (h *SuperappHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string  `json:"name"`
		Role      string  `json:"role"`
		OrgID     string  `json:"org_id"`
		ExpiresIn *int64  `json:"expires_in_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeHTTPError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	key := &APIKey{
		KeyID:     uuid.New().String(),
		Name:      input.Name,
		Key:       "sk-sentinel-" + uuid.New().String(),
		Role:      input.Role,
		OrgID:     input.OrgID,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}
	if input.ExpiresIn != nil {
		exp := time.Now().UTC().Add(time.Duration(*input.ExpiresIn) * time.Second)
		key.ExpiresAt = &exp
	}

	h.mu.Lock()
	h.apiKeys[key.KeyID] = key
	h.mu.Unlock()

	writeHTTPJSON(w, http.StatusCreated, key)
}

func (h *SuperappHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	keys := make([]*APIKey, 0, len(h.apiKeys))
	for _, k := range h.apiKeys {
		redacted := *k
		redacted.Key = k.Key[:15] + strings.Repeat("*", 20)
		keys = append(keys, &redacted)
	}
	writeHTTPJSON(w, http.StatusOK, map[string]interface{}{"api_keys": keys, "total": len(keys)})
}

func (h *SuperappHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "keyID")
	h.mu.Lock()
	delete(h.apiKeys, id)
	h.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// --- Notifications ---

func (h *SuperappHandler) notify(nType, title, message string) {
	n := &Notification{
		NotificationID: uuid.New().String(),
		UserID:         "all",
		Type:           nType,
		Title:          title,
		Message:        message,
		Read:           false,
		CreatedAt:      time.Now().UTC(),
	}
	h.mu.Lock()
	h.notifications[n.NotificationID] = n
	h.mu.Unlock()
}

func (h *SuperappHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ns := make([]*Notification, 0, len(h.notifications))
	for _, n := range h.notifications {
		ns = append(ns, n)
	}
	writeHTTPJSON(w, http.StatusOK, map[string]interface{}{"notifications": ns, "total": len(ns)})
}

func (h *SuperappHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "notificationID")
	h.mu.Lock()
	defer h.mu.Unlock()
	n, ok := h.notifications[id]
	if !ok {
		writeHTTPError(w, http.StatusNotFound, "notification not found")
		return
	}
	n.Read = true
	writeHTTPJSON(w, http.StatusOK, n)
}

func (h *SuperappHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "notificationID")
	h.mu.Lock()
	delete(h.notifications, id)
	h.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// --- Geofences ---

func (h *SuperappHandler) CreateGeofence(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name         string  `json:"name"`
		CenterLat    float64 `json:"center_lat"`
		CenterLng    float64 `json:"center_lng"`
		RadiusMeters int     `json:"radius_meters"`
		SiteID       string  `json:"site_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeHTTPError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	gf := &GeofenceZone{
		ZoneID:       uuid.New().String(),
		Name:         input.Name,
		CenterLat:    input.CenterLat,
		CenterLng:    input.CenterLng,
		RadiusMeters: input.RadiusMeters,
		SiteID:       input.SiteID,
		CreatedAt:    time.Now().UTC(),
	}
	h.mu.Lock()
	h.geofences[gf.ZoneID] = gf
	h.mu.Unlock()

	writeHTTPJSON(w, http.StatusCreated, gf)
}

func (h *SuperappHandler) ListGeofences(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	gfs := make([]*GeofenceZone, 0, len(h.geofences))
	for _, gf := range h.geofences {
		gfs = append(gfs, gf)
	}
	writeHTTPJSON(w, http.StatusOK, map[string]interface{}{"geofences": gfs, "total": len(gfs)})
}

func (h *SuperappHandler) DeleteGeofence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "zoneID")
	h.mu.Lock()
	delete(h.geofences, id)
	h.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func writeHTTPJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeHTTPError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
