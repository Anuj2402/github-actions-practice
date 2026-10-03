#!/bin/bash
set -e

IMAGE_NAME="chat-app-healthcheck-test"
NETWORK_NAME="healthcheck-net"
DB_CONTAINER="healthcheck-mysql"
APP_CONTAINER="healthcheck-app"

cleanup() {
  echo "Cleaning up..."
  docker rm -f "$APP_CONTAINER" "$DB_CONTAINER" >/dev/null 2>&1 || true
  docker network rm "$NETWORK_NAME" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "Building image..."
docker build -t "$IMAGE_NAME" .

echo "Creating network..."
docker network create "$NETWORK_NAME" >/dev/null

echo "Starting MySQL..."
docker run -d --name "$DB_CONTAINER" \
  --network "$NETWORK_NAME" \
  -e MYSQL_ROOT_PASSWORD=testpass \
  -e MYSQL_DATABASE=chatdb \
  -e MYSQL_USER=chatapp \
  -e MYSQL_PASSWORD=testpass \
  mysql:8.4 >/dev/null

echo "Waiting for MySQL to be ready..."
for i in $(seq 1 30); do
  if docker exec "$DB_CONTAINER" mysqladmin ping -h localhost -u root -ptestpass --silent 2>/dev/null; then
    echo "MySQL is ready."
    break
  fi
  sleep 2
done

echo "Starting app..."
docker run -d --name "$APP_CONTAINER" \
  --network "$NETWORK_NAME" \
  -p 8080:8080 \
  -e DB_HOST="$DB_CONTAINER" \
  -e DB_PORT=3306 \
  -e DB_USER=chatapp \
  -e DB_PASSWORD=testpass \
  -e DB_NAME=chatdb \
  "$IMAGE_NAME" >/dev/null

echo "Waiting for app to respond..."
for i in $(seq 1 15); do
  if curl -sf "http://localhost:8080/api/chat-history?from=healthcheck&to=healthcheck" >/dev/null; then
    echo "Health check passed!"
    exit 0
  fi
  sleep 2
done

echo "Health check failed: app did not respond in time."
docker logs "$APP_CONTAINER" || true
exit 1
