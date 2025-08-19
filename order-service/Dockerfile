FROM golang:1.24-bullseye AS builder

WORKDIR /ecommerce-order-processing-system

COPY go.mod go.sum ./
RUN go mod download

COPY order-service/ order-service/
COPY pkg/ pkg/

WORKDIR /ecommerce-order-processing-system/order-service
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service .

FROM alpine:3.19

WORKDIR /root/

COPY --from=builder /ecommerce-order-processing-system/order-service/order-service .

EXPOSE 8080

CMD ["./order-service"]