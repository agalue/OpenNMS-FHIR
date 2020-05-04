OpenNMS and FHIR
====

The goal here is a PoC to verify sample generation via NX-OS prior working with [Azure API for FHIR](https://azure.microsoft.com/en-us/services/azure-api-for-fhir/).

## Architecture

![Diagram](assets/FHIR-Architecture.png)

The Sample Generator uses the [telemetry_bis.proto](https://github.com/CiscoDevNet/nx-telemetry-proto/blob/master/telemetry_bis.proto) from Cisco, to generate the health metrics using Protobuf the same way a Nexus Switch would do to send streaming telemetry metrics via UDP.

That Protobuf definition is a very generic and vendor-agnostic definition for Telemetry data to send random numbers for Heart Rate and Steps to OpenNMS via Minion.

The reason for this is that OpenNMS already supports receiving and parsing NX-OS Telemetry metrics via UDP. To have a source of data we can use, I decided to reuse this pattern to have a constant stream of data comming into OpenNMS.

The [sample-generator](sample-generator) folder contains the code of it.

The generator sends the UDP data to a Minion, which in turn forwards the data via `Sink API` to OpenNMS.

OpenNMS receives the data via the NX-OS GPB Adapter, and use a simple Groovy Script to parse and persist the data.

It is crucial to notice this solution assumes the usage of `node-level` variables only. Also, the node-label of the sender (i.e., the one that represents the Sample Generator in OpenNMS) will be used as the `Device ID` for `FHIR`.

Then, the Kafka Producer pushes the Collection Sets to a Kafka Topic. From there, a simple receiver parses the data and show it to standard output.

## Run Test Environment

This lab was designed to run with [Docker](https://docker.io), so make sure you have it installed on your system.

You can start the lab using Docker Compose, from the root directory after checking out this repository on your machine:

```bash
docker-compose up -d
```

When OpenNMS is up and running, you should create a requisition with the node that represents the Sample Generator and associate it with the Location used for the Minion, for example:

```bash
LOCATION="Docker"
REQUISITION="Test"
DEVICE_ID="mock-device-001"
DEVICE_IP=$(docker container inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' generator)
cat <<EOF > generate-requisition.sh
#!/bin/sh
/opt/opennms/bin/provision.pl requisition add $REQUISITION
/opt/opennms/bin/provision.pl node add $REQUISITION $DEVICE_ID $DEVICE_ID
/opt/opennms/bin/provision.pl node set $REQUISITION $DEVICE_ID location $LOCATION
/opt/opennms/bin/provision.pl interface add $REQUISITION $DEVICE_ID $DEVICE_IP
/opt/opennms/bin/provision.pl interface set $REQUISITION $DEVICE_ID $DEVICE_IP snmp-primary N
/opt/opennms/bin/provision.pl requisition import $REQUISITION
EOF
chmod +x generate-requisition.sh
docker cp generate-requisition.sh opennms:/opt/opennms/bin/
docker exec -it opennms /opt/opennms/bin/generate-requisition.sh
rm -f generate-requisition.sh
```

## Verify the solution

Verify the output of the forwarder in DEBUG mode to see the messages sent by the Kafka Producer in Event Hub format:

```bash
docker-compose logs -f forwarder | grep "Sending message"
```

Or,

```bash
docker logs -f forwarder | grep "Sending message"
```

You must see messages like this:

```
2020/05/03 10:45:29 Sending message to Event Hub: {"body":{"deviceId":"mock-device","endDate":"2020-05-03T10:45:29-04:00","heartRate":"72","stepCount":"100"},"properties":{},"systemProperties":{}}
```

## Clean up

From the root directory after checking out this repository on your machine:

```bash
docker-compose down -v
```
