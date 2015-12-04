// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package gateway offers a dummy representation of a gateway.
//
// The package can be used to create a dummy gateway.
// Its former use is to provide a handy simulator for further testing of the whole network chain.
package gateway

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"math/rand"
)

type Gateway struct {
	Coord   GPSCoord // Gateway's GPS coordinates
	Id      string   // Gateway's Identifier
	Routers []string // List of routers addresses

	ackr float64 // Percentage of upstream datagrams that were acknowledged
	dwnb uint    // Number of downlink datagrams received
	rxfw uint    // Number of radio packets forwarded
	rxnb uint    // Number of radio packets received
	rxok uint    // Number of radio packets received with a valid  PHY CRC
	txnb uint    // Number of packets emitted

	stderr <-chan error          // Output error channel
	stdout <-chan semtech.Packet // Output communication channel
}

type GPSCoord struct {
	altitude  int     // GPS altitude in RX meters
	latitude  float64 // GPS latitude, North is +
	longitude float64 // GPS longitude, East is +
}

func New(id string, routers ...string) (*Gateway, error) {
	fmt.Printf("")
	if id == "" {
		return nil, errors.New("Invalid gateway id provided")
	}

	if len(routers) == 0 {
		return nil, errors.New("At least one router address should be provided")
	}

	wrongAddress := false
	for _, r := range routers {
		wrongAddress = wrongAddress || (r == "")
	}

	if wrongAddress {
		return nil, errors.New("Invalid router address")
	}

	return &Gateway{
		Id: id,
		Coord: GPSCoord{
			altitude:  120, // TEMPORARY
			latitude:  53.3702,
			longitude: 4.8952,
		},
		Routers: routers,
	}, nil
}

func genToken() []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, rand.Uint32())
	return b[0:2]
}

type Forwarder interface {
	Forward(packet semtech.Packet)
	Start() (<-chan semtech.Packet, <-chan error)
	Stat() semtech.Stat
}

type Imitator interface {
	Mimic()
}
