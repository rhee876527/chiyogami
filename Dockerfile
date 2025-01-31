# Start with the official Golang image
FROM golang:1.23.1-alpine3.20

# Install necessary packages
RUN apk add --no-cache git

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod file first and create go.sum inside the container
COPY go.mod ./

# Initialize Go modules and download dependencies (this will generate go.sum)
RUN go mod download

# Copy the remaining source code
COPY . .

# Ensure build deps are in place
RUN go get checksum-calculator

#Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Use non-root user
RUN addgroup -g 919 appgroup && adduser -u 919 --disabled-password -G appgroup appuser
RUN chown -R appuser:appgroup /app

USER appuser

# Run in production mode
#ENV GIN_MODE=release

# Command to run the executable
CMD ["./main"]
