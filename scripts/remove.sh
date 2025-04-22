#!/bin/bash
set -e

# Variables from environment (provided during Batch submit)
STACK_NAME=${STACK_NAME:-default-stack}
ENVIRONMENT=${ENVIRONMENT:-dev}
REGISTRY_CREDENTIALS_ARN=${REGISTRY_CREDENTIALS_ARN:-none}
S3_BUCKET="ludofy"
AWS_REGION="ap-southeast-2"

echo "Removing stack: $STACK_NAME from environment: $ENVIRONMENT"

aws s3 rb "s3://$STACK_NAME-$ENVIRONMENT-avatars" --force
aws s3 rb "s3://$STACK_NAME-$ENVIRONMENT-images" --force

sam delete \
	--region "$AWS_REGION" \
	--stack-name "$STACK_NAME" \
	--s3-bucket "$S3_BUCKET" \
	--s3-prefix "$USER_ID/$BACKEND_ID/deployment" \
	--no-prompts
