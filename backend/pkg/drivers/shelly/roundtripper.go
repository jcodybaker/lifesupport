package shelly

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RequestFrame struct {
	ID     uint64 `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
	Src    string `json:"src"`
}

type ResponseFrame struct {
	ID     uint64           `json:"id"`
	Src    string           `json:"src"`
	Dst    string           `json:"dst,omitempty"`
	Error  *ErrorResponse   `json:"error,omitempty"`
	Result *json.RawMessage `json:"result,omitempty"`
}

func (r *Driver) buildSrc() string {
	return r.baseName + "/" + r.clientName
}

func (r *Driver) buildTopic() string {
	return r.buildSrc() + "/rpc"
}

func (r *Driver) handleMessage(_ mqtt.Client, m mqtt.Message) {
	var resp ResponseFrame
	if err := json.Unmarshal(m.Payload(), &resp); err != nil {
		// Log and ignore malformed messages.
		return
	}

	r.lock.Lock()
	respCh, ok := r.router[resp.ID]
	delete(r.router, resp.ID)
	r.lock.Unlock()
	if !ok {
		return
	}
	respCh <- m.Payload()
}

func (r *Driver) roundTrip(ctx context.Context, dst string, method string, params any, reply any, timeout time.Duration) error {
	id := atomic.AddUint64(&r.nextID, 1)
	ll := r.logCtx(ctx, "mqtt").With().Uint64("request_id", id).Str("method", method).Str("dst", dst).Logger()
	ll.Debug().Msg("Initiating round trip to device")
	if params == nil {
		params = json.RawMessage("{}")
	}

	// Build and publish the request message here, including the ID and parameters.
	req := RequestFrame{
		ID:     id,
		Method: method,
		Params: params,
		Src:    r.buildSrc(),
	}

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	defer func() {
		r.lock.Lock()
		delete(r.router, id)
		r.lock.Unlock()
	}()

	respCh := make(chan []byte, 1)
	r.lock.Lock()
	r.router[id] = respCh
	r.lock.Unlock()

	dstTopic := dst + "/rpc"

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	t := r.mqttClient.Publish(dstTopic, 1, false, b)
	select {
	case <-t.Done():
		if err := t.Error(); err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case resp := <-respCh:
		ll.Debug().RawJSON("resp", resp).Msg("Received response")
		var respFrame ResponseFrame
		if err := json.Unmarshal(resp, &respFrame); err != nil {
			ll.Err(err).Msg("Failed to unmarshal response frame")
			return err
		}
		if respFrame.Error != nil {
			ll.Error().Int("code", respFrame.Error.Code).Str("message", respFrame.Error.Message).Msg("Received error response from device")
			return errors.New(respFrame.Error.Message)
		}
		if respFrame.Result == nil {
			ll.Error().Msg("Received response with no result")
			return nil
		}
		return json.Unmarshal(*respFrame.Result, reply)
	case <-ctx.Done():
		return ctx.Err()
	}
}
