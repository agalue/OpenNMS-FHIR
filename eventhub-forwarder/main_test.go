package main

import (
	"testing"
	"time"

	"eventhub-forwarder/api/producer"
)

var nodeLevelResource = &producer.CollectionSetResource_Node{
	Node: &producer.NodeLevelResource{
		NodeId:        666,
		NodeLabel:     "mock-device",
		ForeignSource: "Test",
		ForeignId:     "mock-device",
	},
}

func TestForwardMethod(t *testing.T) {
	var err error
	cli := &eventHubClient{
		variables: "heartRate,stepCount",
		debug:     true,
	}

	// Case 1: Empty CollectionSet
	cs1 := &producer.CollectionSet{
		Timestamp: time.Now().Unix(),
		Resource:  []*producer.CollectionSetResource{},
	}
	// An error is expected
	if err = cli.forward(cs1); err == nil {
		t.FailNow()
	}

	// Case 2: CollectionSet without node-level resource
	cs2 := &producer.CollectionSet{
		Timestamp: time.Now().Unix(),
		Resource: []*producer.CollectionSetResource{
			{
				Resource: &producer.CollectionSetResource_Response{
					Response: &producer.ResponseTimeResource{
						Instance: "127.0.0.1",
						Location: "Test",
					},
				},
				Numeric: []*producer.NumericAttribute{
					{
						Name:  "responseTime",
						Type:  producer.NumericAttribute_GAUGE,
						Value: float64(100),
					},
				},
			},
		},
	}
	// An error is expected
	if err = cli.forward(cs2); err == nil {
		t.FailNow()
	}

	// Case 3: CollectionSet with node-level resource but no FHIR metrics
	cs3 := &producer.CollectionSet{
		Timestamp: time.Now().Unix(),
		Resource: []*producer.CollectionSetResource{
			{
				Resource: nodeLevelResource,
				Numeric: []*producer.NumericAttribute{
					{
						Name:  "someVariable",
						Type:  producer.NumericAttribute_GAUGE,
						Value: float64(10),
					},
				},
			},
		},
	}
	// An error is expected
	if err = cli.forward(cs3); err == nil {
		t.FailNow()
	}

	// Case 4: Collection set with node-level resource and valid FHIR metrics
	cs4 := &producer.CollectionSet{
		Timestamp: time.Now().Unix(),
		Resource: []*producer.CollectionSetResource{
			{
				Resource: nodeLevelResource,
				Numeric: []*producer.NumericAttribute{
					{
						Name:  "heartRate",
						Type:  producer.NumericAttribute_GAUGE,
						Value: float64(72),
					},
					{
						Name:  "stepCount",
						Type:  producer.NumericAttribute_COUNTER,
						Value: float64(100),
					},
				},
			},
		},
	}
	// No errors should be received
	if err = cli.forward(cs4); err != nil {
		t.FailNow()
	}
}
