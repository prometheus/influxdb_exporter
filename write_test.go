// Copyright 2026 The Prometheus Authors
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
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWritePrecision(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		path      string
		precision string
		timestamp int64
	}{
		{"/write", "", now.UnixNano()},
		{"/write", "n", now.UnixNano()},
		{"/write", "u", now.UnixMicro()},
		{"/write", "ms", now.UnixMilli()},
		{"/write", "s", now.Unix()},
		{"/api/v2/write", "ns", now.UnixNano()},
		{"/api/v2/write", "us", now.UnixMicro()},
		{"/api/v2/write", "ms", now.UnixMilli()},
		{"/api/v2/write", "s", now.Unix()},
	}

	for _, tt := range tests {
		t.Run(tt.path+"?precision="+tt.precision, func(t *testing.T) {
			c := newInfluxDBCollector(slog.New(slog.DiscardHandler))

			url := tt.path
			if tt.precision != "" {
				url += "?precision=" + tt.precision
			}
			body := fmt.Sprintf("cpu,host=a value=1 %d\n", tt.timestamp)
			req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
			rec := httptest.NewRecorder()
			c.influxDBPost(rec, req)
			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, body %q", rec.Code, rec.Body.String())
			}

			var got time.Time
			deadline := time.Now().Add(time.Second)
			for time.Now().Before(deadline) {
				c.mu.Lock()
				s, ok := c.samples["cpu.host.a"]
				if ok {
					got = s.Timestamp
				}
				c.mu.Unlock()
				if ok {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}

			if !got.Equal(now) {
				t.Errorf("sample timestamp = %v, want %v", got, now)
			}
		})
	}
}
