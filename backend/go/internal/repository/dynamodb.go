// Package repository provides the data access layer for device persistence.
//
// The DynamoDB implementation stores devices in a single table with device_id
// as the partition key. Scan operations support filtering by type, status,
// site_id, and organization_id using DynamoDB filter expressions.
package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/sentinel-device-manager/backend/go/internal/model"
)

// DeviceRepository defines the persistence contract for device data.
type DeviceRepository interface {
	Create(ctx context.Context, device *model.Device) error
	GetByID(ctx context.Context, deviceID string) (*model.Device, error)
	List(ctx context.Context, filter DeviceFilter) ([]model.Device, int, error)
	Update(ctx context.Context, device *model.Device) error
	Delete(ctx context.Context, deviceID string) error
}

// DeviceFilter provides optional criteria for narrowing device list queries.
type DeviceFilter struct {
	DeviceType     *model.DeviceType
	Status         *model.DeviceStatus
	SiteID         *string
	OrganizationID *string
}

type deviceRepo struct {
	client    *dynamodb.Client
	tableName string
}

func NewDeviceRepository(client *dynamodb.Client, tableName string) DeviceRepository {
	return &deviceRepo{client: client, tableName: tableName}
}

func (r *deviceRepo) Create(ctx context.Context, device *model.Device) error {
	if device.DeviceID == "" {
		device.DeviceID = uuid.New().String()
	}
	item, err := attributevalue.MarshalMap(device)
	if err != nil {
		return fmt.Errorf("marshal device: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(device_id)"),
	})
	if err != nil {
		return fmt.Errorf("put device: %w", err)
	}
	return nil
}

func (r *deviceRepo) GetByID(ctx context.Context, deviceID string) (*model.Device, error) {
	resp, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"device_id": &types.AttributeValueMemberS{Value: deviceID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}
	if resp.Item == nil {
		return nil, ErrNotFound
	}
	var device model.Device
	if err := attributevalue.UnmarshalMap(resp.Item, &device); err != nil {
		return nil, fmt.Errorf("unmarshal device: %w", err)
	}
	return &device, nil
}

func (r *deviceRepo) List(ctx context.Context, filter DeviceFilter) ([]model.Device, int, error) {
	expr, attrNames, attrValues, err := r.buildFilterExpression(filter)
	if err != nil {
		return nil, 0, err
	}

	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}
	if expr != "" {
		input.FilterExpression = aws.String(expr)
		input.ExpressionAttributeNames = attrNames
		input.ExpressionAttributeValues = attrValues
	}

	resp, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, 0, fmt.Errorf("scan devices: %w", err)
	}

	devices := make([]model.Device, 0, len(resp.Items))
	for _, item := range resp.Items {
		var d model.Device
		if err := attributevalue.UnmarshalMap(item, &d); err != nil {
			return nil, 0, fmt.Errorf("unmarshal device: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, len(devices), nil
}

func (r *deviceRepo) Update(ctx context.Context, device *model.Device) error {
	item, err := attributevalue.MarshalMap(device)
	if err != nil {
		return fmt.Errorf("marshal device: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("update device: %w", err)
	}
	return nil
}

func (r *deviceRepo) Delete(ctx context.Context, deviceID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"device_id": &types.AttributeValueMemberS{Value: deviceID},
		},
	})
	if err != nil {
		return fmt.Errorf("delete device: %w", err)
	}
	return nil
}

func (r *deviceRepo) buildFilterExpression(f DeviceFilter) (string, map[string]string, map[string]types.AttributeValue, error) {
	var conditions []string
	attrNames := make(map[string]string)
	attrValues := make(map[string]types.AttributeValue)

	if f.DeviceType != nil {
		attrNames["#dt"] = "device_type"
		attrValues[":dt"] = &types.AttributeValueMemberS{Value: string(*f.DeviceType)}
		conditions = append(conditions, "#dt = :dt")
	}
	if f.Status != nil {
		attrNames["#st"] = "status"
		attrValues[":st"] = &types.AttributeValueMemberS{Value: string(*f.Status)}
		conditions = append(conditions, "#st = :st")
	}
	if f.SiteID != nil {
		attrNames["#si"] = "site_id"
		attrValues[":si"] = &types.AttributeValueMemberS{Value: *f.SiteID}
		conditions = append(conditions, "#si = :si")
	}
	if f.OrganizationID != nil {
		attrNames["#oi"] = "organization_id"
		attrValues[":oi"] = &types.AttributeValueMemberS{Value: *f.OrganizationID}
		conditions = append(conditions, "#oi = :oi")
	}

	expr := ""
	for i, c := range conditions {
		if i > 0 {
			expr += " AND "
		}
		expr += c
	}
	return expr, attrNames, attrValues, nil
}
