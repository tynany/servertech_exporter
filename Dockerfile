FROM golang:1.14
WORKDIR /go/src/github.com/tynany/servertech_exporter
COPY . /go/src/github.com/tynany/servertech_exporter
RUN make setup_promu
RUN ./promu build
RUN ls -lah

FROM alpine:3.12.3
WORKDIR /app
COPY --from=0 /go/src/github.com/tynany/servertech_exporter/servertech_exporter .
EXPOSE 9783
CMD ["./servertech_exporter", "--web.certificate=/server.crt","--web.key=/server.key"]
