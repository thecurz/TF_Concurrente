
FROM golang:1.20

# Set the working directory to /app
WORKDIR /app

# Copy go.mod and go.sum to /app
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the server source code into /app/server
COPY server/*.go ./server/

# Copy the utils and config packages
COPY utils ./utils
COPY config ./config

# Set the working directory to /app/server
WORKDIR /app/server

# Build the server binary
RUN go build -o server .

# Expose the TCP port
EXPOSE 8000

# Command to run when starting the container
CMD ["./server"]

