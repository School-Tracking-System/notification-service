# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy the modules needed
COPY proto/ ./proto/
COPY services/notification/ ./services/notification/

# Create a service-specific go.work to avoid loading other services
RUN printf "go 1.24.0\n\nuse (\n\t./proto\n\t./services/notification\n)\n" > go.work

# Build the application
WORKDIR /app/services/notification
RUN go build -o bin/api cmd/api/main.go cmd/api/module.go

# Final stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/services/notification/bin/api .
COPY --from=builder /app/services/notification/.env.template .env

EXPOSE 8085 9095

CMD ["./api"]
