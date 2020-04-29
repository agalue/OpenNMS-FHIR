package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"golang.org/x/net/context"
)

func main() {
	connectionStr := flag.String("connection-str", "", "Azure Event Hub Connection String")
	flag.Parse()

	if *connectionStr == "" {
		log.Fatal("Azure Event Hub Connection String cannot be empty")
	}

	hub, err := eventhub.NewHubFromConnectionString(*connectionStr)
	if err != nil {
		log.Fatal(err)
	}

	handler := func(c context.Context, event *eventhub.Event) error {
		log.Println(string(event.Data))
		return nil
	}

	ctx := context.Background()
	runtimeInfo, err := hub.GetRuntimeInformation(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, partitionID := range runtimeInfo.PartitionIDs {
		_, err := hub.Receive(ctx, partitionID, handler, eventhub.ReceiveWithLatestOffset())
		if err != nil {
			log.Println(err)
		}
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	err = hub.Close(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
