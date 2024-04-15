FROM golang:1.22-alpine as buildbase

RUN apk add git build-base

WORKDIR /go/src/github.com/rarimo/airdrop-svc
COPY vendor .
COPY . .

RUN GOOS=linux go build  -o /usr/local/bin/airdrop-svc /go/src/github.com/rarimo/airdrop-svc


FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/airdrop-svc /usr/local/bin/airdrop-svc
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["airdrop-svc"]
