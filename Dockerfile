FROM golang:1.22-alpine3.19 as builder

WORKDIR /src
COPY . ./
COPY logrotate.conf /src

RUN go build -o /app .

FROM alpine:3.19

RUN apk add logrotate


RUN mkdir /command-standby
COPY --from=builder /app /command-standby
COPY --from=builder /src/logrotate.conf /etc/logrotate.d/command-standby

CMD ["/command-standby/app"]
