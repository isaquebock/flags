FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /api ./cmd/api

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /api /api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
