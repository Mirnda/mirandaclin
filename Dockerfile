FROM golang:1.26-alpine AS builder

ARG APP_ENV=development

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN if [ "$APP_ENV" != "production" ]; then \
  go install github.com/swaggo/swag/cmd/swag@v1.16.4; \
  fi
RUN if [ "$APP_ENV" != "production" ]; then \
  swag init -g cmd/api/main.go -o docs; \
  fi

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/mirandaclin ./cmd/api


FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/mirandaclin .
EXPOSE 8080
CMD ["./mirandaclin"]
