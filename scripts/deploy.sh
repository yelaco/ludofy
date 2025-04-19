#!/bin/bash
set -e

# Variables from environment (provided during Batch submit)
STACK_NAME=${STACK_NAME:-default-stack}
GIT_BRANCH=${GIT_BRANCH:-main}
ENVIRONMENT=${ENVIRONMENT:-dev}
GIT_REPO="https://github.com/yelaco/ludofy.git"
S3_BUCKET="ludofy"
AWS_REGION="ap-southeast-2"

echo "Deploying stack: $STACK_NAME from branch: $GIT_BRANCH into environment: $ENVIRONMENT"

# Clone GitHub repo
git clone --branch "$GIT_BRANCH" "$GIT_REPO" repo

mkdir workspace

cp -r repo/cmd repo/internal repo/pkg repo/go.* workspace/
cd workspace

aws s3 cp "s3://ludofy/$USER_ID/$BACKEND_ID/templates/" ./templates --recursive

mv ./templates/template.yaml .
export PATH=$PATH:/usr/local/go/bin

# Build project
sam build

# Deploy project
sam deploy \
	--stack-name "$STACK_NAME" \
	--region "$AWS_REGION" \
	--parameter-overrides "ServerImageUri=$SERVER_IMAGE_URI" \
	--s3-bucket "$S3_BUCKET" \
	--s3-prefix "$USER_ID/$BACKEND_ID/deployment" \
	--capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND \
	--no-confirm-changeset \
	--on-failure DELETE
