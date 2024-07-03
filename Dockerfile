# Build stage
FROM golang:1.22 AS builder

# Enable Go modules and cgo
ENV CGO_ENABLED=1
ENV GO111MODULE=on

# Set the current working directory inside the container
WORKDIR /go/src/job_sender

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the working Directory inside the container
COPY . .

# Copy the templates directory
COPY templates templates

# Build the Go app
RUN go build -a -installsuffix cgo -o /go/bin/job_sender .

# Start a new stage from Ubuntu 22.04
FROM ubuntu:22.04

# Install runtime dependencies and any other necessary packages, including CA certificates
RUN apt-get update && \
    apt-get install -y \
    ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install curl
RUN apt-get update && apt-get install -y curl

# Install Node.js and npm
RUN curl -sL https://deb.nodesource.com/setup_20.x | bash -
RUN apt-get install -y nodejs
    
# Set a working directory
WORKDIR /app

# Install Firebase Admin SDK
RUN npm install firebase-admin

# Install Google Cloud SDK
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg add - && \
    apt-get update -y && apt-get install google-cloud-sdk -y

# Copy the pre-built binary file from the previous stage
COPY --from=builder /go/bin/job_sender /go/bin/job_sender

# Copy the templates directory from the builder stage
COPY --from=builder /go/src/job_sender/templates /templates

# Document that the service listens on port 8080.
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["/go/bin/job_sender"]