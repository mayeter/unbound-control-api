#!/bin/sh
set -e

SOCKET="/opt/unbound/unbound.sock"

# Start Unbound in the background
unbound -d -c /opt/unbound/etc/unbound/unbound.conf &
UNBOUND_PID=$!

# Wait for the UNIX socket to be created and ready
echo "Waiting for $SOCKET to be ready..."
while [ ! -S "$SOCKET" ]; do
  sleep 1
done

echo "Unbound is up. Starting API..."
./unbound-control-api

# Optionally: Wait for Unbound to exit (if API ever stops)
wait $UNBOUND_PID