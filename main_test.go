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
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/influxdata/influxdb/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
)

const (
	original = "namespace_pod.container:container_cpu_usage_seconds_total:sum_rate"
)

var (
	labels = map[string]string{"name1": "value1", "name2": "value2", "name3": "value3", "name4": "value4"}
	name   = "name"
)

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

func TestReplaceInvalidChars(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
	}{
		{"cpu", "cpu"},
		{"cpu.load:avg", "cpu_load_avg"},
		{"9lives", "_9lives"},
		{"température", "temp_rature"},
		{"温度", "__"},
		{"a b\tc", "a_b_c"},
		{"", ""},
	} {
		got := tc.in
		ReplaceInvalidChars(&got)
		if got != tc.want {
			t.Errorf("ReplaceInvalidChars(%q) = %q, want %q", tc.in, got, tc.want)
		}
		if !utf8.ValidString(got) {
			t.Errorf("ReplaceInvalidChars(%q) = %q, which is not valid UTF-8", tc.in, got)
		}
	}
}

func TestNonASCIIPointDoesNotBreakScrape(t *testing.T) {
	expiry := 5 * time.Minute
	oldExpiry := sampleExpiry
	sampleExpiry = &expiry
	defer func() { sampleExpiry = oldExpiry }()

	c := &influxDBCollector{
		samples: map[string]*influxDBSample{},
		ch:      make(chan *influxDBSample, 10),
		logger:  promslog.NewNopLogger(),
	}
	points, err := models.ParsePointsWithPrecision(
		[]byte("température,zoné=é value=21.5\nplain value=1"), time.Now().UTC(), "ns")
	if err != nil {
		t.Fatal(err)
	}
	c.parsePointsToSample(points)
	for range points {
		s := <-c.ch
		c.samples[s.ID] = s
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(c)
	rr := httptest.NewRecorder()
	promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("scrape returned status %d: %s", rr.Code, rr.Body.String())
	}
	for _, want := range []string{"temp_rature{zon_=\"é\"} 21.5", "plain 1"} {
		if !strings.Contains(rr.Body.String(), want) {
			t.Errorf("scrape output is missing %q:\n%s", want, rr.Body.String())
		}
	}
}
