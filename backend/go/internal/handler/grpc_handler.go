package handler

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/sentinel-device-manager/backend/go/internal/model"
	"github.com/sentinel-device-manager/backend/go/internal/service"
	pb "github.com/sentinel-device-manager/backend/go/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	pb.UnimplementedDeviceServiceServer
	deviceService *service.DeviceService
}

func NewGRPCHandler(deviceService *service.DeviceService) *GRPCHandler {
	return &GRPCHandler{deviceService: deviceService}
}

func (h *GRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.SerialNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "serial_number is required")
	}

	input := service.CreateDeviceInput{
		SerialNumber:   req.SerialNumber,
		DeviceType:     protoToDeviceType(req.DeviceType),
		Model:          req.Model,
		SiteID:         req.SiteId,
		OrganizationID: req.OrganizationId,
		IPAddress:      req.IpAddress,
		MACAddress:     req.MacAddress,
		Config:         req.Config,
	}

	device, err := h.deviceService.CreateDevice(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create device: %v", err)
	}

	return &pb.RegisterResponse{
		Device: deviceToProto(device),
	}, nil
}

func (h *GRPCHandler) SendHeartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if req.DeviceId == "" {
		return nil, status.Error(codes.InvalidArgument, "device_id is required")
	}

	device, err := h.deviceService.GetDevice(ctx, req.DeviceId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "device not found: %v", err)
	}

	fw := device.FirmwareVersion
	_, err = h.deviceService.UpdateDevice(ctx, req.DeviceId, service.UpdateDeviceInput{
		FirmwareVersion: &fw,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update heartbeat: %v", err)
	}

	return &pb.HeartbeatResponse{
		ConfigUpdated: false,
		Config:        device.Config,
	}, nil
}

func (h *GRPCHandler) SendEvent(ctx context.Context, req *pb.EventRequest) (*pb.EventResponse, error) {
	if req.Event == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}
	if req.Event.DeviceId == "" {
		return nil, status.Error(codes.InvalidArgument, "device_id is required")
	}

	eventID := uuid.New().String()
	// In production, publish to SQS/event bus here
	_ = model.Event{
		EventID:   eventID,
		DeviceID:  req.Event.DeviceId,
		EventType: req.Event.EventType,
	}

	return &pb.EventResponse{
		EventId:      eventID,
		Acknowledged: true,
	}, nil
}

func (h *GRPCHandler) StreamFirmware(req *pb.FirmwareRequest, stream pb.DeviceService_StreamFirmwareServer) error {
	if req.DeviceId == "" {
		return status.Error(codes.InvalidArgument, "device_id is required")
	}
	if req.Version == "" {
		return status.Error(codes.InvalidArgument, "version is required")
	}

	// Simulate streaming firmware in 64KB chunks
	// In production, read from S3
	totalChunks := int32(3)
	dummyData := make([]byte, 64*1024)
	for i := int32(0); i < totalChunks; i++ {
		if err := stream.Send(&pb.FirmwareChunk{
			Version:       req.Version,
			ChunkIndex:    i,
			TotalChunks:   totalChunks,
			Data:          dummyData,
			ChecksumSha256: fmt.Sprintf("sha256:%s:chunk%d", req.Version, i),
		}); err != nil {
			return status.Errorf(codes.Internal, "stream chunk %d: %v", i, err)
		}
	}
	return nil
}

func (h *GRPCHandler) ReportFirmwareStatus(ctx context.Context, req *pb.FirmwareStatusRequest) (*pb.FirmwareStatusResponse, error) {
	if req.DeviceId == "" {
		return nil, status.Error(codes.InvalidArgument, "device_id is required")
	}

	// In production, persist firmware deployment status to DynamoDB
	return &pb.FirmwareStatusResponse{Acknowledged: true}, nil
}

func (h *GRPCHandler) GetConfig(ctx context.Context, req *pb.ConfigRequest) (*pb.ConfigResponse, error) {
	if req.DeviceId == "" {
		return nil, status.Error(codes.InvalidArgument, "device_id is required")
	}

	device, err := h.deviceService.GetDevice(ctx, req.DeviceId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "device not found: %v", err)
	}

	return &pb.ConfigResponse{
		Config:          device.Config,
		FirmwareVersion: device.FirmwareVersion,
	}, nil
}

func protoToDeviceType(dt pb.DeviceType) model.DeviceType {
	switch dt {
	case pb.DeviceType_CAMERA:
		return model.DeviceTypeCamera
	case pb.DeviceType_ACCESS_CONTROL:
		return model.DeviceTypeAccessControl
	case pb.DeviceType_ALARM:
		return model.DeviceTypeAlarm
	case pb.DeviceType_SENSOR:
		return model.DeviceTypeSensor
	default:
		return ""
	}
}

func deviceToProto(d *model.Device) *pb.Device {
	return &pb.Device{
		DeviceId:       d.DeviceID,
		SerialNumber:   d.SerialNumber,
		DeviceType:     deviceTypeToProto(d.DeviceType),
		Model:          d.Model,
		FirmwareVersion: d.FirmwareVersion,
		Status:         deviceStatusToProto(d.Status),
		SiteId:         d.SiteID,
		OrganizationId: d.OrganizationID,
		IpAddress:      d.IPAddress,
		MacAddress:     d.MACAddress,
		LastHeartbeat:  d.LastHeartbeat.Unix(),
		Config:         d.Config,
		CreatedAt:      d.CreatedAt.Unix(),
		UpdatedAt:      d.UpdatedAt.Unix(),
	}
}

func deviceTypeToProto(dt model.DeviceType) pb.DeviceType {
	switch dt {
	case model.DeviceTypeCamera:
		return pb.DeviceType_CAMERA
	case model.DeviceTypeAccessControl:
		return pb.DeviceType_ACCESS_CONTROL
	case model.DeviceTypeAlarm:
		return pb.DeviceType_ALARM
	case model.DeviceTypeSensor:
		return pb.DeviceType_SENSOR
	default:
		return pb.DeviceType_DEVICE_TYPE_UNSPECIFIED
	}
}

func deviceStatusToProto(s model.DeviceStatus) pb.DeviceStatus {
	switch s {
	case model.StatusOnline:
		return pb.DeviceStatus_ONLINE
	case model.StatusOffline:
		return pb.DeviceStatus_OFFLINE
	case model.StatusMaintenance:
		return pb.DeviceStatus_MAINTENANCE
	case model.StatusDecommissioned:
		return pb.DeviceStatus_DECOMMISSIONED
	default:
		return pb.DeviceStatus_DEVICE_STATUS_UNSPECIFIED
	}
}

// Ensure io import is used
var _ = io.EOF
var _ = time.UTC
