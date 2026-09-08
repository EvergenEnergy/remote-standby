FROM golang:1.27-alpine3.24 as builder

WORKDIR /src
COPY . ./

RUN go build -o /app .

FROM alpine:3.24

RUN mkdir /command-standby
COPY --from=builder /app /command-standby

CMD ["/command-standby/app"]
