package metrics

import (
	"fmt"
	"strconv"
	"time"
)

type Emitter struct {
	Channel chan []MetricsMsg
}

// Measures runtime, sends out runtime metrics
func (e *Emitter) MeasureRunTime(start int64, caller string) {

	elapsed := time.Now().Unix() - start

	mPrefix := fmt.Sprintf("%s.%s", caller, "runTime")
	metric := MetricsMsg{
		Name:      mPrefix,
		Value:     strconv.FormatInt(elapsed, 10),
		Timestamp: time.Now().Unix(),
	}

	e.Channel <- []MetricsMsg{metric}
}

// Emits a map[string]int of metrics
func (e *Emitter) EmitMetricMap(metricz map[string]int) {

	var metrics []MetricsMsg
	ts := time.Now().Unix()

	for name, value := range metricz {
		metric := MetricsMsg{
			Name:      name,
			Value:     strconv.Itoa(value),
			Timestamp: ts,
		}
		metrics = append(metrics, metric)
	}

	e.Channel <- metrics
}

func (e *Emitter) Emit(name string, value int) {

	metric := MetricsMsg{
		Name:      name,
		Value:     strconv.Itoa(value),
		Timestamp: time.Now().Unix(),
	}

	e.Channel <- []MetricsMsg{metric}
}
