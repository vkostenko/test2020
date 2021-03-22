FROM golang:1.14 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o hellofresh .

FROM scratch
WORKDIR /root/
COPY --from=builder /app/hellofresh .
ENTRYPOINT ["./hellofresh"]
