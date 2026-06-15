#!/usr/bin/env bash

set -eu

echo "Campus Agent MVP bootstrap"
echo "1. Review configs/config.yaml"
echo "2. Run: go test ./..."
echo "3. Run: docker compose -f deployments/docker-compose.yml up --build"
