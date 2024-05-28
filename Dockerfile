# Build environment
# -----------------
#FROM golang:1.22-alpine as builder
FROM cgr.dev/chainguard/go AS builder

#RUN apk add --no-cache gcc musl-dev

COPY src/ /app/
COPY go.mod go.sum /app/

WORKDIR /app

RUN go mod download
RUN go build -o app ./cmd


# Deployment environment
# ----------------------
#FROM alpine as runtime
FROM cgr.dev/chainguard/glibc-dynamic

#WORKDIR /app

COPY --from=builder /app/app /usr/bin/
#RUN ls -l
#RUN pwd
EXPOSE 8080

CMD ["/usr/bin/app"]
