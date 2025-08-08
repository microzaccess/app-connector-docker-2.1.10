#!/bin/bash
# Input
INTERFACE="$1"
LOGSTASH_IP="51.79.141.108"
LOGSTASH_PORT="9996"
PIDFILE="/tmp/softflowd-${INTERFACE}.pid"
SOCKFILE="/tmp/softflowd-${INTERFACE}.sock"
FLOW_VERSION=5

if [ -z "$INTERFACE" ]; then
    echo "Interface name required as argument"
    exit 1
fi

# Create required directories
mkdir -p /var/run/softflowd/chroot

# Optional: Start syslog for better logging
service rsyslog start 2>/dev/null || true

# Stop previous instance
if [ -f "$PIDFILE" ]; then
    echo "Killing existing softflowd on $INTERFACE"
    kill "$(cat "$PIDFILE")" 2>/dev/null
    rm -f "$PIDFILE"
fi
[ -f "$SOCKFILE" ] && rm -f "$SOCKFILE"

# Start new instance
echo "Starting softflowd on $INTERFACE..."
softflowd -i "$INTERFACE" -n "${LOGSTASH_IP}:${LOGSTASH_PORT}" -v "$FLOW_VERSION" -m 1000 -t maxlife=30 -p "$PIDFILE" -c "$SOCKFILE"

# Check if process started
sleep 1
if [ -f "$PIDFILE" ] && ps -p "$(cat "$PIDFILE")" > /dev/null 2>&1; then
    echo "softflowd started successfully on $INTERFACE with PID $(cat "$PIDFILE")"
else
    echo "Failed to start softflowd"
    exit 1
fi