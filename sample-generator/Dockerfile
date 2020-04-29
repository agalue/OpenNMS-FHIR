FROM golang:alpine AS builder
WORKDIR /app
ADD ./ /app/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags musl -a -o sample-generator .

FROM alpine
ENV TARGET="localhost:4444" FREQUENCY="30s"
RUN addgroup -S onms && adduser -S -G onms onms
COPY --from=builder /app/sample-generator /sample-generator
USER onms
LABEL maintainer="Alejandro Galue <agalue@opennms.org>" \
      name="OpenNMS Sample Generator for Health Data"
ENTRYPOINT /sample-generator -target "$TARGET" -frequency "$FREQUENCY"
