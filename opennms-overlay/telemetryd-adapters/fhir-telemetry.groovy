import groovy.util.logging.Slf4j

import org.opennms.netmgt.collection.api.AttributeType
import org.opennms.netmgt.collection.support.builder.NodeLevelResource
import org.opennms.netmgt.telemetry.protocols.nxos.adapter.proto.TelemetryBis
import org.opennms.netmgt.telemetry.protocols.nxos.adapter.NxosGpbParserUtil

@Slf4j
class CollectionSetGenerator {
  static generate(agent, builder, telemetryMsg) {
    log.debug("Generating collection set for node {} from message: {}", agent.getNodeId(), telemetryMsg)
    def nodeLevelResource = new NodeLevelResource(agent.getNodeId())
    builder.withNumericAttribute(nodeLevelResource, "fhir-stats", "heartRate",
      NxosGpbParserUtil.getValueAsDouble(telemetryMsg, "heartRate"), AttributeType.GAUGE)
    builder.withNumericAttribute(nodeLevelResource, "fhir-stats", "stepCount",
      NxosGpbParserUtil.getValueAsDouble(telemetryMsg, "stepCount"), AttributeType.COUNTER)
  }
}

TelemetryBis.Telemetry telemetryMsg = msg
CollectionSetGenerator.generate(agent, builder, telemetryMsg)
