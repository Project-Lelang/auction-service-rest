FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Patch conf.yml: replace localhost with Docker service names
RUN sed -i '/^mysql:/,/^[a-z]/ s/host: localhost/host: mysql/' conf.yml && \
    sed -i '/^redis:/,/^[a-z]/ s/host: localhost/host: redis/' conf.yml

RUN CGO_ENABLED=0 go build -o /rest_api .

FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D nonroot
USER nonroot

WORKDIR /home/nonroot/

COPY --from=builder /app/conf.yml .
COPY --from=builder /app/storage ./storage
COPY --from=builder /rest_api .

EXPOSE 8080

CMD ["./rest_api"]