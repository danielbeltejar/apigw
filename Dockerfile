# Build environment
# -----------------
FROM golang:1.22-alpine as builder
WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY src/ ./
COPY go.mod go.sum ./

RUN go mod download
RUN go build -o app ./cmd


# Deployment environment
# ----------------------
FROM alpine as runtime
WORKDIR /app

COPY --from=builder /app/app .
RUN ls -l
RUN pwd
EXPOSE 8080

CMD ["./app"]