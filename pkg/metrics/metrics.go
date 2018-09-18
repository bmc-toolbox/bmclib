// Copyright © 2018 Joel Rebello <joel.rebello@booking.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"fmt"
	"github.com/cyberdelia/go-metrics-graphite"
	gometrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmcbutler/pkg/config"
)

var (
	metricsChan chan Metric
)

//TODO:
// Implement a counter increment method that accepts string, float32 value
// increment method sends the metric down the channel
// a go routine reads from the channel and updates the metricsData map

type Emitter struct {
	Config      *config.Params
	Logger      *logrus.Logger
	Registry    gometrics.Registry
	metricsChan chan Metric
	metricsData map[string]map[string]float32
}

type Metric struct {
	mType string   //counter/gauge
	mKey  []string //metric key
	mVal  float32  //metric value
}

// init sets up external and internal metric sinks.
func (m *Emitter) Init() {

	var host, prefix string
	var port int
	var flushInterval time.Duration

	m.metricsChan = make(chan Metric)
	m.metricsData = make(map[string]map[string]float32)

	component := "Metrics emitter"
	log := m.Logger

	//go routine that records metricsData
	go m.store()

	//setup metrics client based on config
	client := m.Config.MetricsParams.Client
	switch client {
	case "graphite":
		host = m.Config.MetricsParams.Host
		port = m.Config.MetricsParams.Port
		prefix = m.Config.MetricsParams.Prefix
		flushInterval = m.Config.MetricsParams.FlushInterval

		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			log.WithFields(logrus.Fields{
				"component":      component,
				"Metrics client": client,
				"Server":         host,
				"Port":           port,
				"Error":          err,
			}).Warn("Error resolving tcp addr.")
		}

		m.Registry = gometrics.NewRegistry()

		go graphite.Graphite(gometrics.DefaultRegistry,
			flushInterval,
			prefix,
			addr)
	default:
		//unknown metrics client declared
		log.WithFields(logrus.Fields{
			"component":      component,
			"Metrics client": client,
		}).Debug("Unknown/no metrics client declared in config.")
	}

}

//- writes/updates metric key/vals to metricsData
//- register and write metrics to the go-metrics registries.
func (m *Emitter) store() {

	//A map of metric names to go-metrics registry
	//the keys to this map could be of type metrics.Counter/metrics.Gauge
	goMetricsRegistry := make(map[string]interface{})

	for {
		data, ok := <-m.metricsChan
		if !ok {
			return
		}

		mIdentifier := data.mKey[0]
		mKey := strings.Join(data.mKey, ".")

		_, keyExists := m.metricsData[mIdentifier]
		if !keyExists {
			m.metricsData[mIdentifier] = make(map[string]float32)
		}

		//register the metric with go-metrics,
		//the metric key is used as the identifier.
		_, registryExists := goMetricsRegistry[mIdentifier]
		if !registryExists {
			switch data.mType {
			case "counter":
				c := gometrics.NewCounter()
				gometrics.Register(mKey, c)
				goMetricsRegistry[mKey] = c
			case "gauge":
				g := gometrics.NewGauge()
				gometrics.Register(mKey, g)
				goMetricsRegistry[mKey] = g
			}
		}

		//based on the metric type, update the store/registry.
		switch data.mType {
		case "counter":
			m.metricsData[mIdentifier][mKey] += data.mVal

			//type assert metrics registry to its type - metrics.Counter
			//type cast float32 metric value type to int64
			goMetricsRegistry[mKey].(gometrics.Counter).Inc(
				int64(m.metricsData[mIdentifier][mKey]))
		case "gauge":
			m.metricsData[mIdentifier][mKey] = data.mVal

			//type assert metrics registry to its type - metrics.Gauge
			//type cast float32 metric value type to int64
			goMetricsRegistry[mKey].(gometrics.Gauge).Update(
				int64(m.metricsData[mIdentifier][mKey]))
		}
	}
}

//Logs current metrics
func (m *Emitter) dumpStats() {

	for mSource, metrics_ := range m.metricsData {

		var metricStr string
		for k, v := range metrics_ {
			metricStr += fmt.Sprintf("%s: %f ", k, v)
		}

		m.Logger.WithFields(logrus.Fields{
			"data": metricStr,
		}).Info(fmt.Sprintf("Metrics: %s", mSource))
	}
}

//Increment counter metric
//key = slice of strings that will be joined with "." to be used as the metric namespace
//val = float32 metric value
func (m *Emitter) IncrCounter(key []string, val float32) {

	d := Metric{
		mType: "counter",
		mKey:  key,
		mVal:  val,
	}

	m.metricsChan <- d
}

//Set gauge metric
//key = slice of strings that will be joined with "." to be used as the metric namespace
//val = float32 metric value
func (m *Emitter) UpdateGauge(key []string, val float32) {

	d := Metric{
		mType: "gauge",
		mKey:  key,
		mVal:  val,
	}

	m.metricsChan <- d
}

//Measure time elapsed since invocation
func (m *Emitter) MeasureRuntime(key []string, start time.Time) {

	//convert time.Duration to miliseconds
	elapsed := float32(time.Since(start).Seconds() * 1e3) //1e3 == 1000
	m.UpdateGauge(key, elapsed)
}

// Any emmiter related clean up actions go here
func (m *Emitter) Close(printStats bool) {
	close(m.metricsChan)

	if printStats {
		m.dumpStats()
	}
}
