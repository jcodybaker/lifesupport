package shelly

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"lifesupport/backend/pkg/api"
	"lifesupport/backend/pkg/drivers"
	"lifesupport/backend/pkg/storer"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jcodybaker/go-shelly"
)

func (d *Driver) DiscoverDevices(ctx context.Context, opt drivers.DiscoveryOptions, s *storer.Storer) (*drivers.DiscoveryResult, error) {
	result := &drivers.DiscoveryResult{}
	var resultMutex sync.Mutex
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
				var shellyConfig shelly.ShellyGetConfigResponse
				if err := d.roundTrip(ctx, deviceInfo.ID, "Shelly.GetConfig", nil, &shellyConfig, time.Second*5); err != nil {
					ll.Err(err).
						Str("device_id", deviceInfo.ID).
						Msg("querying shelly for full device config")
					return
				}
				dev := d.deviceInfoToDevice(deviceInfo, &shellyConfig)
				if err := s.CreateDevice(ctx, dev); err != nil {
					if errors.Is(err, storer.ErrAlreadyExists) {
						ll.Debug().
							Err(err).
							Str("device_id", deviceInfo.ID).
							Msg("device already exists in store")
						return
					}
					ll.Err(err).
						Str("device_id", deviceInfo.ID).
						Msg("storing discovered device")
					return
				}
				resultMutex.Lock()
				result.DiscoveredTags = append(result.DiscoveredTags, dev.DefaultTag())
				resultMutex.Unlock()
				ll.Info().
					Str("device_id", deviceInfo.ID).
					Msg("discovered new device")
			}(deviceInfo)
		}
	}()

	// Ok, we're ready for responses; make our request.
	token := d.mqttClient.Publish("shellies/command", 1, false, []byte("announce"))
	token.Wait()
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("publishing search message to mqtt: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(d.discoveryTimeout):
	}

	// We can't guarantee that the mqtt has coallesed and processed all incoming messages. So it's difficult
	// be certain we can close the channel. The atomic stopSearch makes this safer, but it's not a guarantee.
	stopSearch.Store(true)
	token = d.mqttClient.Unsubscribe("shellies/announce")
	token.Wait()
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("unsubscribing from mqtt search message responses: %w", err)
	}
	close(searchBuf)

	wg.Wait()
	close(workerLimiter)
	return result, nil
}

func (d *Driver) deviceInfoToDevice(info *shelly.ShellyGetDeviceInfoResponse, config *shelly.ShellyGetConfigResponse) *api.Device {
	dev := &api.Device{
		ID:          info.ID,
		Driver:      api.DriverShelly,
		Name:        info.ID,
		Description: fmt.Sprintf("Shelly %s %s", info.App, info.MAC),
	}
	for _, s := range config.Switches {
		r := &api.Relay{
			BaseActuator: api.BaseActuator{
				ID:           fmt.Sprintf("switch:%d", s.ID),
				ActuatorType: api.ActuatorTypeRelay,
				DeviceID:     dev.ID,
				Name:         *s.Name,
			},
		}
		if r.Name == "" {
			r.Name = fmt.Sprintf("%s Switch %d", dev.Name, s.ID)
		}
		r.Tags = []string{r.DefaultTag(dev.ID)}
		dev.Actuators = append(dev.Actuators, r)
	}
	return dev
}
