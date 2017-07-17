// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package track

import (
	"net/http"
	"testing"

	"fmt"
	"net/http/httptest"
	"sync"
	"time"
)

type mockTimeoutRoundTripper struct {
}

func (r mockTimeoutRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for {
		<-time.After(timeout + 1*time.Minute)
		fmt.Print("after timeout")
		return nil, nil
	}
}

type mockRoundTripper struct {
}

func (r mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return httptest.NewRecorder().Result(), nil
}

func TestTelemetryDisabled(t *testing.T) {
	gaHTTPTransport = mockRoundTripper{}
	telemetryEnabled = false
	telemetryLogEnabled = false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	expected := send("foo", "bar", "baz", "test", wg)
	wg.Wait()

	if expected {
		t.Error("Expected to not send request")
	}
}

func TestTimeout(t *testing.T) {
	telemetryEnabled = true
	telemetryLogEnabled = false
	gaHTTPTransport = mockTimeoutRoundTripper{}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	expected := send("foo", "bar", "baz", "test", wg)
	wg.Wait()

	if expected {
		t.Error("Expected request to timeout")
	}
}

func TestSend(t *testing.T) {
	gaHTTPTransport = mockRoundTripper{}
	telemetryEnabled = true
	telemetryLogEnabled = false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	expected := send("foo", "bar", "baz", "test", wg)
	wg.Wait()

	if !expected {
		t.Error("Expected to send request")
	}
}
