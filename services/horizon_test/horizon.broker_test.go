package horizon

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lands-horizon/horizon-server/services/horizon"
)

// go test ./services/horizon_test/horizon.broker_test.go

func TestHorizonMessageBroker_PublishSubscribe(t *testing.T) {
	env := horizon.NewEnvironmentService("../../.env")
	host := env.GetString("NATS_HOST", "localhost")
	port := env.GetInt("NATS_CLIENT_PORT", 4222)

	ctx := context.Background()
	broker := horizon.NewHorizonMessageBroker(host, port)

	err := broker.Run(ctx)
	if err != nil {
		t.Fatalf("failed to run broker: %v", err)
	}
	defer broker.Stop(ctx)

	topic := "test.topic"
	expectedMsg := map[string]interface{}{
		"message": "hello",
	}

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedMsg map[string]interface{}

	err = broker.Subscribe(ctx, topic, func(msg any) error {
		defer wg.Done()
		if data, ok := msg.(map[string]interface{}); ok {
			receivedMsg = data
		} else {
			t.Errorf("received message in unexpected format: %T", msg)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Allow the subscription to be ready
	time.Sleep(500 * time.Millisecond)

	err = broker.Publish(ctx, topic, expectedMsg)
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message to be received")
	}

	if receivedMsg["message"] != expectedMsg["message"] {
		t.Errorf("expected %v, got %v", expectedMsg["message"], receivedMsg["message"])
	}
}
