package horizon

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go test -v ./services/horizon_test/horizon.broker_test.go

func TestHorizonMessageBroker(t *testing.T) {
	env := horizon.NewEnvironmentService("../../.env")
	host := env.GetString("NATS_HOST", "localhost")
	port := env.GetInt("NATS_CLIENT_PORT", 4222)

	ctx := context.Background()

	t.Run("Connect and Disconnect", func(t *testing.T) {
		broker := horizon.NewHorizonMessageBroker(host, port)
		err := broker.Run(ctx)
		require.NoError(t, err, "should connect without error")
		err = broker.Stop(ctx)
		require.NoError(t, err, "should disconnect without error")
	})

	t.Run("Publish and Subscribe", func(t *testing.T) {
		broker := horizon.NewHorizonMessageBroker(host, port)
		err := broker.Run(ctx)
		require.NoError(t, err)
		defer broker.Stop(ctx)

		topic := fmt.Sprintf("test.topic.%d", time.Now().UnixNano())
		received := make(chan struct{})
		errChan := make(chan error, 1)

		// Subscribe to topic
		err = broker.Subscribe(ctx, topic, func(payload any) error {
			data, ok := payload.(map[string]any)
			if !ok {
				errChan <- fmt.Errorf("expected map payload, got %T", payload)
				return nil
			}
			if data["message"] != "hello" {
				errChan <- fmt.Errorf("unexpected message: %v", data["message"])
				return nil
			}
			close(received)
			return nil
		})
		require.NoError(t, err)

		// Publish message
		err = broker.Publish(ctx, topic, map[string]string{"message": "hello"})
		require.NoError(t, err)

		// Wait for result
		select {
		case <-received:
			// Success
		case err := <-errChan:
			t.Fatalf("Handler error: %v", err)
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for message")
		}
	})

	t.Run("Dispatch to Multiple Topics", func(t *testing.T) {
		broker := horizon.NewHorizonMessageBroker(host, port)
		err := broker.Run(ctx)
		require.NoError(t, err)
		defer broker.Stop(ctx)

		topic1 := fmt.Sprintf("test.topic1.%d", time.Now().UnixNano())
		topic2 := fmt.Sprintf("test.topic2.%d", time.Now().UnixNano())

		var wg sync.WaitGroup
		wg.Add(2)
		errChan := make(chan error, 2)

		// Subscribe to topic1
		err = broker.Subscribe(ctx, topic1, func(payload any) error {
			defer wg.Done()
			if _, ok := payload.(map[string]interface{}); !ok {
				errChan <- fmt.Errorf("topic1: expected map payload")
				return nil
			}
			return nil
		})
		require.NoError(t, err)

		// Subscribe to topic2
		err = broker.Subscribe(ctx, topic2, func(payload any) error {
			defer wg.Done()
			if _, ok := payload.(map[string]interface{}); !ok {
				errChan <- fmt.Errorf("topic2: expected map payload")
				return nil
			}
			return nil
		})
		require.NoError(t, err)

		// Dispatch to both topics
		err = broker.Dispatch(ctx, []string{topic1, topic2}, map[string]string{"data": "value"})
		require.NoError(t, err)

		// Wait for both messages or timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Check for handler errors
			select {
			case err := <-errChan:
				t.Fatalf("Handler error: %v", err)
			default:
				// No errors
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for messages")
		}
	})

	t.Run("Publish Without Connection", func(t *testing.T) {
		broker := horizon.NewHorizonMessageBroker(host, port)
		// Intentionally not calling Run

		err := broker.Publish(ctx, "test.topic", "payload")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "NATS connection not initialized")
	})

	t.Run("Subscribe Without Connection", func(t *testing.T) {
		broker := horizon.NewHorizonMessageBroker(host, port)
		// Intentionally not calling Run

		err := broker.Subscribe(ctx, "test.topic", func(any) error { return nil })
		require.Error(t, err)
		assert.Contains(t, err.Error(), "NATS connection not initialized")
	})
}
