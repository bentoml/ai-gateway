// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import "go.opentelemetry.io/otel/metric"

// mustRegisterCounter registers a Counter with the meter and panics if it fails.
func mustRegisterCounter(meter metric.Meter, name string, options ...metric.Float64CounterOption) metric.Float64Counter {
	h, err := meter.Float64Counter(name, options...)
	if err != nil {
		panic(err)
	}
	return h
}

// mustRegisterHistogram registers a histogram with the meter and panics if it fails.
func mustRegisterHistogram(meter metric.Meter, name string, options ...metric.Float64HistogramOption) metric.Float64Histogram {
	h, err := meter.Float64Histogram(name, options...)
	if err != nil {
		panic(err)
	}
	return h
}

// mustRegisterInt64UpDownCounter registers an Int64UpDownCounter with the meter and panics if it fails.
func mustRegisterInt64UpDownCounter(meter metric.Meter, name string, options ...metric.Int64UpDownCounterOption) metric.Int64UpDownCounter {
	h, err := meter.Int64UpDownCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return h
}
