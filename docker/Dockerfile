FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mongodb_exporter .

FROM alpine:3.21

COPY --from=builder /build/mongodb_exporter /mongodb_exporter

ENTRYPOINT ["/bin/sh", "-c", "/mongodb_exporter $CONFIG"]
