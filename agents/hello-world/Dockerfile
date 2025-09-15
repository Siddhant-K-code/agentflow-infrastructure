# Mock Agent Docker Image
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY agents/hello-world/main.go .
RUN go mod init mock-agent && go mod tidy
RUN go build -o agent main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/agent .
EXPOSE 3000
CMD ["./agent"]