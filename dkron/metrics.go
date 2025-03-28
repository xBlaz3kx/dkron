package dkron

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/hashicorp/go-metrics/datadog"
	"github.com/hashicorp/go-metrics/prometheus"
)

func initMetrics(a *Agent) error {
	// Setup the inmem sink and signal handler
	inm := metrics.NewInmemSink(10*time.Second, time.Minute)
	metrics.DefaultInmemSignal(inm)

	var fanout metrics.FanoutSink

	// Configure the prometheus sink
	if a.config.EnablePrometheus {
		promSink, err := prometheus.NewPrometheusSink()
		if err != nil {
			return err
		}

		fanout = append(fanout, promSink)
	}

	// Configure the statsd sink
	if a.config.StatsdAddr != "" {
		sink, err := metrics.NewStatsdSink(a.config.StatsdAddr)
		if err != nil {
			return fmt.Errorf("failed to start statsd sink. Got: %s", err)
		}
		fanout = append(fanout, sink)
	}

	// Configure the DogStatsd sink
	if a.config.DogStatsdAddr != "" {
		var tags []string

		if a.config.DogStatsdTags != nil {
			tags = a.config.DogStatsdTags
		}

		sink, err := datadog.NewDogStatsdSink(a.config.DogStatsdAddr, a.config.NodeName)
		if err != nil {
			return fmt.Errorf("failed to start DogStatsd sink. Got: %s", err)
		}
		sink.SetTags(tags)
		fanout = append(fanout, sink)
	}

	// Initialize the global sink
	if len(fanout) > 0 {
		fanout = append(fanout, inm)
		if _, err := metrics.NewGlobal(metrics.DefaultConfig("dkron"), fanout); err != nil {
			return err
		}
	} else {
		if _, err := metrics.NewGlobal(metrics.DefaultConfig("dkron"), inm); err != nil {
			return err
		}
	}

	return nil
}
