
FROM golang:1.26-alpine AS builder
WORKDIR /app


COPY go.mod go.sum* ./
RUN go mod tidy

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service-app ./cmd/main.go


FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/payment-service-app .


EXPOSE 8081
CMD ["./payment-service-app"]