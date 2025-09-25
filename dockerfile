FROM golang:1.25.1-alpine3.22 AS builder
WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-compressdwarf -extldflags '-static'" -a -o /go-app-temp .


FROM alpine:3.22

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN apk add --no-cache tzdata
ENV TZ=Asia/Jakarta

WORKDIR /app

COPY --from=builder /go-app-temp /app/go-app-temp
COPY .env /app/.env
COPY cred.json /app/cred.json
RUN mkdir -p migration
COPY migration/* migration/

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

CMD ["/app/go-app-temp", "run", "-e", "production"]