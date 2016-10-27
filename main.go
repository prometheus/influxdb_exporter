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
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"

	"github.com/influxdata/influxdb/models"
)

const (
	MAX_UDP_PAYLOAD = 64 * 1024
)

var (
	showVersion   = flag.Bool("version", false, "Print version information.")
	listenAddress = flag.String("web.listen-address", ":9122", "Address on which to expose metrics and web interface.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose Prometheus metrics.")
	sampleExpiry  = flag.Duration("influxdb.sample-expiry", 5*time.Minute, "How long a sample is valid for.")
	lastPush      = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "influxdb_last_push_timestamp_seconds",
			Help: "Unix timestamp of the last received influxdb metrics push in seconds.",
		},
	)
	invalidChars = regexp.MustCompile("[^a-zA-Z0-9_]")

	// Udp
	bindAddress = flag.String("udp.bind-address", ":9122", "Address on which to listen for udp packets.")
	listenOnUdp = flag.Bool("udp.listen", false, "Listen on udp socket.")
)

type influxDBSample struct {
	ID        string
	Name      string
	Labels    map[string]string
	Value     float64
	Timestamp time.Time
}

func (c *influxDBCollector) serveUdp() {
	buf := make([]byte, MAX_UDP_PAYLOAD)
	for {

		select {
		default:
			n, _, err := c.conn.ReadFromUDP(buf)
			if err != nil {
				log.Warnf("Failed to read UDP message: %s", err)
				continue
			}

			bufCopy := make([]byte, n)
			copy(bufCopy, buf[:n])

			precision := "ns"
			points, err := models.ParsePointsWithPrecision(bufCopy, time.Now().UTC(), precision)
			if err != nil {
				log.Errorf("error parsing udp packet: %s", err)
				c.udpParseErrors.Inc()
				return
			}

			c.parsePointsToSample(points)
		}
	}
}

type influxDBCollector struct {
	samples map[string]*influxDBSample
	mu      sync.Mutex
	ch      chan *influxDBSample

	// Udp
	conn           *net.UDPConn
	udpParseErrors prometheus.Counter
}

func newInfluxDBCollector() *influxDBCollector {
	c := &influxDBCollector{
		ch:      make(chan *influxDBSample),
		samples: map[string]*influxDBSample{},
		udpParseErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "influxdb",
			Name:      "exporter_udp_parse_errors",
			Help:      "Current total udp parse errors.",
		}),
	}
	go c.processSamples()
	return c
}

func (c *influxDBCollector) influxDBPost(w http.ResponseWriter, r *http.Request) {
	lastPush.Set(float64(time.Now().UnixNano()) / 1e9)
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("error reading body: %s", err), 500)
		return
	}

	precision := "ns"
	if r.FormValue("precision") != "" {
		precision = r.FormValue("precision")
	}
	points, err := models.ParsePointsWithPrecision(buf, time.Now().UTC(), precision)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing request: %s", err), 400)
		return
	}

	c.parsePointsToSample(points)

	// InfluxDB returns a 204 on success.
	http.Error(w, "", 204)
}

func (c *influxDBCollector) parsePointsToSample(points []models.Point) {
	for _, s := range points {
		for field, v := range s.Fields() {
			var value float64
			switch v := v.(type) {
			case float64:
				value = v
			case int64:
				value = float64(v)
			case bool:
				if v {
					value = 1
				} else {
					value = 0
				}
			default:
				continue
			}

			var name string
			if field == "value" {
				name = s.Name()
			} else {
				name = fmt.Sprintf("%s_%s", s.Name(), field)
			}

			sample := &influxDBSample{
				Name:      invalidChars.ReplaceAllString(name, "_"),
				Timestamp: s.Time(),
				Value:     value,
				Labels:    map[string]string{},
			}
			for k, v := range s.Tags() {
				sample.Labels[invalidChars.ReplaceAllString(k, "_")] = v
			}

			// Calculate a consistent unique ID for the sample.
			labelnames := make([]string, 0, len(sample.Labels))
			for k := range sample.Labels {
				labelnames = append(labelnames, k)
			}
			sort.Strings(labelnames)
			parts := make([]string, 0, len(sample.Labels)*2+1)
			parts = append(parts, name)
			for _, l := range labelnames {
				parts = append(parts, l, sample.Labels[l])
			}
			sample.ID = fmt.Sprintf("%q", parts)

			c.ch <- sample
		}
	}
}

func (c *influxDBCollector) processSamples() {
	ticker := time.NewTicker(time.Minute).C
	for {
		select {
		case s := <-c.ch:
			c.mu.Lock()
			c.samples[s.ID] = s
			c.mu.Unlock()

		case <-ticker:
			// Garbage collect expired value lists.
			ageLimit := time.Now().Add(-*sampleExpiry)
			c.mu.Lock()
			for k, sample := range c.samples {
				if ageLimit.After(sample.Timestamp) {
					delete(c.samples, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

// Collect implements prometheus.Collector.
func (c *influxDBCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- lastPush

	c.mu.Lock()
	samples := make([]*influxDBSample, 0, len(c.samples))
	for _, sample := range c.samples {
		samples = append(samples, sample)
	}
	c.mu.Unlock()

	ageLimit := time.Now().Add(-*sampleExpiry)
	for _, sample := range samples {
		if ageLimit.After(sample.Timestamp) {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(sample.Name, "InfluxDB Metric", []string{}, sample.Labels),
			prometheus.UntypedValue,
			sample.Value,
		)
	}
}

// Describe implements prometheus.Collector.
func (c *influxDBCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- lastPush.Desc()
}

func init() {
	prometheus.MustRegister(version.NewCollector("influxdb_exporter"))
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("influxdb_exporter"))
		os.Exit(0)
	}

	log.Infoln("Starting influxdb_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	c := newInfluxDBCollector()
	prometheus.MustRegister(c)
	prometheus.MustRegister(c.udpParseErrors)

	if *listenOnUdp {
		addr, err := net.ResolveUDPAddr("udp", *bindAddress)
		if err != nil {
			fmt.Printf("Failed to resolve UDP address %s: %s", *bindAddress, err)
			os.Exit(1)
		}

		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			fmt.Printf("Failed to set up UDP listener at address %s: %s", addr, err)
			os.Exit(1)
		}

		c.conn = conn
		go c.serveUdp()
	}

	http.HandleFunc("/write", c.influxDBPost)
	// Some InfluxDB clients try to create a database.
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"results": []}`)
	})

	http.Handle(*metricsPath, prometheus.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
    <head><title>InfluxDB Exporter</title></head>
    <body>
    <h1>InfluxDB Exporter</h1>
    <p><a href="` + *metricsPath + `">Metrics</a></p>
    </body>
    </html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	http.ListenAndServe(*listenAddress, nil)
}
