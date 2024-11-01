# Blob Store Service

The `blob_store_service` is a microservice built in Go that manages file storage and retrieval. This service registers with Eureka for discovery and communicates with other services in a microservices architecture. It relies on an external `pkg` repository for utilities.

To view all services for this social media system, lets visit: `https://github.com/goddie9x?tab=repositories&q=social`

## Prerequisites

- Go 1.16+
- Docker and Docker Compose
- Git

## Setup

### 1. Clone the Repository

Clone the `blob_store_service` repository and its utility dependencies:

```bash
git clone https://github.com/goddie9x/social_blob_store_service.git
cd blob_store_service
```

### 2. Clone Utility Package

Clone the required `social_utils_go` package as a subdirectory in the project root:

```bash
git clone https://github.com/goddie9x/social_utils_go.git pkg
```

### 3. Configuration

Create a `config.yaml` file in the project root directory with the following configuration:

```yaml
databaseURL: "<your_postgres_connection_string>"
port: 6543
eurekaDiscoveryServerUrl: "http://localhost:8761/eureka"
eurekaAppName: "blob-service"
ipAddr: "127.0.0.1"
hostName: "localhost"
```

Ensure this file is not tracked in version control, as it contains sensitive information.

## Building the Service

To build the binary for Docker, use the following command from the project root:

```bash
GOOS=linux GOARCH=amd64 go build -o ./main ./cmd/server/main.go
```

This command creates a binary named `main` optimized for Linux (necessary for Docker containers).

## Running with Docker

1. **Build the Docker Image**:

   Create a `Dockerfile` in the root directory:

   ```dockerfile
   FROM alpine:latest
   RUN apk add --no-cache file
   WORKDIR /app
   COPY ./main ./config.yaml .
   EXPOSE 6543
   CMD ["./main"]
   ```

2. **Build and Run the Docker Container**:

   Use the following commands to build and run the Docker image:

   ```bash
   docker build -t blob-store-service .
   docker run -p 6543:6543 blob-store-service
   ```

   This will start the `blob_store_service` on port `6543`.

## Running with Docker Compose

You can integrate `blob_store_service` in a larger setup using Docker Compose. Below is a sample `docker-compose.yaml` snippet to add in a larger microservices environment.

```yaml
blob-store-service:
  image: blob-store-service
  build:
    context: .
  ports:
    - 6543:6543
  networks:
    - social-media-network
```

Start all services with Docker Compose:

```bash
docker-compose up --build
```

## Accessing the Service

Once the service is running, it will be accessible at `http://localhost:6543`.

---

### Services Overview

The `docker-compose.yaml` file defines several key services:
- **Kafka**: Handles messaging between microservices.
- **MongoDB**: Database for storing application data.
- **Oracle, Postgres**: Databases for specific services.
- **Elasticsearch and Kibana**: For search and data visualization.
- **Discovery Server (Eureka)**: Service registry for load balancing.
- **API Gateway**: Routes client requests to backend microservices.
  
Each service is assigned a port and added to the `social-media-network` to facilitate inter-service communication.

### Useful Commands

- **Stop Containers**: Use `docker-compose down` to stop all services and remove the containers.
- **Restart Containers**: Run `docker-compose restart` to restart the services without rebuilding the images.

This setup enables seamless orchestration of the social media microservices with an API Gateway for managing external client requests.

## Contributing

Contributions are welcome. Please clone this repository and submit a pull request with your changes. Ensure that your changes are well-tested and documented.

## License

This project is licensed under the MIT License. See `LICENSE` for more details.