#!/bin/sh

echo "Waiting for DynamoDB Local..."
for i in $(seq 1 30); do
  if aws dynamodb list-tables --endpoint-url http://dynamodb:8000 --region us-east-1 2>/dev/null; then
    echo "DynamoDB is ready."
    break
  fi
  echo "Attempt $i: DynamoDB not ready, retrying in 2s..."
  sleep 2
done

echo "Current tables:"
aws dynamodb list-tables --endpoint-url http://dynamodb:8000 --region us-east-1

echo "Creating sentinel-devices table..."
aws dynamodb create-table \
  --table-name sentinel-devices \
  --attribute-definitions AttributeName=device_id,AttributeType=S \
  --key-schema AttributeName=device_id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --endpoint-url http://dynamodb:8000 \
  --region us-east-1 || echo "Table already exists, continuing..."

echo "Verifying table exists..."
aws dynamodb describe-table --table-name sentinel-devices --endpoint-url http://dynamodb:8000 --region us-east-1 --query 'Table.TableName'

echo "DynamoDB initialization complete."
