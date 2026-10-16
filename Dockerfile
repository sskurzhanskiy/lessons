FROM golang:1.26 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o habit ./cmd/api

FROM alpine AS runtime
WORKDIR /app
COPY --from=builder /app/habit .
RUN adduser -D -u 10001 habituser
USER 10001
CMD ["./habit"]
