# Build
FROM golang:1.23 AS build

WORKDIR /build

COPY ./ .

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go build -o binary ./cmd/main.go

# Migrations
FROM amacneil/dbmate:2.14 AS dbmate

# Prod
FROM alpine:3.20

COPY ./migrations migrations

COPY --from=build /build/binary .
COPY --from=dbmate /usr/local/bin/dbmate .

EXPOSE 3000

ENTRYPOINT ["./binary"]
