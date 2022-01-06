FROM golang:1.17.5-buster as builder

WORKDIR /src/app
COPY . .
RUN cat /etc/resolv.conf && go build .

FROM debian:buster

WORKDIR /src/app
COPY --from=builder /src/app .
ENV PORT=8080
EXPOSE ${PORT}
CMD ["/src/app/dandelion"]
