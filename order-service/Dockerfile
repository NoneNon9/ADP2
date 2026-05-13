
FROM golang:1.26-alpine AS builder
WORKDIR /app


COPY go.mod go.sum* ./
RUN go mod tidy

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service-app ./cmd/main.go


FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/order-service-app .


EXPOSE 8080
CMD ["./order-service-app"]