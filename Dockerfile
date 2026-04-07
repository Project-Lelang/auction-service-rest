FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY dev.env . 

COPY . . 

RUN CGO_ENABLED=0 go build -o /rest_api ./cmd/rest_api

FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D nonroot
USER nonroot

WORKDIR /home/nonroot/

COPY --from=builder /app/dev.env . 

COPY --from=builder /rest_api .

EXPOSE 8080

CMD ["./rest_api"]