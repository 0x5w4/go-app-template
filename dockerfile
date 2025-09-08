FROM golang:1.24-alpine AS builder
WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-w -s" -a -o /m3s-reeng .


FROM alpine:3.20

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN apk add --no-cache tzdata
ENV TZ=Asia/Jakarta

WORKDIR /app

COPY --from=builder /m3s-reeng /app/m3s-reeng
COPY .env /app/.env
COPY cred.json /app/cred.json
RUN mkdir -p internal/adapter/repository/mysql/db/migration
COPY internal/adapter/repository/mysql/db/migration/* internal/adapter/repository/mysql/db/migration/

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

CMD ["/app/m3s-reeng", "run", "-e", "production"]