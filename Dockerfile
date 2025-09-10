# Stage 1: Create Frontend Assets
FROM node:22-alpine AS frontend
WORKDIR /app

# Copy npm configs and tailwind config
COPY frameworks/package.json frameworks/package-lock.json ./

# Set npm start dir
WORKDIR /app/frameworks

# Install dependencies
RUN npm ci

# Generate tailwind CSS output
COPY public /app/public
COPY frameworks/tailwind.config.js ./
RUN echo -e "@tailwind base;\n@tailwind components;\n@tailwind utilities;" > /app/public/tailwind.css
RUN npx tailwindcss -i /app/public/tailwind.css -o /app/public/output.css --minify

# Stage 2: Build Go Application
FROM golang:1.25-alpine AS build
WORKDIR /app

# Go files and dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy src
COPY . .

# Gorm's sqlite requires CGO and additional libraries
RUN apk add --no-cache gcc musl-dev

# Build Go application
RUN go build -o main .

# Stage 3: Run the application
FROM alpine:3.22
WORKDIR /app

# Include license
COPY LICENSE .

# Copy Go binary and assets
COPY --from=build /app/main .
COPY --from=build /app/public ./public

# Fetch generated Tailwind output
COPY --from=frontend /app/public/output.css ./public

# Copy entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Use su-exec to run as non-root user
RUN apk add --no-cache su-exec

# Expose port
EXPOSE 8000

# Tag docker environments
ENV DOCKER_ENV=1

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O - http://localhost:8000/health | grep -q '"status":"ok".*"db_status":"ok"' || exit 1

# Create entrypoint
ENTRYPOINT ["/entrypoint.sh"]
