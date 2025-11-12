# ---------- Build Stage ----------
    FROM golang:1.24.2 AS builder

    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -a -o webhook-server ./cmd/webhook
    
# ---------- Runtime Stage ----------
    FROM gcr.io/distroless/base-debian11:nonroot
    
    WORKDIR /app
    
    COPY --from=builder /app/webhook-server /app/webhook-server
    
    USER nonroot:nonroot
    
    EXPOSE 443
    ENTRYPOINT ["/app/webhook-server"]