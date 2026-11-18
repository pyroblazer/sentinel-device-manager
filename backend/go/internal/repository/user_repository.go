package repository

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/sentinel-device-manager/backend/go/internal/model"
)

// ErrUserNotFound is returned when no user has the given username.
var ErrUserNotFound = errors.New("user not found")

// UserRepository defines the persistence contract for operator accounts.
type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Put(ctx context.Context, user *model.User) error
}

type userRepo struct {
	client    *dynamodb.Client
	tableName string
}

// NewUserRepository returns a DynamoDB-backed user store keyed by username.
func NewUserRepository(client *dynamodb.Client, tableName string) UserRepository {
	return &userRepo{client: client, tableName: tableName}
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{Value: username},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, ErrUserNotFound
	}
	var user model.User
	if err := attributevalue.UnmarshalMap(out.Item, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Put(ctx context.Context, user *model.User) error {
	if user.UserID == "" {
		user.UserID = uuid.NewString()
	}
	now := time.Now().UTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	return err
}
