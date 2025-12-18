# Build stage
FROM golang:1.25.5-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY scraper.go ./
COPY keywords-news.txt ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scraper ./scraper.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/scraper .

# Copy keywords-news.txt
COPY --from=builder /app/keywords-news.txt .

# Run the scraper
CMD ["./scraper"]

