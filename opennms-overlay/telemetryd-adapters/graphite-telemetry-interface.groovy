import groovy.util.logging.Slf4j
import org.opennms.netmgt.telemetry.protocols.graphite.adapter.GraphiteMetric
import org.opennms.netmgt.collection.support.builder.NodeLevelResource

@Slf4j
class CollectionSetGenerator {
    static generate(agent, builder, graphiteMsg) {
        log.debug("Generating collection set for message: {}", graphiteMsg)
        def nodeLevelResource = new NodeLevelResource(agent.getNodeId())
        builder.withGauge(nodeLevelResource, "fitness", "heartRate", graphiteMsg.floatValue());
    }
}

GraphiteMetric graphiteMsg = msg
CollectionSetGenerator.generate(agent, builder, graphiteMsg)
