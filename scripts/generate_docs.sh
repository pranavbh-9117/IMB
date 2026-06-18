#!/bin/bash

# Generates Swaggo OpenAPI 2.0 documentation
# The docs package will be placed in cmd/server/docs (or root docs, depending on config)

echo "Generating Swagger documentation..."
$(go env GOPATH)/bin/swag init -g cmd/server/main.go --parseDependency --parseInternal --output ./docs

echo "Swagger documentation generated successfully in ./docs"
