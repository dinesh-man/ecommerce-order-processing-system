FROM golang:1.24-bullseye AS builder

WORKDIR /ecommerce-order-processing-system

COPY go.mod go.sum ./
RUN go mod download

COPY inventory-service/ inventory-service/
COPY pkg/ pkg/

WORKDIR /ecommerce-order-processing-system/inventory-service
RUN CGO_ENABLED=0 GOOS=linux go build -o inventory-service .

FROM alpine:3.19

WORKDIR /root/

COPY --from=builder /ecommerce-order-processing-system/inventory-service/inventory-service .

EXPOSE 8080

CMD ["./inventory-service"]