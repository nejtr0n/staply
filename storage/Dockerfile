FROM golang:1.11.4 as builder
RUN mkdir /build
ADD . /build/storage
WORKDIR /build/storage
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main pkg/storage.go

FROM alpine:latest
RUN addgroup -g 1000 appuser
RUN adduser -D -H -h /app -u 1000 -G appuser appuser
USER appuser
COPY --from=builder /build/storage/main /app/main
WORKDIR /app
CMD ["./main"]