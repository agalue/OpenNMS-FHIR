Event Hub Forwarder
====

This component consumes Protobuf Messages from Kafka, extracting the [CollectionSet](https://github.com/OpenNMS/opennms/blob/master/features/kafka/producer/src/main/proto/collectionset.proto) sent by the Kafka Producer feature in OpenNMS with the metrics. If there is Health Data present, generates valid message that could be used with `JsonPathContentTemplate` before consuming them for the FHIR server.

This is a stand alone project that can be tested without OpenNMS. The `docker-compose.yaml` outlines all the components required. It creates a container for Zookeeper, Kafka, the forwarder itself, a [generator](./generator) that emulates the Kafka Producer feature injecting mock `CollectionSet` data to Kafka, and a [consumer](./consumer) that reads data from the Event Hub for verification purposes.

Of course, to use the compose file, the `forwarder` requires the existence of the `FORWARDER_EVENT_HUB_CONNECTION_STR` environment variable on your machine with the connection string with `Send` permissions, and the `CONSUMER_EVENT_HUB_CONNECTION_STR` environment variable with the conection string with `Listen` permissions.

Once the variables exist, you could start the test with:

```bash
docker-compose -d up
```

All the components were implemented in [Go](https://golang.org).

The `Dockerfile` for each component is available, and the images where generated in the following way:

```bash
docker build -t agalue/fhir-eventhub-forwarder .
docker build -t agalue/fhir-sample-generator-kafka -f Dockerfile.generator .
docker build -t agalue/fhir-sample-consumer -f Dockerfile.consumer .

docker push agalue/fhir-eventhub-forwarder
docker push agalue/fhir-sample-generator-kafka
docker push agalue/fhir-sample-consumer
```

To use the programs locally without `Docker`, you must have a `Go` compiler installed on your machine.