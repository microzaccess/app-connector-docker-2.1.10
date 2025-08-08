#!/bin/bash
set -euo pipefail

log() {
    echo -e "\033[1;34m[INFO]\033[0m $1"
}

error_exit() {
    echo -e "\033[1;31m[ERROR]\033[0m $1" >&2
    exit 1
}

# Step 1: Update package index
log "Updating apt package index..."
sudo apt-get update || error_exit "Failed to update apt index."

# Step 2: Install dependencies
log "Installing required packages..."
sudo apt-get install -y ca-certificates curl gnupg lsb-release || error_exit "Failed to install dependencies."

# Step 3: Create keyring directory
log "Creating keyring directory..."
sudo install -m 0755 -d /etc/apt/keyrings || error_exit "Failed to create /etc/apt/keyrings directory."

# Step 4: Download Docker’s official GPG key
log "Downloading Docker GPG key..."
if ! sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc; then
    error_exit "Failed to download Docker GPG key."
fi

# Step 5: Set correct permissions
log "Setting permissions on GPG key..."
sudo chmod a+r /etc/apt/keyrings/docker.asc || error_exit "Failed to set permissions on GPG key."

# Step 6: Add Docker repository
UBUNTU_CODENAME=$(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
ARCH=$(dpkg --print-architecture)
REPO_LINE="deb [arch=${ARCH} signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu ${UBUNTU_CODENAME} stable"

log "Adding Docker repository for ${UBUNTU_CODENAME} (${ARCH})..."
echo "$REPO_LINE" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null || error_exit "Failed to add Docker repo."

# Step 7: Update apt package index again
log "Updating apt package index again..."
sudo apt-get update || error_exit "Failed to update apt index after adding Docker repo."

# Step 8: Install Docker packages
log "Installing Docker components..."
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin || error_exit "Failed to install Docker packages."

# Step 9: Verify installation
log "Verifying Docker installation..."
if ! sudo docker --version >/dev/null 2>&1; then
    error_exit "Docker installation appears to have failed."
fi

log "Docker installed successfully!"
