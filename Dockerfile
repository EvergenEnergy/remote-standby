FROM golang:1.23-alpine3.20 as builder

WORKDIR /src
COPY . ./

RUN go build -o /app .

FROM alpine:3.20

RUN mkdir /command-standby
COPY --from=builder /app /command-standby

CMD ["/command-standby/app"]
