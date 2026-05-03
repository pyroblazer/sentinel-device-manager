// Package handler implements the HTTP/gRPC transport layer for the device API.
//
// The RESTHandler exposes JSON-over-HTTP endpoints using the chi router,
// and the GRPCHandler implements the protobuf DeviceServiceServer interface.
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sentinel-device-manager/backend/go/internal/model"
	"github.com/sentinel-device-manager/backend/go/internal/repository"
	"github.com/sentinel-device-manager/backend/go/internal/service"
)

// RESTHandler handles HTTP requests for device management.
type RESTHandler struct {
	deviceService *service.DeviceService
}

// NewRESTHandler creates a REST handler backed by the given DeviceService.
func NewRESTHandler(deviceService *service.DeviceService) *RESTHandler {
	return &RESTHandler{deviceService: deviceService}
}

// RegisterRoutes mounts device CRUD routes on the chi router at /api/v1/devices.
func (h *RESTHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/devices", func(r chi.Router) {
		r.Get("/", h.ListDevices)
		r.Post("/", h.CreateDevice)
		r.Route("/{deviceID}", func(r chi.Router) {
			r.Get("/", h.GetDevice)
			r.Put("/", h.UpdateDevice)
			r.Delete("/", h.DeleteDevice)
			r.Post("/heartbeat", h.Heartbeat)
		})
	})
}

// CreateDevice handles POST /api/v1/devices. Validates and registers a new device.
func (h *RESTHandler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var input service.CreateDeviceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	device, err := h.deviceService.CreateDevice(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, device)
}

// GetDevice handles GET /api/v1/devices/{deviceID}. Returns device details.
func (h *RESTHandler) GetDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "deviceID")
	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "device_id is required")
		return
	}

	device, err := h.deviceService.GetDevice(r.Context(), deviceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}

	writeJSON(w, http.StatusOK, device)
}

// ListDevices handles GET /api/v1/devices with optional query filters.
func (h *RESTHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	filter := repository.DeviceFilter{
		DeviceType:     parseDeviceType(r.URL.Query().Get("type")),
		Status:         parseDeviceStatus(r.URL.Query().Get("status")),
		SiteID:         parseStringPtr(r.URL.Query().Get("site_id")),
		OrganizationID: parseStringPtr(r.URL.Query().Get("organization_id")),
	}

	devices, count, err := h.deviceService.ListDevices(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list devices")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"devices": devices,
		"total":   count,
		"page":    page,
		"limit":   limit,
	})
}

// UpdateDevice handles PUT /api/v1/devices/{deviceID} with partial updates.
func (h *RESTHandler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "deviceID")
	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "device_id is required")
		return
	}

	var input service.UpdateDeviceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	device, err := h.deviceService.UpdateDevice(r.Context(), deviceID, input)
	if err != nil {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}

	writeJSON(w, http.StatusOK, device)
}

// DeleteDevice handles DELETE /api/v1/devices/{deviceID}.
func (h *RESTHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "deviceID")
	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "device_id is required")
		return
	}

	if err := h.deviceService.DeleteDevice(r.Context(), deviceID); err != nil {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Heartbeat handles POST /api/v1/devices/{deviceID}/heartbeat.
// Updates the device's last_heartbeat timestamp and returns the updated device.
func (h *RESTHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "deviceID")
	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "device_id is required")
		return
	}

	var input service.HeartbeatInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// Accept empty body for simple heartbeat pings
		input = service.HeartbeatInput{}
	}

	device, err := h.deviceService.Heartbeat(r.Context(), deviceID, input)
	if err != nil {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}

	writeJSON(w, http.StatusOK, device)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func parseDeviceType(s string) *model.DeviceType {
	if s == "" {
		return nil
	}
	dt := model.DeviceType(s)
	return &dt
}

func parseDeviceStatus(s string) *model.DeviceStatus {
	if s == "" {
		return nil
	}
	st := model.DeviceStatus(s)
	return &st
}

func parseStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
