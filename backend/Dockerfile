# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build

# Set destination for COPY
WORKDIR /app

# Download any Go modules
COPY container_src/go.mod ./
RUN go mod download

# Copy container source code
COPY container_src/ ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /server /server
COPY --from=build /app/infas/schema.sql /app/schema.sql
WORKDIR /data
ENV DB_PATH=/data/moca_test.db SCHEMA_PATH=/app/schema.sql
EXPOSE 8080

# Run
CMD ["/server"]
