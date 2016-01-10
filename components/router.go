// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/utils/log"
	"time"
)

const (
	EXPIRY_DELAY = time.Hour * 8
)

type Router struct {
	loggers []log.Logger
	brokers []core.Recipient
	db      addressKeeper // Local storage that maps end-device addresses to broker addresses
}

var ErrBadOptions = fmt.Errorf("Invalid supplied options")

// NewRouter constructs a Router and setup its internal structure
func NewRouter(brokers []core.Recipient, loggers ...log.Logger) (*Router, error) {
	localDB, err := NewLocalDB(EXPIRY_DELAY)

	if err != nil {
		return nil, err
	}

	if len(brokers) == 0 {
		return nil, ErrBadOptions
	}

	return &Router{
		loggers: loggers,
		brokers: brokers,
		db:      localDB,
	}, nil
}
