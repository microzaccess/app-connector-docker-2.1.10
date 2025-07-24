# app-connector-docker-2.1.10

This project provides a Docker Compose setup for running the COSGrid App Connector. It includes all necessary configuration and shared files to get started quickly.

---

## 🚀 Getting Started

### ✅ Prerequisites

Ensure that you have [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) installed on your system.

You can check if Docker is installed by running:

```bash
docker --version
docker compose version
```

If Docker is not installed, follow the [official installation guide](https://docs.docker.com/get-docker/).

---

## 📦 Clone the Repository

Clone the repository using the following command:

```bash
git clone https://github.com/microzaccess/app-connector-docker-2.1.10.git
cd app-connector-docker-2.1.10
```

---

## 🐳 Start the Containers

Run the following command to start the services in detached mode:

```bash
docker compose up -d
```

This will build and run all the containers defined in the `docker-compose.yml` file.

---

## 🛑 Stopping the Containers

To stop and remove the containers:

```bash
docker compose down
```

---
