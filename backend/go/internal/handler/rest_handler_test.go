package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sentinel-device-manager/backend/go/internal/model"
	"github.com/sentinel-device-manager/backend/go/internal/repository"
	"github.com/sentinel-device-manager/backend/go/internal/service"
)

type notFoundErr struct{}

func (e *notFoundErr) Error() string { return "not found" }

func newTestRepo() *testRepo {
	return &testRepo{devices: make(map[string]*model.Device)}
}

type testRepo struct {
	devices map[string]*model.Device
}

func (r *testRepo) Create(_ context.Context, d *model.Device) error {
	r.devices[d.DeviceID] = d
	return nil
}

func (r *testRepo) GetByID(_ context.Context, id string) (*model.Device, error) {
	d, ok := r.devices[id]
	if !ok {
		return nil, &notFoundErr{}
	}
	return d, nil
}

func (r *testRepo) List(_ context.Context, _ repository.DeviceFilter) ([]model.Device, int, error) {
	result := make([]model.Device, 0, len(r.devices))
	for _, d := range r.devices {
		result = append(result, *d)
	}
	return result, len(result), nil
}

func (r *testRepo) Update(_ context.Context, d *model.Device) error {
	r.devices[d.DeviceID] = d
	return nil
}

func (r *testRepo) Delete(_ context.Context, id string) error {
	delete(r.devices, id)
	return nil
}

func TestCreateDeviceEndpoint(t *testing.T) {
	repo := newTestRepo()
	svc := service.NewDeviceService(repo)
	h := NewRESTHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	body := map[string]interface{}{
		"serial_number":   "VKD-CAM-001",
		"device_type":     "CAMERA",
		"model":           "D30",
		"site_id":         "site-001",
		"organization_id": "org-001",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["serial_number"] != "VKD-CAM-001" {
		t.Errorf("expected serial_number VKD-CAM-001, got %v", resp["serial_number"])
	}
}

func TestGetDeviceEndpoint(t *testing.T) {
	repo := newTestRepo()
	svc := service.NewDeviceService(repo)
	h := NewRESTHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	// Create a device first
	device, _ := svc.CreateDevice(context.Background(), service.CreateDeviceInput{
		SerialNumber: "VKD-CAM-001", DeviceType: model.DeviceTypeCamera,
		SiteID: "site-001", OrganizationID: "org-001",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/"+device.DeviceID, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetDevice_NotFound(t *testing.T) {
	repo := newTestRepo()
	svc := service.NewDeviceService(repo)
	h := NewRESTHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/nonexistent", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestListDevicesEndpoint(t *testing.T) {
	repo := newTestRepo()
	svc := service.NewDeviceService(repo)
	h := NewRESTHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	_, _ = svc.CreateDevice(context.Background(), service.CreateDeviceInput{
		SerialNumber: "VKD-CAM-001", DeviceType: model.DeviceTypeCamera,
		SiteID: "site-001", OrganizationID: "org-001",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["total"] != float64(1) {
		t.Errorf("expected total 1, got %v", resp["total"])
	}
}

func TestDeleteDeviceEndpoint(t *testing.T) {
	repo := newTestRepo()
	svc := service.NewDeviceService(repo)
	h := NewRESTHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	device, _ := svc.CreateDevice(context.Background(), service.CreateDeviceInput{
		SerialNumber: "VKD-CAM-001", DeviceType: model.DeviceTypeCamera,
		SiteID: "site-001", OrganizationID: "org-001",
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/devices/"+device.DeviceID, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestHealthEndpoint(t *testing.T) {
	repo := newTestRepo()
	svc := service.NewDeviceService(repo)
	h := NewRESTHandler(svc)

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateDevice_InvalidBody(t *testing.T) {
	repo := newTestRepo()
	svc := service.NewDeviceService(repo)
	h := NewRESTHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateDeviceEndpoint(t *testing.T) {
	repo := newTestRepo()
	svc := service.NewDeviceService(repo)
	h := NewRESTHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	device, _ := svc.CreateDevice(context.Background(), service.CreateDeviceInput{
		SerialNumber: "VKD-CAM-001", DeviceType: model.DeviceTypeCamera,
		SiteID: "site-001", OrganizationID: "org-001",
	})

	body := map[string]interface{}{"model": "D50"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/devices/"+device.DeviceID, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
