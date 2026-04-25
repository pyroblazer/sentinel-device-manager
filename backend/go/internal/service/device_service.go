// Package service implements the business logic layer for device management.
//
// The DeviceService provides CRUD operations for devices with validation,
// UUID generation, and lifecycle management (registration through decommission).
// It sits between the HTTP/gRPC handlers and the data repository.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sentinel-device-manager/backend/go/internal/model"
	"github.com/sentinel-device-manager/backend/go/internal/repository"
)

// DeviceService provides business logic for device lifecycle management.
// All operations require a context for cancellation and tracing support.
type DeviceService struct {
	repo repository.DeviceRepository
}

// NewDeviceService creates a DeviceService backed by the given repository.
func NewDeviceService(repo repository.DeviceRepository) *DeviceService {
	return &DeviceService{repo: repo}
}

// CreateDeviceInput contains the required and optional fields for registering a new device.
type CreateDeviceInput struct {
	SerialNumber   string            `json:"serial_number"`
	DeviceType     model.DeviceType  `json:"device_type"`
	Model          string            `json:"model"`
	SiteID         string            `json:"site_id"`
	OrganizationID string            `json:"organization_id"`
	IPAddress      string            `json:"ip_address"`
	MACAddress     string            `json:"mac_address"`
	Config         map[string]string `json:"config"`
}

// UpdateDeviceInput supports partial updates with pointer fields.
// Only non-nil fields will be applied to the existing device.
type UpdateDeviceInput struct {
	DeviceType      *model.DeviceType  `json:"device_type,omitempty"`
	Model           *string            `json:"model,omitempty"`
	FirmwareVersion *string            `json:"firmware_version,omitempty"`
	Status          *model.DeviceStatus `json:"status,omitempty"`
	SiteID          *string            `json:"site_id,omitempty"`
	IPAddress       *string            `json:"ip_address,omitempty"`
	Config          map[string]string  `json:"config,omitempty"`
}

// CreateDevice validates input, generates a UUID, sets defaults, and persists the device.
// Returns an error if required fields (serial_number, device_type, site_id, organization_id) are missing.
func (s *DeviceService) CreateDevice(ctx context.Context, input CreateDeviceInput) (*model.Device, error) {
	if input.SerialNumber == "" {
		return nil, fmt.Errorf("serial_number is required")
	}
	if input.DeviceType == "" {
		return nil, fmt.Errorf("device_type is required")
	}
	if input.SiteID == "" {
		return nil, fmt.Errorf("site_id is required")
	}
	if input.OrganizationID == "" {
		return nil, fmt.Errorf("organization_id is required")
	}

	now := time.Now().UTC()
	if input.Config == nil {
		input.Config = make(map[string]string)
	}

	device := &model.Device{
		DeviceID:        uuid.New().String(),
		SerialNumber:    input.SerialNumber,
		DeviceType:      input.DeviceType,
		Model:           input.Model,
		FirmwareVersion: "0.0.0",
		Status:          model.StatusOnline,
		SiteID:          input.SiteID,
		OrganizationID:  input.OrganizationID,
		IPAddress:       input.IPAddress,
		MACAddress:      input.MACAddress,
		LastHeartbeat:   now,
		Config:          input.Config,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("create device: %w", err)
	}
	return device, nil
}

// GetDevice retrieves a device by its unique identifier.
func (s *DeviceService) GetDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	device, err := s.repo.GetByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get device %s: %w", deviceID, err)
	}
	return device, nil
}

// ListDevices returns devices matching the given filter with a total count.
func (s *DeviceService) ListDevices(ctx context.Context, filter repository.DeviceFilter) ([]model.Device, int, error) {
	devices, count, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list devices: %w", err)
	}
	return devices, count, nil
}

// UpdateDevice applies a partial update to an existing device.
// Only non-nil fields in the input are applied; updated_at is always refreshed.
func (s *DeviceService) UpdateDevice(ctx context.Context, deviceID string, input UpdateDeviceInput) (*model.Device, error) {
	device, err := s.repo.GetByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get device %s: %w", deviceID, err)
	}

	if input.DeviceType != nil {
		device.DeviceType = *input.DeviceType
	}
	if input.Model != nil {
		device.Model = *input.Model
	}
	if input.FirmwareVersion != nil {
		device.FirmwareVersion = *input.FirmwareVersion
	}
	if input.Status != nil {
		device.Status = *input.Status
	}
	if input.SiteID != nil {
		device.SiteID = *input.SiteID
	}
	if input.IPAddress != nil {
		device.IPAddress = *input.IPAddress
	}
	if input.Config != nil {
		device.Config = input.Config
	}
	device.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, device); err != nil {
		return nil, fmt.Errorf("update device %s: %w", deviceID, err)
	}
	return device, nil
}

// DeleteDevice permanently removes a device. For GDPR-compliant removal,
// use DecommissionDevice first, then DeleteDevice after retention period.
func (s *DeviceService) DeleteDevice(ctx context.Context, deviceID string) error {
	if _, err := s.repo.GetByID(ctx, deviceID); err != nil {
		return fmt.Errorf("get device %s: %w", deviceID, err)
	}
	if err := s.repo.Delete(ctx, deviceID); err != nil {
		return fmt.Errorf("delete device %s: %w", deviceID, err)
	}
	return nil
}

// DecommissionDevice marks a device as decommissioned while preserving its record
// for audit purposes during the data retention period (GDPR Art.5(1)(e)).
func (s *DeviceService) DecommissionDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	status := model.StatusDecommissioned
	return s.UpdateDevice(ctx, deviceID, UpdateDeviceInput{Status: &status})
}
