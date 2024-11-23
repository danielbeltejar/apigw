# Build environment
# -----------------
FROM cgr.dev/chainguard/go as builder

WORKDIR /app

COPY ./ ./

RUN go mod download

RUN go build -o app ./cmd

# Deployment environment
# ----------------------
FROM cgr.dev/chainguard/glibc-dynamic

WORKDIR /app

COPY --from=builder /app/app /usr/bin/

EXPOSE 8080

CMD ["/usr/bin/app"]
