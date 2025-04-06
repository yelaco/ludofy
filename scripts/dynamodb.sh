#!/bin/bash

# Define prefix for table names
PREFIX="slchess-dev"

# Output file
ENV_FILE="./configs/aws/dynamodb.env"

# Fetch all DynamoDB table names that start with PREFIX
TABLES=$(aws dynamodb list-tables --query "TableNames[?starts_with(@, '$PREFIX')]" --output text)

# Check if any tables were found
if [ -z "$TABLES" ]; then
	echo "No tables found with prefix $PREFIX"
	exit 1
fi

# Create or overwrite .env file
echo "# DynamoDB Table Names" >$ENV_FILE

# Convert table names to environment variables
for TABLE in $TABLES; do
	# Remove prefix
	TABLE_NAME=${TABLE#"$PREFIX-"}

	# Add underscore before each capital letter except the first one
	VAR_NAME=$(echo "$TABLE_NAME" | sed -E 's/([A-Z])/_\1/g' | sed 's/^_//' | tr '[:lower:]' '[:upper:]')

	# Append to .env file
	echo "${VAR_NAME}_TABLE_NAME=$TABLE" >>$ENV_FILE
done

echo ".env file generated successfully!"
