// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	errutil "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

var dutyManagerDB = path.Join(os.TempDir(), "TestDutyCycleStorage.db")

func TestGetSubBand(t *testing.T) {
	{
		Desc(t, "Test EuropeRX1_A")
		sb, err := GetSubBand(868.38)
		errutil.CheckErrors(t, nil, err)
		CheckSubBands(t, EuropeRX1_A, sb)
	}

	{
		Desc(t, "Test EuropeRX1_B")
		sb, err := GetSubBand(867.127)
		errutil.CheckErrors(t, nil, err)
		CheckSubBands(t, EuropeRX1_B, sb)
	}

	{
		Desc(t, "Test EuropeRX2")
		sb, err := GetSubBand(869.567)
		errutil.CheckErrors(t, nil, err)
		CheckSubBands(t, EuropeRX2, sb)
	}

	{
		Desc(t, "Test Unknown")
		sb, err := GetSubBand(433.5)
		errutil.CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckSubBands(t, 0, sb)
	}
}

func TestNewManager(t *testing.T) {
	defer func() {
		os.Remove(dutyManagerDB)
	}()
	{
		Desc(t, "Europe with valid cycleLength")
		_, err := NewDutyManager(dutyManagerDB, time.Minute, Europe)
		errutil.CheckErrors(t, nil, err)
	}

	{
		Desc(t, "Europe with invalid cycleLength")
		_, err := NewDutyManager(dutyManagerDB, 0, Europe)
		errutil.CheckErrors(t, pointer.String(string(errors.Structural)), err)
	}

	{
		Desc(t, "Not europe with valid cycleLength")
		_, err := NewDutyManager(dutyManagerDB, time.Minute, China)
		errutil.CheckErrors(t, pointer.String(string(errors.Implementation)), err)
	}
}
