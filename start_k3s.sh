#!/bin/bash

# Check if k3s is installed
if ! command -v k3s &> /dev/null; then
    echo "k3s is not installed. Installing now..."
    
    # Install k3s
    curl -sfL https://get.k3s.io | sh -
    
    if [ $? -ne 0 ]; then
        echo "Failed to install k3s. Please install it manually:"
        echo "curl -sfL https://get.k3s.io | sh -"
        exit 1
    fi
else
    echo "k3s is already installed."
fi

# Check if k3s service is running
if sudo systemctl is-active --quiet k3s; then
    echo "k3s is already running."
else
    echo "Starting k3s service..."
    sudo systemctl start k3s
    
    if [ $? -eq 0 ]; then
        echo "k3s service started successfully."
    else
        echo "Failed to start k3s service."
        exit 1
    fi
fi

# Wait a bit for k3s to be fully ready
echo "Waiting for k3s to be ready..."
sleep 10

# Verify k3s is working
if kubectl cluster-info &> /dev/null; then
    echo "k3s is ready and kubectl is working!"
    kubectl get nodes
else
    echo "k3s may not be fully ready yet. This can take up to 1-2 minutes."
    echo "You can check status with: kubectl cluster-info"
fi