Health Data Sample Generator using NX-OS
====

Uses the [telemetry_bis.proto](https://github.com/CiscoDevNet/nx-telemetry-proto/blob/master/telemetry_bis.proto) from Cisco, to generate the health metrics using Protobuf the same way a Nexus Switch would do to send streaming telemetry metrics via UDP.

The `Dockerfile` is available, and the image was generated in the following way:

```bash
docker build -t agalue/fhir-sample-generator .
docker push agalue/fhir-sample-generator
```
