package service

import (
	"context"
	"testing"
	"time"

	"github.com/sentinel-device-manager/backend/go/internal/model"
	"github.com/sentinel-device-manager/backend/go/internal/repository"
)

// mockRepo implements repository.DeviceRepository for testing.
type mockRepo struct {
	devices map[string]*model.Device
}

func newMockRepo() *mockRepo {
	return &mockRepo{devices: make(map[string]*model.Device)}
}

func (m *mockRepo) Create(_ context.Context, d *model.Device) error {
	m.devices[d.DeviceID] = d
	return nil
}

func (m *mockRepo) GetByID(_ context.Context, id string) (*model.Device, error) {
	d, ok := m.devices[id]
	if !ok {
		return nil, errNotFound
	}
	return d, nil
}

func (m *mockRepo) List(_ context.Context, _ repository.DeviceFilter) ([]model.Device, int, error) {
	result := make([]model.Device, 0, len(m.devices))
	for _, d := range m.devices {
		result = append(result, *d)
	}
	return result, len(result), nil
}

func (m *mockRepo) Update(_ context.Context, d *model.Device) error {
	m.devices[d.DeviceID] = d
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id string) error {
	delete(m.devices, id)
	return nil
}

type notFoundError struct{}

func (notFoundError) Error() string { return "not found" }

var errNotFound = notFoundError{}

func TestCreateDevice_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	input := CreateDeviceInput{
		SerialNumber:   "VKD-CAM-001",
		DeviceType:     model.DeviceTypeCamera,
		Model:          "D30",
		SiteID:         "site-001",
		OrganizationID: "org-001",
	}

	device, err := svc.CreateDevice(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if device.DeviceID == "" {
		t.Error("expected device_id to be set")
	}
	if device.SerialNumber != "VKD-CAM-001" {
		t.Errorf("expected serial VKD-CAM-001, got %s", device.SerialNumber)
	}
	if device.Status != model.StatusOnline {
		t.Errorf("expected status ONLINE, got %s", device.Status)
	}
	if device.FirmwareVersion != "0.0.0" {
		t.Errorf("expected firmware 0.0.0, got %s", device.FirmwareVersion)
	}
	if device.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
}

func TestCreateDevice_MissingSerial(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	input := CreateDeviceInput{
		DeviceType:     model.DeviceTypeCamera,
		SiteID:         "site-001",
		OrganizationID: "org-001",
	}

	_, err := svc.CreateDevice(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for missing serial_number")
	}
}

func TestCreateDevice_MissingSiteID(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	input := CreateDeviceInput{
		SerialNumber:   "VKD-CAM-001",
		DeviceType:     model.DeviceTypeCamera,
		OrganizationID: "org-001",
	}

	_, err := svc.CreateDevice(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for missing site_id")
	}
}

func TestCreateDevice_MissingOrgID(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	input := CreateDeviceInput{
		SerialNumber: "VKD-CAM-001",
		DeviceType:   model.DeviceTypeCamera,
		SiteID:       "site-001",
	}

	_, err := svc.CreateDevice(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for missing organization_id")
	}
}

func TestGetDevice_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	device, _ := svc.CreateDevice(context.Background(), CreateDeviceInput{
		SerialNumber:   "VKD-CAM-001",
		DeviceType:     model.DeviceTypeCamera,
		Model:          "D30",
		SiteID:         "site-001",
		OrganizationID: "org-001",
	})

	found, err := svc.GetDevice(context.Background(), device.DeviceID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.DeviceID != device.DeviceID {
		t.Errorf("expected device_id %s, got %s", device.DeviceID, found.DeviceID)
	}
}

func TestGetDevice_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	_, err := svc.GetDevice(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent device")
	}
}

func TestListDevices(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	_, _ = svc.CreateDevice(context.Background(), CreateDeviceInput{
		SerialNumber: "VKD-CAM-001", DeviceType: model.DeviceTypeCamera,
		SiteID: "site-001", OrganizationID: "org-001",
	})
	_, _ = svc.CreateDevice(context.Background(), CreateDeviceInput{
		SerialNumber: "VKD-SEN-002", DeviceType: model.DeviceTypeSensor,
		SiteID: "site-001", OrganizationID: "org-001",
	})

	devices, count, err := svc.ListDevices(context.Background(), repository.DeviceFilter{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 devices, got %d", count)
	}
	if len(devices) != 2 {
		t.Errorf("expected 2 devices in slice, got %d", len(devices))
	}
}

func TestUpdateDevice_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	device, _ := svc.CreateDevice(context.Background(), CreateDeviceInput{
		SerialNumber: "VKD-CAM-001", DeviceType: model.DeviceTypeCamera,
		SiteID: "site-001", OrganizationID: "org-001",
	})

	newModel := "D50"
	newFW := "1.2.3"
	updated, err := svc.UpdateDevice(context.Background(), device.DeviceID, UpdateDeviceInput{
		Model:           &newModel,
		FirmwareVersion: &newFW,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Model != "D50" {
		t.Errorf("expected model D50, got %s", updated.Model)
	}
	if updated.FirmwareVersion != "1.2.3" {
		t.Errorf("expected firmware 1.2.3, got %s", updated.FirmwareVersion)
	}
}

func TestDeleteDevice_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	device, _ := svc.CreateDevice(context.Background(), CreateDeviceInput{
		SerialNumber: "VKD-CAM-001", DeviceType: model.DeviceTypeCamera,
		SiteID: "site-001", OrganizationID: "org-001",
	})

	err := svc.DeleteDevice(context.Background(), device.DeviceID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = svc.GetDevice(context.Background(), device.DeviceID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestDecommissionDevice(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	device, _ := svc.CreateDevice(context.Background(), CreateDeviceInput{
		SerialNumber: "VKD-CAM-001", DeviceType: model.DeviceTypeCamera,
		SiteID: "site-001", OrganizationID: "org-001",
	})

	updated, err := svc.DecommissionDevice(context.Background(), device.DeviceID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != model.StatusDecommissioned {
		t.Errorf("expected DECOMMISSIONED, got %s", updated.Status)
	}
}

func TestUpdateDevice_UpdatedAtChanges(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	device, _ := svc.CreateDevice(context.Background(), CreateDeviceInput{
		SerialNumber: "VKD-CAM-001", DeviceType: model.DeviceTypeCamera,
		SiteID: "site-001", OrganizationID: "org-001",
	})
	originalUpdatedAt := device.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	newModel := "D50"
	updated, _ := svc.UpdateDevice(context.Background(), device.DeviceID, UpdateDeviceInput{Model: &newModel})
	if !updated.UpdatedAt.After(originalUpdatedAt) {
		t.Error("expected updated_at to change after update")
	}
}

func TestCreateDevice_DefaultConfig(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	device, err := svc.CreateDevice(context.Background(), CreateDeviceInput{
		SerialNumber: "VKD-CAM-001", DeviceType: model.DeviceTypeCamera,
		SiteID: "site-001", OrganizationID: "org-001",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if device.Config == nil {
		t.Error("expected config to be initialized")
	}
}

func TestCreateDevice_AllTypes(t *testing.T) {
	types := []model.DeviceType{
		model.DeviceTypeCamera,
		model.DeviceTypeAccessControl,
		model.DeviceTypeAlarm,
		model.DeviceTypeSensor,
	}

	for _, dt := range types {
		repo := newMockRepo()
		svc := NewDeviceService(repo)
		device, err := svc.CreateDevice(context.Background(), CreateDeviceInput{
			SerialNumber: "VKD-" + string(dt) + "-001",
			DeviceType:   dt,
			SiteID:       "site-001",
			OrganizationID: "org-001",
		})
		if err != nil {
			t.Errorf("failed to create device type %s: %v", dt, err)
		}
		if device.DeviceType != dt {
			t.Errorf("expected type %s, got %s", dt, device.DeviceType)
		}
	}
}

func TestDeleteDevice_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	err := svc.DeleteDevice(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for deleting nonexistent device")
	}
}

func TestHeartbeat_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	device, _ := svc.CreateDevice(context.Background(), CreateDeviceInput{
		SerialNumber: "VKD-CAM-001", DeviceType: model.DeviceTypeCamera,
		SiteID: "site-001", OrganizationID: "org-001",
	})

	updated, err := svc.Heartbeat(context.Background(), device.DeviceID, HeartbeatInput{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != model.StatusOnline {
		t.Errorf("expected ONLINE, got %s", updated.Status)
	}
	if updated.LastHeartbeat.IsZero() {
		t.Error("expected last_heartbeat to be set")
	}
}

func TestHeartbeat_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	_, err := svc.Heartbeat(context.Background(), "nonexistent", HeartbeatInput{})
	if err == nil {
		t.Fatal("expected error for nonexistent device")
	}
}

func TestCreateDevice_MissingDeviceType(t *testing.T) {
	repo := newMockRepo()
	svc := NewDeviceService(repo)

	_, err := svc.CreateDevice(context.Background(), CreateDeviceInput{
		SerialNumber: "VKD-CAM-001",
		SiteID:       "site-001",
		OrganizationID: "org-001",
	})
	if err == nil {
		t.Fatal("expected error for missing device_type")
	}
}
