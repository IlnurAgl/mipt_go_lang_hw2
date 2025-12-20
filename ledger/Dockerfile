FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ledger ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

RUN GRPC_HEALTH_PROBE_VERSION=v0.4.24 && \
    wget -qO/bin/grpc-health-probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc-health-probe

WORKDIR /root/

COPY --from=builder /app/ledger .

COPY --from=builder /app/migrations ./migrations

EXPOSE 50051

CMD ["./ledger"]