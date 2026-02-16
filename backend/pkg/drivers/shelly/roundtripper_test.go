package shelly

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MockToken implements mqtt.Token interface
type MockToken struct {
	done  chan struct{}
	err   error
	mutex sync.Mutex
}

func NewMockToken(err error) *MockToken {
	return &MockToken{
		done: make(chan struct{}),
		err:  err,
	}
}

func (m *MockToken) Wait() bool {
	<-m.done
	return true
}

func (m *MockToken) WaitTimeout(duration time.Duration) bool {
	select {
	case <-m.done:
		return true
	case <-time.After(duration):
		return false
	}
}

func (m *MockToken) Done() <-chan struct{} {
	return m.done
}

func (m *MockToken) Error() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.err
}

func (m *MockToken) Complete() {
	close(m.done)
}

// MockMQTTClient implements mqtt.Client interface for testing
type MockMQTTClient struct {
	publishFunc   func(topic string, qos byte, retained bool, payload interface{}) mqtt.Token
	subscribeFunc func(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token
}

func (m *MockMQTTClient) IsConnected() bool       { return true }
func (m *MockMQTTClient) IsConnectionOpen() bool  { return true }
func (m *MockMQTTClient) Connect() mqtt.Token     { return nil }
func (m *MockMQTTClient) Disconnect(quiesce uint) {}
func (m *MockMQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	if m.publishFunc != nil {
		return m.publishFunc(topic, qos, retained, payload)
	}
	token := NewMockToken(nil)
	token.Complete()
	return token
}
func (m *MockMQTTClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	if m.subscribeFunc != nil {
		return m.subscribeFunc(topic, qos, callback)
	}
	token := NewMockToken(nil)
	token.Complete()
	return token
}
func (m *MockMQTTClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	return nil
}
func (m *MockMQTTClient) Unsubscribe(topics ...string) mqtt.Token {
	token := NewMockToken(nil)
	token.Complete()
	return token
}
func (m *MockMQTTClient) AddRoute(topic string, callback mqtt.MessageHandler) {}
func (m *MockMQTTClient) OptionsReader() mqtt.ClientOptionsReader             { return mqtt.ClientOptionsReader{} }

// MockMessage implements mqtt.Message interface
type MockMessage struct {
	payload []byte
	topic   string
}

func (m *MockMessage) Duplicate() bool   { return false }
func (m *MockMessage) Qos() byte         { return 0 }
func (m *MockMessage) Retained() bool    { return false }
func (m *MockMessage) Topic() string     { return m.topic }
func (m *MockMessage) MessageID() uint16 { return 0 }
func (m *MockMessage) Payload() []byte   { return m.payload }
func (m *MockMessage) Ack()              {}

func TestRoundTrip_Success(t *testing.T) {
	// Create a mock MQTT client
	var publishedPayload []byte
	var publishedTopic string

	// Create expected response
	result := json.RawMessage(`{"status":"ok"}`)
	response := ResponseFrame{
		ID:     1,
		Src:    "shelly/test-device",
		Result: &result,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Create a driver with the mock client
	driver := &Driver{
		mqttClient: nil, // Will be set after mockClient is created
		nextID:     0,
		clientName: "test-client",
		baseName:   "lifesupport",
		router:     make(map[uint64]chan []byte),
	}

	mockClient := &MockMQTTClient{}
	mockClient.publishFunc = func(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
		publishedTopic = topic
		publishedPayload = payload.([]byte)

		// Simulate receiving a response immediately after publish
		mockMessage := &MockMessage{
			payload: responseBytes,
			topic:   "lifesupport/rpc",
		}
		go driver.handleMessage(mockClient, mockMessage)

		token := NewMockToken(nil)
		token.Complete()
		return token
	}

	driver.mqttClient = mockClient

	// Create test parameters
	params := map[string]interface{}{
		"key": "value",
	}

	// Start roundTrip in a goroutine
	var reply ResponseFrame
	errCh := make(chan error, 1)
	go func() {
		errCh <- driver.roundTrip(
			context.Background(),
			"test-device",
			"Shelly.GetStatus",
			params,
			&reply,
			5*time.Second,
		)
	}()

	// Check for errors
	err = <-errCh
	if err != nil {
		t.Fatalf("roundTrip failed: %v", err)
	}

	// Verify the published topic
	expectedTopic := "shelly/test-device/rpc"
	if publishedTopic != expectedTopic {
		t.Errorf("Expected topic %s, got %s", expectedTopic, publishedTopic)
	}

	// Verify the published payload
	var req RequestFrame
	if err := json.Unmarshal(publishedPayload, &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}
	if req.Method != "Shelly.GetStatus" {
		t.Errorf("Expected method Shelly.GetStatus, got %s", req.Method)
	}
	if req.ID != 1 {
		t.Errorf("Expected ID 1, got %d", req.ID)
	}

	// Verify the reply
	if reply.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	var replyData map[string]interface{}
	if err := json.Unmarshal(*reply.Result, &replyData); err != nil {
		t.Fatalf("Failed to unmarshal reply result: %v", err)
	}

	if replyData["status"] != "ok" {
		t.Errorf("Expected status ok, got %v", replyData["status"])
	}
}

func TestRoundTrip_PublishError(t *testing.T) {
	publishErr := errors.New("publish failed")

	mockClient := &MockMQTTClient{
		publishFunc: func(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
			token := NewMockToken(publishErr)
			token.Complete()
			return token
		},
	}

	driver := &Driver{
		mqttClient: mockClient,
		nextID:     0,
		clientName: "test-client",
		baseName:   "lifesupport",
		router:     make(map[uint64]chan []byte),
	}

	var reply map[string]interface{}
	err := driver.roundTrip(
		context.Background(),
		"test-device",
		"Shelly.GetStatus",
		nil,
		&reply,
		5*time.Second,
	)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !errors.Is(err, publishErr) && err.Error() != publishErr.Error() {
		t.Errorf("Expected error %v, got %v", publishErr, err)
	}
}

func TestRoundTrip_ContextTimeout(t *testing.T) {
	mockClient := &MockMQTTClient{
		publishFunc: func(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
			token := NewMockToken(nil)
			token.Complete()
			return token
		},
	}

	driver := &Driver{
		mqttClient: mockClient,
		nextID:     0,
		clientName: "test-client",
		baseName:   "lifesupport",
		router:     make(map[uint64]chan []byte),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var reply map[string]interface{}
	err := driver.roundTrip(
		ctx,
		"test-device",
		"Shelly.GetStatus",
		nil,
		&reply,
		0, // No additional timeout
	)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}

	// Verify router was cleaned up
	driver.lock.Lock()
	routerSize := len(driver.router)
	driver.lock.Unlock()
	if routerSize != 0 {
		t.Errorf("Expected router to be empty, got %d entries", routerSize)
	}
}

func TestRoundTrip_ResponseTimeout(t *testing.T) {
	mockClient := &MockMQTTClient{
		publishFunc: func(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
			token := NewMockToken(nil)
			token.Complete()
			return token
		},
	}

	driver := &Driver{
		mqttClient: mockClient,
		nextID:     0,
		clientName: "test-client",
		baseName:   "lifesupport",
		router:     make(map[uint64]chan []byte),
	}

	var reply map[string]interface{}
	err := driver.roundTrip(
		context.Background(),
		"test-device",
		"Shelly.GetStatus",
		nil,
		&reply,
		100*time.Millisecond, // Short timeout
	)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}

	// Verify router was cleaned up
	driver.lock.Lock()
	routerSize := len(driver.router)
	driver.lock.Unlock()
	if routerSize != 0 {
		t.Errorf("Expected router to be empty, got %d entries", routerSize)
	}
}

func TestRoundTrip_ErrorResponse(t *testing.T) {
	// Create error response
	response := ResponseFrame{
		ID:  1,
		Src: "shelly/test-device",
		Error: &ErrorResponse{
			Code:    -1,
			Message: "Device error",
		},
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	driver := &Driver{
		mqttClient: nil, // Will be set after mockClient is created
		nextID:     0,
		clientName: "test-client",
		baseName:   "lifesupport",
		router:     make(map[uint64]chan []byte),
	}

	mockClient := &MockMQTTClient{}
	mockClient.publishFunc = func(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
		// Simulate receiving an error response immediately after publish
		mockMessage := &MockMessage{
			payload: responseBytes,
			topic:   "lifesupport/rpc",
		}
		go driver.handleMessage(mockClient, mockMessage)

		token := NewMockToken(nil)
		token.Complete()
		return token
	}

	driver.mqttClient = mockClient

	// Start roundTrip in a goroutine
	var reply map[string]interface{}
	errCh := make(chan error, 1)
	go func() {
		errCh <- driver.roundTrip(
			context.Background(),
			"test-device",
			"Shelly.GetStatus",
			nil,
			&reply,
			5*time.Second,
		)
	}()

	// The roundTrip should complete successfully (no error from roundTrip itself)
	// The error is in the response payload
	err = <-errCh
	if err != nil {
		t.Fatalf("roundTrip failed: %v", err)
	}
}

func TestRoundTrip_ConcurrentRequests(t *testing.T) {
	var mu sync.Mutex
	publishCount := 0
	responses := make(map[uint64][]byte)

	driver := &Driver{
		mqttClient: nil, // Will be set after mockClient is created
		nextID:     0,
		clientName: "test-client",
		baseName:   "lifesupport",
		router:     make(map[uint64]chan []byte),
	}

	mockClient := &MockMQTTClient{}
	mockClient.publishFunc = func(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
		mu.Lock()
		publishCount++

		// Parse the request to get the ID
		var req RequestFrame
		json.Unmarshal(payload.([]byte), &req)

		// Get the pre-created response for this ID
		responseBytes := responses[req.ID]
		mu.Unlock()

		// Simulate receiving a response immediately after publish
		if responseBytes != nil {
			mockMessage := &MockMessage{
				payload: responseBytes,
				topic:   "lifesupport/rpc",
			}
			go driver.handleMessage(mockClient, mockMessage)
		}

		token := NewMockToken(nil)
		token.Complete()
		return token
	}

	driver.mqttClient = mockClient

	// Launch multiple concurrent requests
	numRequests := 10

	// Pre-create all responses before starting requests
	for i := 0; i < numRequests; i++ {
		result := json.RawMessage(`{"index":` + string(rune(i+'0')) + `}`)
		response := ResponseFrame{
			ID:     uint64(i + 1),
			Src:    "shelly/test-device",
			Result: &result,
		}
		responseBytes, _ := json.Marshal(response)
		responses[uint64(i+1)] = responseBytes
	}

	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			defer wg.Done()

			// Start roundTrip
			var reply map[string]interface{}
			errCh := make(chan error, 1)
			go func() {
				errCh <- driver.roundTrip(
					context.Background(),
					"test-device",
					"Shelly.GetStatus",
					nil,
					&reply,
					5*time.Second,
				)
			}()

			// Check result
			if err := <-errCh; err != nil {
				t.Errorf("Request %d failed: %v", index, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all requests were published
	mu.Lock()
	count := publishCount
	mu.Unlock()
	if count != numRequests {
		t.Errorf("Expected %d publishes, got %d", numRequests, count)
	}

	// Verify router is empty
	driver.lock.Lock()
	routerSize := len(driver.router)
	driver.lock.Unlock()
	if routerSize != 0 {
		t.Errorf("Expected router to be empty, got %d entries", routerSize)
	}
}
