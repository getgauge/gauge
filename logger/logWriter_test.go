// Copyright 2019 ThoughtWorks, Inc.

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

package logger

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/getgauge/gauge/config"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.RemoveAll(filepath.Join("_testdata", "logs"))
	os.Exit(exitCode)
}

func TestLogWriterToOutputInfoLogInCorrectFormat(t *testing.T) {
	defer tearDown(t)
	setupLogger("info")
	l := newLogWriter("js")

	if _, err := l.Stdout.Write([]byte("{\"logLevel\": \"info\", \"message\": \"Foo\"}\n")); err != nil {
		t.Fatalf("Unable to write to logWriter")
	}

	assertLogContains(t, []string{"[js] [INFO] Foo"})
}

func TestLogWriterToOutputInfoLogWithMultipleLines(t *testing.T) {
	defer tearDown(t)
	setupLogger("debug")
	l := newLogWriter("js")

	if _, err := l.Stdout.Write([]byte("{\"logLevel\": \"info\", \"message\": \"Foo\"}\n{\"logLevel\": \"debug\", \"message\": \"Bar\"}\n")); err != nil {
		t.Fatalf("Unable to write to logWriter")
	}

	assertLogContains(t, []string{"[js] [INFO] Foo", "[js] [DEBUG] Bar"})
}

func TestLogWriterToLogPlainStrings(t *testing.T) {
	defer tearDown(t)
	setupLogger("debug")
	l := newLogWriter("js")

	if _, err := l.Stdout.Write([]byte("Foo Bar\n")); err != nil {
		t.Fatalf("Unable to write to logWriter")
	}

	assertLogContains(t, []string{"Foo Bar"})
}

func TestLoggingFromDifferentWritersAtSameTime(t *testing.T) {
	defer tearDown(t)
	setupLogger("info")
	j := newLogWriter("js")
	h := newLogWriter("html-report")

	var wg sync.WaitGroup
	var err error
	wg.Add(5)
	go func() {
		Debug(false, "debug msg")
		wg.Done()
	}()
	go func() {
		_, err = h.Stdout.Write([]byte("{\"logLevel\": \"warning\", \"message\": \"warning msg\"}\n{\"logLevel\": \"debug\", \"message\": \"debug msg\"}\n"))
		wg.Done()
	}()
	go func() {
		_, err = j.Stdout.Write([]byte("{\"logLevel\": \"info\", \"message\": \"info msg\"}\n{\"logLevel\": \"error\", \"message\": \"error msg\"}\n"))
		wg.Done()
	}()
	go func() {
		_, err = h.Stdout.Write([]byte("{\"logLevel\": \"info\", \"message\": \"info msg\"}\n{\"logLevel\": \"error\", \"message\": \"error msg\"}\n"))
		wg.Done()
	}()
	go func() {
		_, err = j.Stdout.Write([]byte("{\"logLevel\": \"warning\", \"message\": \"warning msg\"}\n{\"logLevel\": \"debug\", \"message\": \"debug msg\"}\n"))
		wg.Done()
	}()
	wg.Wait()
	if err != nil {
		t.Fatalf("Unable to write to logWriter")
	}

	assertLogContains(t, []string{
		"[js] [WARNING] warning msg",
		"[js] [ERROR] error msg",
		"[js] [INFO] info msg",
		"[js] [DEBUG] debug msg",
		"[html-report] [WARNING] warning msg",
		"[html-report] [ERROR] error msg",
		"[html-report] [INFO] info msg",
		"[html-report] [DEBUG] debug msg",
		"[Gauge] [DEBUG] debug msg",
	})
}

func tearDown(t *testing.T) {
	config.ProjectRoot = ""
	initialized = false
	if err := os.Truncate(ActiveLogFile, 0); err != nil {
		t.Logf("Could not truncate file")
	}

}

func setupLogger(level string) {
	config.ProjectRoot, _ = filepath.Abs("_testdata")
	Initialize(false, "info", CLI)
}

func newLogWriter(loggerID string) *LogWriter {
	f, _ := os.OpenFile(ActiveLogFile, os.O_RDWR, 0)
	return &LogWriter{
		Stderr: Writer{ShouldWriteToStdout: false, stream: 0, LoggerID: loggerID, File: f},
		Stdout: Writer{ShouldWriteToStdout: false, stream: 0, LoggerID: loggerID, File: f},
	}
}

func assertLogContains(t *testing.T, want []string) {
	got, err := ioutil.ReadFile(ActiveLogFile)
	if err != nil {
		t.Fatalf("Unable to read log file. Error: %s", err.Error())
	}
	for _, w := range want {
		if !strings.Contains(string(got), w) {
			t.Errorf("Expected %s to contain %s.", string(got), w)
		}
	}
}
