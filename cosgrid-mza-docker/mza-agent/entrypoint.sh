#!/bin/bash

# Fix: Set SUDO_USER env so installer writes to /home/dockeruser
export SUDO_USER=dockeruser

# Remove password for dockeruser (if ever needed inside container)
passwd -d dockeruser

# Run installer
echo "[INFO] Running cosgrid_installer..."
/opt/cosgrid/cosgrid_installer

# Start MZA Agent
echo "[INFO] Starting MZA Agent..."
/opt/cosgrid/cosgrid-mza-agent /home/dockeruser/.config/cosgrid &

# Keep container alive while agent runs
wait
