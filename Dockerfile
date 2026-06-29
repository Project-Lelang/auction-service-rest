FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /rest_api .
RUN CGO_ENABLED=0 go build -o /migrate ./cmd/migrate
RUN CGO_ENABLED=0 go build -o /seed ./cmd/seed

FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D nonroot
USER nonroot

WORKDIR /home/nonroot/

COPY --from=builder /rest_api .
COPY --from=builder /migrate .
COPY --from=builder /seed .
COPY --from=builder /app/migration ./migration

EXPOSE 8080

CMD ["./rest_api"]