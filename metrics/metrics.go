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
	"strings"
	"sync"

	graphite "github.com/marpaia/graphite-golang"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Metrics struct {
	Logger  *logrus.Logger
	SyncWG  *sync.WaitGroup
	Channel <-chan []MetricsMsg
}

type MetricsMsg struct {
	Name      string
	Value     string
	Timestamp int64
}

func (m *Metrics) Run() {

	var gClient *graphite.Graphite
	var server string
	var port int
	var err error

	component := "Metrics sender"
	log := m.Logger

	defer m.SyncWG.Done()

	metricsTarget := viper.GetString("metrics.receiver.target")
	switch metricsTarget {
	case "graphite":
		server = viper.GetString("metrics.receiver.graphite.host")
		port = viper.GetInt("metrics.receiver.graphite.port")
		prefix := viper.GetString("metrics.receiver.graphite.prefix")
		//FIXME: validate config parameters

		gClient, err = graphite.NewGraphiteWithMetricPrefix(server, port, prefix)
		if err != nil {
			log.WithFields(logrus.Fields{
				"component": component,
				"Error":     err,
			}).Warn("Unable to spawn graphite sender.")
			return
		}
	default:
		log.WithFields(logrus.Fields{
			"component": component,
		}).Debug("A metrics target was not declared in the config, no metrics will be sent.")
	}

	log.WithFields(logrus.Fields{
		"component":      component,
		"Metrics target": metricsTarget,
		"Server":         server,
		"Port":           port,
	}).Info("Spawned metrics forwarder.")

	for metrics := range m.Channel {
		switch metricsTarget {
		case "graphite":
			go graphiteSend(gClient, metrics, log, m.SyncWG)
			m.SyncWG.Add(1)
		default:
			continue
		}
	}

	log.WithFields(logrus.Fields{
		"component": component,
	}).Debug("Graphite metrics channel closed, goodbye.")

	return
}

func graphiteSend(client *graphite.Graphite, metrics []MetricsMsg, logger *logrus.Logger, wg *sync.WaitGroup) {

	var gMetrics []graphite.Metric
	component := "graphiteSend"

	defer wg.Done()

	for _, metric := range metrics {

		//if a metric starts with '.' or has '..' its invalid, ignore.
		if strings.HasPrefix(metric.Name, ".") || strings.Contains(metric.Name, "..") {
			logger.WithFields(logrus.Fields{
				"component": component,
				"Metric":    fmt.Sprintf("%+v", metric),
			}).Debug("Invalid metric.")
			return
		}

		gMetric := graphite.Metric{
			Name:      metric.Name,
			Value:     metric.Value,
			Timestamp: metric.Timestamp,
		}

		gMetrics = append(gMetrics, gMetric)
	}

	logger.WithFields(logrus.Fields{
		"component": component,
		"Metric":    fmt.Sprintf("%+v", gMetrics),
	}).Debug("Sending metrics...")

	err := client.SendMetrics(gMetrics)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"component": component,
			"Error":     err,
		}).Debug("Unable to send metrics.")
	}

	return
}
