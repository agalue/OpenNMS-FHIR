package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"eventhub-forwarder/api/producer"

	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"github.com/golang/protobuf/proto"

	"golang.org/x/net/context"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

// https://github.com/microsoft/iomt-fhir/blob/master/docs/Configuration.md
type eventHubMessage struct {
	Body             map[string]string `json:"body"`
	Properties       map[string]string `json:"properties"`
	SystemProperties map[string]string `json:"systemProperties"`
}

type eventHubClient struct {
	connectionStr string
	variables     string
	hub           *eventhub.Hub
	debug         bool
}

func (cli *eventHubClient) init() error {
	if cli.debug {
		fmt.Printf("DEBUG mode enabled. Event Hub forwarding will be ignored.")
		return nil
	}
	if cli.connectionStr == "" {
		return fmt.Errorf("Azure Event Hub Connection String cannot be empty")
	}
	var err error
	if cli.hub, err = eventhub.NewHubFromConnectionString(cli.connectionStr); err != nil {
		return err
	}
	return nil
}

func (cli *eventHubClient) stop() {
	if cli.hub == nil {
		return
	}
	if err := cli.hub.Close(context.Background()); err != nil {
		log.Println(err)
	}
}

func (cli *eventHubClient) getMetricValue(r *producer.CollectionSetResource, metricName string) (float64, bool) {
	for _, metric := range r.Numeric {
		if metric.Name == metricName {
			return metric.Value, true
		}
	}
	return 0, false
}

func (cli *eventHubClient) forward(cset *producer.CollectionSet) error {
	if len(cset.Resource) == 0 {
		return fmt.Errorf("There are no resources on the collection-set, ignoring")
	}
	var resource *producer.CollectionSetResource = nil
	for _, r := range cset.Resource {
		if r.GetNode() != nil {
			resource = r
			break
		}
	}
	if resource == nil {
		return fmt.Errorf("Cannot find node-level resource on the collection-set, ignoring")
	}
	message := &eventHubMessage{
		Body:             make(map[string]string),
		Properties:       make(map[string]string),
		SystemProperties: make(map[string]string),
	}
	message.Body["deviceId"] = resource.GetNode().NodeLabel
	message.Body["endDate"] = time.Unix(cset.Timestamp, 0).Format(time.RFC3339)
	metricsAdded := 0
	for _, metric := range strings.Split(cli.variables, ",") {
		if value, ok := cli.getMetricValue(resource, metric); ok {
			message.Body[metric] = fmt.Sprintf("%.0f", value)
			metricsAdded++
		}
	}
	if metricsAdded > 0 {
		dataBytes, _ := json.Marshal(message)
		log.Printf("Sending message to Event Hub: %s", string(dataBytes))
		if cli.debug {
			return nil
		}
		return cli.hub.Send(context.Background(), eventhub.NewEvent(dataBytes))
	}
	return fmt.Errorf("Couldn't find FHIR variables on the collection set for node %s, ignoring", resource.GetNode().NodeLabel)
}

type kafkaClient struct {
	bootstrap        string
	sourceTopic      string
	groupID          string
	consumerSettings string
	consumer         *kafka.Consumer
	eventHubClient   *eventHubClient
}

func (cli *kafkaClient) getKafkaConfig() *kafka.ConfigMap {
	config := &kafka.ConfigMap{
		"bootstrap.servers": cli.bootstrap,
		"group.id":          cli.groupID,
	}
	if cli.consumerSettings != "" {
		for _, kv := range strings.Split(cli.consumerSettings, ", ") {
			array := strings.Split(kv, "=")
			if len(array) == 2 {
				if err := config.SetKey(array[0], array[1]); err != nil {
					fmt.Printf("Invalid property %s=%s: %v", array[0], array[1], err)
				}
			} else {
				fmt.Printf("Invalid key-value pair %s", kv)
			}
		}
	}
	return config
}

func (cli *kafkaClient) start() error {
	if cli.eventHubClient == nil {
		return fmt.Errorf("Azure Event Hub Client cannot be null")
	}
	if err := cli.eventHubClient.init(); err != nil {
		return err
	}

	jsonBytes, _ := json.MarshalIndent(cli, "", "  ")
	log.Println(string(jsonBytes))

	if cli.sourceTopic == "" {
		return fmt.Errorf("Source topic cannot be empty")
	}

	var err error
	cli.consumer, err = kafka.NewConsumer(cli.getKafkaConfig())
	if err != nil {
		return fmt.Errorf("Could not create consumer: %v", err)
	}
	cli.consumer.SubscribeTopics([]string{cli.sourceTopic}, nil)

	go func() {
		for {
			msg, err := cli.consumer.ReadMessage(-1)
			if err == nil {
				var data = &producer.CollectionSet{}
				if err := proto.Unmarshal(msg.Value, data); err != nil {
					log.Printf("ERROR: Invalid message received: %v\n", err)
				} else {
					log.Printf("Received %s", data.String())
					if err := cli.eventHubClient.forward(data); err != nil {
						log.Printf("ERROR: Cannot send message: %v\n", err)
					}
				}
			} else {
				log.Printf("ERROR: Kafka consumer problem: %v\n", err)
			}
		}
	}()

	log.Printf("Kafka consumer started against %s\n", cli.bootstrap)
	return nil
}

func (cli *kafkaClient) stop() {
	cli.eventHubClient.stop()
	cli.consumer.Close()
	log.Println("good bye!")
}

// bootstrap function

func main() {
	ehcli := &eventHubClient{}
	kcli := &kafkaClient{eventHubClient: ehcli}

	flag.StringVar(&kcli.bootstrap, "bootstrap", "localhost:9092", "kafka bootstrap server")
	flag.StringVar(&kcli.sourceTopic, "source-topic", "metrics", "kafka source topic with OpenNMS Producer GPB messages")
	flag.StringVar(&kcli.groupID, "group-id", "kafka-converter", "kafka consumer group ID")
	flag.StringVar(&kcli.consumerSettings, "consumer-params", "", "optional kafka consumer parameters as a CSV of Key-Value pairs")
	flag.BoolVar(&ehcli.debug, "debug", false, "Enable DEBUG mode (print data to stdout and ignore Event Hub forwarding)")
	flag.StringVar(&ehcli.connectionStr, "connection-str", "", "Azure Event Hub Connection String")
	flag.StringVar(&ehcli.variables, "metrics-list", "heartRate,stepCount", "List of metrics to forward to Event Hub as a CSV list")
	flag.Parse()

	if err := kcli.start(); err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	kcli.stop()
}
