FROM golang:1.24-alpine3.21 as builder

WORKDIR /src
COPY . ./

RUN go build -o /app .

FROM alpine:3.21

RUN mkdir /command-standby
COPY --from=builder /app /command-standby

CMD ["/command-standby/app"]
