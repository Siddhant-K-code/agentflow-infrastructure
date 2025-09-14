# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o agentflow-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o agentctl ./cmd/agentctl

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/agentflow-server .
COPY --from=builder /app/agentctl .
COPY --from=builder /app/examples ./examples

EXPOSE 8080

CMD ["./agentflow-server"]