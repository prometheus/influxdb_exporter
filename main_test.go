// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

const (
	original = "namespace_pod.container:container_cpu_usage_seconds_total:sum_rate"

	// Sample InfluxDB line protocol payload used by the HTTP write tests. The
	// timestamp is in nanoseconds, matching the default precision used by
	// influxDBPost when the request omits the "precision" form value.
	sampleInfluxLine = "weather,location=us-midwest temperature=82 1465839830100400200"
)

var (
	labels = map[string]string{"name1": "value1", "name2": "value2", "name3": "value3", "name4": "value4"}
	name   = "name"
)

// gzipBytes returns data compressed with the default gzip writer.
func gzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

// waitForSamples polls the collector until at least want samples have been
// processed off the channel into the samples map, or until the deadline.
func waitForSamples(t *testing.T, c *influxDBCollector, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		got := len(c.samples)
		c.mu.Unlock()
		if got >= want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	c.mu.Lock()
	got := len(c.samples)
	c.mu.Unlock()
	t.Fatalf("timed out waiting for %d samples (got %d)", want, got)
}

func newTestCollector() *influxDBCollector {
	return newInfluxDBCollector(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// TestInfluxDBPostGzip verifies that a gzip-compressed line-protocol payload
// (as sent by Telegraf and other clients with Content-Encoding: gzip) is
// accepted, decompressed, parsed, and turned into a sample. This is the path
// originally introduced in #78.
func TestInfluxDBPostGzip(t *testing.T) {
	c := newTestCollector()

	body := gzipBytes(t, []byte(sampleInfluxLine))
	req := httptest.NewRequest(http.MethodPost, "/write", bytes.NewReader(body))
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()

	c.influxDBPost(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d (body: %q)", rr.Code, http.StatusNoContent, rr.Body.String())
	}
	waitForSamples(t, c, 1)
}

// TestInfluxDBPostUncompressed is the baseline path: the same payload posted
// without Content-Encoding must also be accepted and produce a sample.
func TestInfluxDBPostUncompressed(t *testing.T) {
	c := newTestCollector()

	req := httptest.NewRequest(http.MethodPost, "/write", strings.NewReader(sampleInfluxLine))
	rr := httptest.NewRecorder()

	c.influxDBPost(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d (body: %q)", rr.Code, http.StatusNoContent, rr.Body.String())
	}
	waitForSamples(t, c, 1)
}

// TestInfluxDBPostGzipMalformed verifies that a request advertising
// Content-Encoding: gzip but carrying a body that cannot be gunzipped is
// rejected with a 500 rather than crashing or being treated as plaintext.
func TestInfluxDBPostGzipMalformed(t *testing.T) {
	c := newTestCollector()

	req := httptest.NewRequest(http.MethodPost, "/write", strings.NewReader("not actually gzip"))
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()

	c.influxDBPost(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d (body: %q)", rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

func BenchmarkRegexpReplaceInvalid(b *testing.B) {
	b.ReportAllocs()
	invalidChars := regexp.MustCompile("[^a-zA-Z0-9_]")

	for i := 0; i < b.N; i++ {
		invalidChars.ReplaceAllString(original, "_")
	}
}

func BenchmarkHardcodedReplace(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var newString = original
		ReplaceInvalidChars(&newString)
	}
}

func BenchmarkSprintfArray(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Calculate a consistent unique ID for the sample.
		labelnames := make([]string, 0, len(labels))
		for k := range labels {
			labelnames = append(labelnames, k)
		}
		sort.Strings(labelnames)
		parts := make([]string, 0, len(labels)*2+1)
		parts = append(parts, name)
		for _, l := range labelnames {
			parts = append(parts, l, labels[l])
		}
		fmt.Sprintf("%q", parts)
	}
}

func BenchmarkStringJoin(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {

		// Calculate a consistent unique ID for the sample.
		labelnames := make([]string, 0, len(labels))
		for k := range labels {
			labelnames = append(labelnames, k)
		}
		sort.Strings(labelnames)
		parts := make([]string, 0, len(labels)*2+1)
		parts = append(parts, name)
		for _, l := range labelnames {
			parts = append(parts, l, labels[l])
		}
		strings.Join(parts, ".")
	}
}
