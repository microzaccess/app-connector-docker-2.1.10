#!/bin/bash
set -e

sleep 5

# Function to check if Redis is already running on port 6379
is_redis_running() {
    ss -ltn | grep -q ':6379'
}

# If Redis is running, just start the app
if is_redis_running; then
    echo "[+] Redis already running on port 6379. Starting app_connector only..."
    exec /home/app_connector
else
    echo "[+] Redis not detected. Starting Redis and app_connector..."
    /home/app_connector &  # Run app in background
    exec redis-server --protected-mode no  # Run Redis in foreground
fi
