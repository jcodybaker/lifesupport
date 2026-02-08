package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jcodybaker/go-shelly"
	"go.temporal.io/sdk/workflow"
)

func (d *Driver) DeviceDiscoveryWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, d.searchMQTT).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (d *Driver) searchMQTT(ctx context.Context) error {
	stopSearch := new(atomic.Bool)
	ll := d.logCtx(ctx, "discovery")

	searchBuf := make(chan *shelly.ShellyGetDeviceInfoResponse, d.discoveryBufferSize)
	var wg sync.WaitGroup

	d.mqttClient.Subscribe("shellies/announce", 1, func(_ mqtt.Client, m mqtt.Message) {
		var deviceInfo shelly.ShellyGetDeviceInfoResponse
		if err := json.Unmarshal(m.Payload(), &deviceInfo); err != nil {
			ll.Err(err).
				Uint16("message_id", m.MessageID()).
				Str("topic", m.Topic()).
				Msg("parsing MQTT message as device info")
			return
		}
		if stopSearch.Load() {
			ll.Warn().
				Uint16("message_id", m.MessageID()).
				Str("topic", m.Topic()).
				Str("device_id", deviceInfo.ID).
				Msg("discarding late MQTT search response")
			return
		}
		ll.Debug().
			Uint16("message_id", m.MessageID()).
			Str("topic", m.Topic()).
			Str("device_id", deviceInfo.ID).
			Msg("got MQTT search response")
		searchBuf <- &deviceInfo
	})

	wg.Add(1)
	workerLimiter := make(chan struct{}, d.discoveryWorkers)
	go func() {
		defer wg.Done()
		for deviceInfo := range searchBuf {
			wg.Add(1)
			// Occupy a space in the workerLimiter buffer or block until one is available.
			workerLimiter <- struct{}{}
			go func(deviceInfo *shelly.ShellyGetDeviceInfoResponse) {
				defer wg.Done()
				defer func() { <-workerLimiter }()
				// dev, err := d.shellyQueryFullDevice(ctx, deviceInfo)
			}(deviceInfo)
		}
	}()

	// Ok, we're ready for responses; make our request.
	token := d.mqttClient.Publish("shellies/command", 1, false, []byte("announce"))
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("publishing search message to mqtt: %w", err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(d.discoveryTimeout):
	}

	// We can't guarantee that the mqtt has coallesed and processed all incoming messages. So it's difficult
	// be certain we can close the channel. The atomic stopSearch makes this safer, but it's not a guarantee.
	stopSearch.Store(true)
	token = d.mqttClient.Unsubscribe("shellies/announce")
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("unsubscribing from mqtt search message responses: %w", err)
	}
	close(searchBuf)

	wg.Wait()
	close(workerLimiter)
	return nil
}
