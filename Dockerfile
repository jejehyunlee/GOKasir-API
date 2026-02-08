# Single stage build - lebih reliable untuk Railway
FROM golang:1.23-alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
