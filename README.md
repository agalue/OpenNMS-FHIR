OpenNMS and FHIR
====

The goal of this project is to have a working lab for testing purposes to generate heart rate metrics from an Apple Watch and forward them to [Azure API for FHIR](https://azure.microsoft.com/en-us/services/azure-api-for-fhir/) via [Event Hub](https://azure.microsoft.com/en-us/services/event-hubs/), as described in the following repository, used as a reference for the work described here:

https://github.com/microsoft/iomt-fhir

## Architecture

![Diagram](assets/FHIR-Architecture.png)

An Apple Watch application called [Graphite Heart](https://github.com/RangerRick/graphite-heart) was designed to send heart rate data over UDP using [Graphite](https://graphiteapp.org/) format. To parse this, a Graphite Adapter was implemented in OpenNMS that will be part of Horizon 26.1.0.
 
The Graphite Heart app sends data via UDP to a Minion, which in turn forwards the data via `Sink API` to OpenNMS.

OpenNMS receives the data via the Graphite Adapter, and use a simple Groovy Script to parse and persist the data.

It is crucial to notice this solution assumes the usage of `node-level` variables only. Also, the node-label of the sender (i.e., the one that represents the Sample Generator in OpenNMS) will be used as the `Device ID` for `FHIR`.

Then, the Kafka Producer pushes the Collection Sets to a Kafka Topic. From there, the Event Hub Forwarder takes the data, parse it, extract the desired list of metrics (specified via parameters or environment variables when using Docker), and if the metrics were extracted, it forwards the data to Azure Event Hub using the format suggested on:

https://github.com/microsoft/iomt-fhir/blob/master/docs/Configuration.md 

For example:

```json
{
  "Body": {
    "heartRate": "78",
    "endDate": "2020-04-29T10:46:01Z",
    "deviceId": "mock-device-001"
  },
  "Properties": {},
  "SystemProperties": {}
}
```

## Run Test Environment

This lab was designed to run with [Docker](https://docker.io), so make sure you have it installed on your system. I recommend 4 Cores with 8 GB of RAM at least, so make your Linux machine, Docker for Mac or Docker for Windows meet these requirements.

It is assume that an Event Hub instance already exists on Azure, with is a Shared Access Policy with Send capabilities, and you have access a connection string for it which is required by the `eventhub-forwarder` to be able to push the translated `CollectionSets` to Event Hub.

Once you have the connection string, declare an environment variable on your machine called `FORWARDER_EVENT_HUB_CONNECTION_STR` with its value, for example:

```bash
export FORWARDER_EVENT_HUB_CONNECTION_STR="Endpoint=sb://onmsfhir.servicebus.windows.net/;SharedAccessKeyName=send;SharedAccessKey=XXXXXXX;EntityPath=fhirhub"
```

Then, you can start the lab using Docker Compose, from the root directory after checking out this repository on your machine:

```bash
docker-compose up -d
```

Then, start the Graphite Heart App on your Apple Watch, configure the Minion IP (use the IP of the machine where Docker is running).

Once OpenNMS is up and running, it will be unable to forward the data until a node that represents the Apple Watch exists in the OpenNMS inventory.

For this check `karaf.log` to find out the IP of the device, you should see something like this:

```
2020-05-03T14:52:23,966 | WARN  | kafka-consumer-48 | AbstractAdapter                  | 312 - org.opennms.features.telemetry.protocols.adapters - 26.1.0.SNAPSHOT | Unable to determine collection agent from location=Docker and address=172.22.0.1
```

Rhen export an environment variable called `APPLE_WATCH_IP` and do the following:

```bash
LOCATION=Docker
REQUISITION=Test
DEVICE_ID=agalue-apple-watch
cat <<EOF > generate-requisition.sh
#!/bin/sh
/opt/opennms/bin/provision.pl requisition add $REQUISITION
/opt/opennms/bin/provision.pl node add $REQUISITION $DEVICE_ID $DEVICE_ID
/opt/opennms/bin/provision.pl node set $REQUISITION $DEVICE_ID location $LOCATION
/opt/opennms/bin/provision.pl interface add $REQUISITION $DEVICE_ID $APPLE_WATCH_IP
/opt/opennms/bin/provision.pl interface set $REQUISITION $DEVICE_ID $APPLE_WATCH_IP snmp-primary N
/opt/opennms/bin/provision.pl requisition import $REQUISITION
EOF
chmod +x generate-requisition.sh
docker cp generate-requisition.sh opennms:/opt/opennms/bin/
docker exec -it opennms /opt/opennms/bin/generate-requisition.sh
rm -f generate-requisition.sh
```

## Verify the solution

To make sure the solution works, inside the [eventhub-forwarder](eventhub-forwarder) directory, there is another directory called [consumer](eventhub-forwarder/consumer). There, you can find a small program that connects to Event Hub and read the messages.

This can be executed standalone, or with Docker:

```bash
docker run -it --rm \
  -e EVENT_HUB_CONNECTION_STR="Endpoint=sb://onmsfhir.servicebus.windows.net;SharedAccessKeyName=listen;SharedAccessKey=XXXXX;EntityPath=fhirhub" \
  agalue/fhir-sample-consumer
```

Note that in this case, the connection string must have `Listen` permissions.

## Clean up

From the root directory after checking out this repository on your machine:

```bash
docker-compose down -v
```

> Make sure the environment variables of the connection strings are set; otherwise the validation and `docker-compose` won't run.

