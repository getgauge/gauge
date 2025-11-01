// Copyright 2019 ThoughtWorks, Inc.

/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package logger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/getgauge/gauge/config"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	_ = os.RemoveAll(filepath.Join("_testdata", "logs"))
	os.Exit(exitCode)
}

func TestLogWriterToOutputInfoLogInCorrectFormat(t *testing.T) {
	defer tearDown(t)
	setupLogger()
	l := newLogWriter("js")

	if _, err := l.Stdout.Write([]byte("{\"logLevel\": \"info\", \"message\": \"Foo\"}\n")); err != nil {
		t.Fatalf("Unable to write to logWriter")
	}

	assertLogContains(t, []string{"[js] [INFO] Foo"})
}

func TestLogWriterToOutputInfoLogInCorrectFormatWhenNewLinesPresent(t *testing.T) {
	defer tearDown(t)
	setupLogger()
	l := newLogWriter("js")

	if _, err := l.Stdout.Write([]byte("{\"logLevel\": \"info\", \"message\": \"Foo\"}\n\r\n{\"logLevel\": \"info\", \"message\": \"Bar\"}\n{\"logLevel\": \"info\", \"message\": \"Baz\"}")); err != nil {
		t.Fatalf("Unable to write to logWriter")
	}

	assertLogContains(t, []string{"[js] [INFO] Foo", "[js] [INFO] Bar", "[js] [INFO] Baz"})
	assertLogDoesNotContains(t, []string{"[js] [INFO] \r"})
}

func TestLogWriterToOutputInfoLogWithMultipleLines(t *testing.T) {
	defer tearDown(t)
	setupLogger()
	l := newLogWriter("js")

	if _, err := l.Stdout.Write([]byte("{\"logLevel\": \"info\", \"message\": \"Foo\"}\n{\"logLevel\": \"debug\", \"message\": \"Bar\"}\n")); err != nil {
		t.Fatalf("Unable to write to logWriter")
	}

	assertLogContains(t, []string{"[js] [INFO] Foo", "[js] [DEBUG] Bar"})
}

func TestLogWriterToLogPlainStrings(t *testing.T) {
	defer tearDown(t)
	setupLogger()
	l := newLogWriter("js")

	if _, err := l.Stdout.Write([]byte("Foo Bar\n")); err != nil {
		t.Fatalf("Unable to write to logWriter")
	}

	assertLogContains(t, []string{"Foo Bar"})
}

func TestUnformattedLogWrittenToStderrShouldBePrefixedWithError(t *testing.T) {
	defer tearDown(t)
	setupLogger()
	l := newLogWriter("js")

	if _, err := l.Stderr.Write([]byte("Foo Bar\n")); err != nil {
		t.Fatalf("Unable to write to logWriter")
	}

	assertLogContains(t, []string{"[ERROR]"})
}

func TestUnformattedLogWrittenToStdoutShouldBePrefixedWithInfo(t *testing.T) {
	defer tearDown(t)
	setupLogger()
	l := newLogWriter("js")

	if _, err := l.Stdout.Write([]byte("Foo Bar\n")); err != nil {
		t.Fatalf("Unable to write to logWriter")
	}

	assertLogContains(t, []string{"[INFO]"})
}

func TestLoggingFromDifferentWritersAtSameTime(t *testing.T) {
	defer tearDown(t)
	setupLogger()
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

func setupLogger() {
	config.ProjectRoot, _ = filepath.Abs("_testdata")
	Initialize(false, "info", CLI)
}

func newLogWriter(loggerID string) *LogWriter {
	f, _ := os.OpenFile(ActiveLogFile, os.O_RDWR, 0)
	return &LogWriter{
		Stderr: Writer{ShouldWriteToStdout: false, stream: 0, LoggerID: loggerID, File: f, isErrorStream: true},
		Stdout: Writer{ShouldWriteToStdout: false, stream: 0, LoggerID: loggerID, File: f},
	}
}

func assertLogContains(t *testing.T, want []string) {
	got, err := os.ReadFile(ActiveLogFile)
	if err != nil {
		t.Fatalf("Unable to read log file. Error: %s", err.Error())
	}
	for _, w := range want {
		if !strings.Contains(string(got), w) {
			t.Errorf("Expected %s to contain %s.", string(got), w)
		}
	}
}

func assertLogDoesNotContains(t *testing.T, want []string) {
	got, err := os.ReadFile(ActiveLogFile)
	if err != nil {
		t.Fatalf("Unable to read log file. Error: %s", err.Error())
	}
	for _, w := range want {
		if strings.Contains(string(got), w) {
			t.Errorf("Expected %s to not contain %s.", string(got), w)
		}
	}
}
