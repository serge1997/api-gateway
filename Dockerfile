FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -v -o ./app ./cmd/gateway/

FROM alpine:latest
COPY --from=builder /app/app /app
CMD ["./app"]