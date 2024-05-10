FROM golang:1.22-alpine3.19 as builder

WORKDIR /src
COPY . ./

RUN go build -o /app .

FROM alpine:3.19

RUN mkdir /command-standby
COPY --from=builder /app /command-standby

CMD ["/command-standby/app"]
