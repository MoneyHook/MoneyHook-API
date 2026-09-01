FROM golang:1.27.0-alpine AS builder

WORKDIR /src

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY app/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/moneyhook .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /out/moneyhook /app/moneyhook
COPY --from=builder --chown=nonroot:nonroot /out/migrate /app/migrate

EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/moneyhook"]
