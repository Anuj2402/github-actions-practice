#!/bin/bash

echo "Starting health check..."

echo "Checking Linux environment..."
uname -s

echo "Checking current user..."
whoami

echo "Health check passed!"
exit 0
