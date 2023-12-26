#docker build -t go-credit .
#docker run -dit --name go-credit -p 5000:5000 go-credit

FROM golang:1.21 As builder

RUN apt-get update && apt-get install bash && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go build -o go-credit -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-credit .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/go-credit"]