FROM golang:1.18.2-buster as go-build
ENV GROUP_ID testgroup
ENV TOPIC test
ENV BOOTSTRAP_SERVERS localhost:9092
ENV LOGIQ_HOST logiq-flash
ENV LOGIQ_PORT 443

WORKDIR /usr/src/hauler-kafka

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
RUN apt-get update
RUN apt-get -y install git gcc librdkafka-dev
COPY go.mod go.sum ./
COPY . .
RUN go build -v -o /usr/local/bin/hauler-kafka ./...

FROM debian:buster-slim
RUN apt-get update
RUN apt-get install ca-certificates -y

COPY --from=go-build /usr/local/bin/hauler-kafka /usr/local/bin/hauler-kafka

CMD ["/usr/local/bin/hauler-kafka"]
