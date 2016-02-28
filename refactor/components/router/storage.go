// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"sync"
	"time"

	. "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	dbutil "github.com/TheThingsNetwork/ttn/utils/storage"
	"github.com/brocaar/lorawan"
)

type Storage interface {
	Lookup(devEUI lorawan.EUI64) (entry, error)
	Store(reg Registration) error
}

type entry struct {
	Recipient []byte
	until     time.Time
}

type storage struct {
	sync.Mutex
	db          dbutil.Interface
	Name        string
	ExpiryDelay time.Duration
}

// newStorage creates a new internal storage for the router
func newStorage(name string, delay time.Duration) (Storage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	tableName := "brokers"
	if err := itf.Init(tableName); err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return storage{db: itf, ExpiryDelay: delay, Name: tableName}, nil
}

// Lookup implements the router.Storage interface
func (s storage) Lookup(devEUI lorawan.EUI64) (entry, error) {
	return s.lookup(devEUI, true)
}

// lookup offers an indirection in order to avoid taking a lock if not needed
func (s storage) lookup(devEUI lorawan.EUI64, lock bool) (entry, error) {
	// NOTE This works under the assumption that a read or write lock is already held by
	// the callee (e.g. Store()
	if lock {
		s.Lock()
		defer s.Unlock()
	}

	itf, err := s.db.Lookup(s.Name, devEUI[:], &entry{})
	if err != nil {
		return entry{}, errors.New(errors.Operational, err)
	}
	entries := itf.([]*entry)

	if len(entries) != 1 {
		if err := s.db.Flush(s.Name, devEUI[:]); err != nil {
			return entry{}, errors.New(errors.Operational, err)
		}
		return entry{}, errors.New(errors.Behavioural, "Not Found")
	}

	e := entries[0]

	if s.ExpiryDelay != 0 && e.until.Before(time.Now()) {
		if err := s.db.Flush(s.Name, devEUI[:]); err != nil {
			return entry{}, errors.New(errors.Operational, err)
		}
		return entry{}, errors.New(errors.Behavioural, "Not Found")
	}

	return *e, nil
}

// Store implements the router.Storage interface
func (s storage) Store(reg Registration) error {
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e entry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.Recipient)
	rw.Write(e.until)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *entry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { e.Recipient = data })
	rw.TryRead(func(data []byte) error {
		return e.until.UnmarshalBinary(data)
	})
	return rw.Err()
}
