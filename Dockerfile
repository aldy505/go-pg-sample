FROM golang:1.17.5-buster as builder

WORKDIR /src/app
COPY . .
RUN go mod download && go build -o main .

FROM debian:buster

WORKDIR /src/app
COPY --from=builder /src/app .
ENV PORT=8080
EXPOSE ${PORT}
ENTRYPOINT ["/src/app/main"]