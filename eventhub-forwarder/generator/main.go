package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"eventhub-forwarder/api/producer"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var totalSteps = 0

func main() {
	bootstrap := flag.String("bootstrap", "localhost:9092", "kafka bootstrap server")
	topic := flag.String("topic", "metrics", "destination kafka topic for the collection sets")
	frequency := flag.Duration("frequency", 30*time.Second, "Frequency of packet generation")
	nodeID := flag.Int64("node-id", 666, "Node ID to use for the generated content")
	nodeLabel := flag.String("node-label", "mock-device-001", "Node Label to use for the generated content")
	flag.Parse()

	kafkaProducer := createKafkaProducer(*bootstrap)
	if kafkaProducer == nil {
		return
	}

	for {
		cs := createCollectionSet(*nodeID, *nodeLabel)
		log.Printf("Sending %s", cs)
		dataBytes, err := cs.Marshal()
		if err != nil {
			log.Printf("ERROR: could not create producer: %v", err)
		}
		kafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: topic, Partition: kafka.PartitionAny},
			Value:          dataBytes,
			Key:            []byte(string(*nodeID)),
		}, nil)
		time.Sleep(*frequency)
	}
}

func createKafkaProducer(bootstrap string) *kafka.Producer {
	config := &kafka.ConfigMap{"bootstrap.servers": bootstrap}
	kafkaProducer, err := kafka.NewProducer(config)
	if err != nil {
		log.Printf("ERROR: could not create producer: %v", err)
		return nil
	}

	go func(producer *kafka.Producer) {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("message delivery failed: %v\n", ev.TopicPartition.Error)
				} else {
					log.Printf("message delivered to %v\n", ev.TopicPartition)
				}
			default:
				log.Printf("kafka producer event: %s\n", ev)
			}
		}
	}(kafkaProducer)

	return kafkaProducer
}

func createCollectionSet(nodeID int64, nodeLabel string) *producer.CollectionSet {
	totalSteps += rand.Intn(100)
	return &producer.CollectionSet{
		Timestamp: time.Now().Unix(),
		Resource: []*producer.CollectionSetResource{
			{
				Resource: &producer.CollectionSetResource_Node{
					Node: &producer.NodeLevelResource{
						NodeId:        nodeID,
						NodeLabel:     nodeLabel,
						ForeignSource: "Test",
						ForeignId:     nodeLabel,
					},
				},
				Numeric: []*producer.NumericAttribute{
					{
						Name:  "heartRate",
						Type:  producer.NumericAttribute_GAUGE,
						Value: float64(60 + rand.Intn(100)),
					},
					{
						Name:  "steps",
						Type:  producer.NumericAttribute_COUNTER,
						Value: float64(totalSteps),
					},
				},
			},
		},
	}
}
