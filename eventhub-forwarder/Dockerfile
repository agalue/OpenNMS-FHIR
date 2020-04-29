FROM golang:alpine AS builder
RUN mkdir /app && \
    echo "@edgecommunity http://nl.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache build-base git librdkafka-dev@edgecommunity
ADD ./ /app/
WORKDIR /app
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags musl -a -o eventhub-forwarder .

FROM alpine
ENV BOOTSTRAP_SERVERS="localhost:9092" \
    SOURCE_TOPIC="metrics" \
    GROUP_ID="opennms" \
    EVENT_HUB_CONNECTION_STR=""
RUN echo "@edgecommunity http://nl.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache bash librdkafka@edgecommunity && \
    addgroup -S onms && adduser -S -G onms onms
COPY --from=builder /app/eventhub-forwarder /eventhub-forwarder
COPY ./docker-entrypoint.sh /
USER onms
LABEL maintainer="Alejandro Galue <agalue@opennms.org>" \
      name="OpenNMS Kafka to Azure Event Hub Producer"
ENTRYPOINT [ "/docker-entrypoint.sh" ]
