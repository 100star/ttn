// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	core "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
)

// Adapter type materializes an mqtt adapter which implements the basic mqtt protocol
type Adapter struct {
	*MQTT.Client
	ctx           log.Interface
	packets       chan PktReq // Channel used to "transforms" incoming request to something we can handle concurrently
	registrations chan RegReq // Incoming registrations
}

// Handler defines topic-specific handler.
type Handler interface {
	Topic() string
	Handle(client *MQTT.Client, chpkt chan<- PktReq, chreg chan<- RegReq, msg MQTT.Message)
}

// Message sent through the response channel of a pktReq or regReq
type MsgRes []byte // The response content.

// Message sent through the packets channel when an incoming request arrives
type PktReq struct {
	Packet []byte      // The actual packet that has been parsed
	Chresp chan MsgRes // A response channel waiting for an success or reject confirmation
}

// Message sent through the registration channel when an incoming registration arrives
type RegReq struct {
	Registration core.Registration
	Chresp       chan MsgRes
}

// MQTT Schemes available
type Scheme string

const (
	Tcp       Scheme = "tcp"
	Tls       Scheme = "tls"
	WebSocket Scheme = "ws"
)

// NewAdapter constructs and allocates a new mqtt adapter
//
// The client is expected to be already connected to the right broker and ready to be used.
func NewAdapter(client *MQTT.Client, ctx log.Interface) *Adapter {
	adapter := &Adapter{
		Client:        client,
		ctx:           ctx,
		packets:       make(chan PktReq),
		registrations: make(chan RegReq),
	}

	return adapter
}

// NewClient generates a new paho MQTT client from an id and a broker url
//
// The broker url is expected to contain a port if needed such as mybroker.com:87354
//
// The scheme has to be the same as the one used by the broker: tcp, tls or web socket
func NewClient(id string, broker string, scheme Scheme) (*MQTT.Client, error) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s", scheme, broker))
	opts.SetClientID(id)
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, errors.New(ErrFailedOperation, token.Error())
	}
	return client, nil
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, recipients ...core.Recipient) ([]byte, error) {
	return nil, nil
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() ([]byte, core.AckNacker, error) {
	return nil, nil, nil
}

// NextRegistration implements the core.Adapter interface. Not implemented for this adapters.
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	return nil, nil, nil
}

// Bind registers a handler to a specific endpoint
func (a *Adapter) Bind(h Handler) error {
	ctx := a.ctx.WithField("topic", h.Topic())
	ctx.Info("Subscribe new handler")
	token := a.Subscribe(h.Topic(), 2, func(client *MQTT.Client, msg MQTT.Message) {
		ctx.Debug("Handle new mqtt message")
		h.Handle(client, a.packets, a.registrations, msg)
	})
	if token.Wait() && token.Error() != nil {
		ctx.WithError(token.Error()).Error("Unable to Subscribe")
		return errors.New(ErrFailedOperation, token.Error())
	}
	return nil
}
