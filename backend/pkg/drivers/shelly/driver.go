package shelly

import (
	"context"
	"errors"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	defaultBaseName            = "lifesupport"
	defaultDiscoveryBufferSize = 10
	defaultDiscoveryTimeout    = 10 * time.Second
	defaultDiscoveryWorkers    = 5
)

func New(mqttClient mqtt.Client, opts ...Option) *Driver {
	hostname, _ := os.Hostname()
	nextID := rand.Uint64()
	rt := &Driver{
		mqttClient:          mqttClient,
		nextID:              nextID,
		clientName:          hostname,
		baseName:            defaultBaseName,
		discoveryBufferSize: defaultDiscoveryBufferSize,
		discoveryTimeout:    defaultDiscoveryTimeout,
		discoveryWorkers:    defaultDiscoveryWorkers,
		router:              make(map[uint64]chan []byte),
	}
	for _, opt := range opts {
		opt(rt)
	}
	return rt
}

type Driver struct {
	mqttClient mqtt.Client

	// discovery
	discoveryBufferSize int
	discoveryTimeout    time.Duration
	discoveryWorkers    int

	// rtt
	nextID     uint64
	clientName string
	baseName   string
	router     map[uint64]chan []byte
	lock       sync.Mutex
}

func (r *Driver) Start(ctx context.Context) error {
	if r.clientName == "" {
		return errors.New("client name cannot be empty")
	}
	t := r.mqttClient.Subscribe(r.buildTopic(), 1, r.handleMessage)
	select {
	case <-t.Done():
		return t.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Driver) Stop(ctx context.Context) error {
	t := r.mqttClient.Unsubscribe(r.buildSrc())
	select {
	case <-t.Done():
		return t.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *Driver) logCtx(ctx context.Context, sub string) zerolog.Logger {
	ll := log.Ctx(ctx).With().Str("component", "shelly")
	if sub != "" {
		ll = ll.Str("subcomponent", sub)
	}
	return ll.Logger()
}

// func (d *Driver) MQTTConnect(ctx context.Context) error {
// 	ll := d.logCtx(ctx, "mqtt")
// 	if d.mqttClientOptions == nil {
// 		ll.Debug().Msg("no MQTT servers defined; skipping mqtt connect")
// 		return nil
// 	}
// 	// opts.SetConnectionLostHandler(c.onConnectionLost)
// 	ll.Info().Str("broker", d.mqttClientOptions.Servers[0].String()).Msg("connecting to MQTT Broker")
// 	d.mqttClient = mqtt.NewClient(d.mqttClientOptions)

// 	token := d.mqttClient.Connect()
// 	token.Wait()
// 	if err := token.Error(); err != nil {
// 		return fmt.Errorf("MQTT connect error: %w", err)
// 	}

// 	for _, t := range d.mqttTopicSubs {
// 		c, err := newMQTTConsumer(ctx, t, d.mqttClient)
// 		if err != nil {
// 			return fmt.Errorf("subscribing to MQTT topic %q: %w", t, err)
// 		}
// 		s := mgrpc.Serve(ctx, c)
// 		d.notifications.register(s)
// 	}
// 	return nil
// }
