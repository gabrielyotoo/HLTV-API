# HLTV-API

HLTV scraper API based on Colly

## Docker Deployment

### Build and Run with Docker

```bash
# Build the Docker image
docker build -t hltv-scraper .

# Run the container
docker run --rm hltv-scraper
```

### Using Docker Compose

```bash
# Build and run
docker-compose up

# Run in detached mode
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the container
docker-compose down
```

### Local Development

```bash
# Run locally
go run scraper.go
```

## Configuration

Edit `keywords.txt` to customize which keywords trigger special alerts. One keyword per line, case-insensitive.
