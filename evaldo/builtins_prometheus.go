//go:build !no_prometheus
// +build !no_prometheus

package evaldo

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/refaktor/rye/env"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Function to create a new Prometheus counter
func __prometheus_new_counter(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch name := arg0.(type) {
	case env.String:
		switch help := arg1.(type) {
		case env.String:
			// Check if we have labels
			var constLabels prometheus.Labels
			if arg2 != nil {
				switch labels := arg2.(type) {
				case env.Dict:
					constLabels = make(prometheus.Labels)
					for k, v := range labels.Data {
						switch val := v.(type) {
						case env.String:
							constLabels[k] = val.Value
						default:
							return MakeBuiltinError(ps, "Label values must be strings", "prometheus//new-counter")
						}
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.DictType}, "prometheus//new-counter")
				}
			}

			counter := prometheus.NewCounter(prometheus.CounterOpts{
				Name:        name.Value,
				Help:        help.Value,
				ConstLabels: constLabels,
			})

			// Register the counter
			err := prometheus.Register(counter)
			if err != nil {
				return MakeBuiltinError(ps, fmt.Sprintf("Failed to register counter: %v", err), "prometheus//new-counter")
			}

			return *env.NewNative(ps.Idx, counter, "prometheus-counter")
		default:
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "prometheus//new-counter")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.StringType}, "prometheus//new-counter")
	}
}

// Function to increment a Prometheus counter
func __prometheus_counter_inc(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch counter := arg0.(type) {
	case env.Native:
		if ps.Idx.GetWord(counter.Kind.Index) != "prometheus-counter" {
			return MakeBuiltinError(ps, "Expected a prometheus-counter", "prometheus-counter//inc")
		}

		counter.Value.(prometheus.Counter).Inc()
		return arg0
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus-counter//inc")
	}
}

// Function to add a value to a Prometheus counter
func __prometheus_counter_add(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch counter := arg0.(type) {
	case env.Native:
		if ps.Idx.GetWord(counter.Kind.Index) != "prometheus-counter" {
			return MakeBuiltinError(ps, "Expected a prometheus-counter", "prometheus-counter//add")
		}

		switch value := arg1.(type) {
		case env.Integer:
			counter.Value.(prometheus.Counter).Add(float64(value.Value))
			return arg0
		case env.Decimal:
			counter.Value.(prometheus.Counter).Add(value.Value)
			return arg0
		default:
			return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "prometheus-counter//add")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus-counter//add")
	}
}

// Function to create a new Prometheus gauge
func __prometheus_new_gauge(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch name := arg0.(type) {
	case env.String:
		switch help := arg1.(type) {
		case env.String:
			// Check if we have labels
			var constLabels prometheus.Labels
			if arg2 != nil {
				switch labels := arg2.(type) {
				case env.Dict:
					constLabels = make(prometheus.Labels)
					for k, v := range labels.Data {
						switch val := v.(type) {
						case env.String:
							constLabels[k] = val.Value
						default:
							return MakeBuiltinError(ps, "Label values must be strings", "prometheus//new-gauge")
						}
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.DictType}, "prometheus//new-gauge")
				}
			}

			gauge := prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        name.Value,
				Help:        help.Value,
				ConstLabels: constLabels,
			})

			// Register the gauge
			err := prometheus.Register(gauge)
			if err != nil {
				return MakeBuiltinError(ps, fmt.Sprintf("Failed to register gauge: %v", err), "prometheus//new-gauge")
			}

			return *env.NewNative(ps.Idx, gauge, "prometheus-gauge")
		default:
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "prometheus//new-gauge")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.StringType}, "prometheus//new-gauge")
	}
}

// Function to set a Prometheus gauge value
func __prometheus_gauge_set(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch gauge := arg0.(type) {
	case env.Native:
		if ps.Idx.GetWord(gauge.Kind.Index) != "prometheus-gauge" {
			return MakeBuiltinError(ps, "Expected a prometheus-gauge", "prometheus-gauge//set")
		}

		switch value := arg1.(type) {
		case env.Integer:
			gauge.Value.(prometheus.Gauge).Set(float64(value.Value))
			return arg0
		case env.Decimal:
			gauge.Value.(prometheus.Gauge).Set(value.Value)
			return arg0
		default:
			return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "prometheus-gauge//set")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus-gauge//set")
	}
}

// Function to increment a Prometheus gauge
func __prometheus_gauge_inc(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch gauge := arg0.(type) {
	case env.Native:
		if ps.Idx.GetWord(gauge.Kind.Index) != "prometheus-gauge" {
			return MakeBuiltinError(ps, "Expected a prometheus-gauge", "prometheus-gauge//inc")
		}

		gauge.Value.(prometheus.Gauge).Inc()
		return arg0
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus-gauge//inc")
	}
}

// Function to decrement a Prometheus gauge
func __prometheus_gauge_dec(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch gauge := arg0.(type) {
	case env.Native:
		if ps.Idx.GetWord(gauge.Kind.Index) != "prometheus-gauge" {
			return MakeBuiltinError(ps, "Expected a prometheus-gauge", "prometheus-gauge//dec")
		}

		gauge.Value.(prometheus.Gauge).Dec()
		return arg0
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus-gauge//dec")
	}
}

// Function to add a value to a Prometheus gauge
func __prometheus_gauge_add(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch gauge := arg0.(type) {
	case env.Native:
		if ps.Idx.GetWord(gauge.Kind.Index) != "prometheus-gauge" {
			return MakeBuiltinError(ps, "Expected a prometheus-gauge", "prometheus-gauge//add")
		}

		switch value := arg1.(type) {
		case env.Integer:
			gauge.Value.(prometheus.Gauge).Add(float64(value.Value))
			return arg0
		case env.Decimal:
			gauge.Value.(prometheus.Gauge).Add(value.Value)
			return arg0
		default:
			return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "prometheus-gauge//add")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus-gauge//add")
	}
}

// Function to subtract a value from a Prometheus gauge
func __prometheus_gauge_sub(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch gauge := arg0.(type) {
	case env.Native:
		if ps.Idx.GetWord(gauge.Kind.Index) != "prometheus-gauge" {
			return MakeBuiltinError(ps, "Expected a prometheus-gauge", "prometheus-gauge//sub")
		}

		switch value := arg1.(type) {
		case env.Integer:
			gauge.Value.(prometheus.Gauge).Sub(float64(value.Value))
			return arg0
		case env.Decimal:
			gauge.Value.(prometheus.Gauge).Sub(value.Value)
			return arg0
		default:
			return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "prometheus-gauge//sub")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus-gauge//sub")
	}
}

// Function to create a new Prometheus histogram
func __prometheus_new_histogram(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch name := arg0.(type) {
	case env.String:
		switch help := arg1.(type) {
		case env.String:
			// Check if we have buckets
			var buckets []float64
			if arg2 != nil {
				switch bucketsObj := arg2.(type) {
				case env.Block:
					buckets = make([]float64, bucketsObj.Series.Len())
					for i := 0; i < bucketsObj.Series.Len(); i++ {
						switch val := bucketsObj.Series.Get(i).(type) {
						case env.Integer:
							buckets[i] = float64(val.Value)
						case env.Decimal:
							buckets[i] = val.Value
						default:
							return MakeBuiltinError(ps, "Bucket values must be numbers", "prometheus//new-histogram")
						}
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "prometheus//new-histogram")
				}
			}

			// Check if we have labels
			var constLabels prometheus.Labels
			if arg3 != nil {
				switch labels := arg3.(type) {
				case env.Dict:
					constLabels = make(prometheus.Labels)
					for k, v := range labels.Data {
						switch val := v.(type) {
						case env.String:
							constLabels[k] = val.Value
						default:
							return MakeBuiltinError(ps, "Label values must be strings", "prometheus//new-histogram")
						}
					}
				default:
					return MakeArgError(ps, 4, []env.Type{env.DictType}, "prometheus//new-histogram")
				}
			}

			var histogram prometheus.Histogram
			if len(buckets) > 0 {
				histogram = prometheus.NewHistogram(prometheus.HistogramOpts{
					Name:        name.Value,
					Help:        help.Value,
					Buckets:     buckets,
					ConstLabels: constLabels,
				})
			} else {
				histogram = prometheus.NewHistogram(prometheus.HistogramOpts{
					Name:        name.Value,
					Help:        help.Value,
					ConstLabels: constLabels,
				})
			}

			// Register the histogram
			err := prometheus.Register(histogram)
			if err != nil {
				return MakeBuiltinError(ps, fmt.Sprintf("Failed to register histogram: %v", err), "prometheus//new-histogram")
			}

			return *env.NewNative(ps.Idx, histogram, "prometheus-histogram")
		default:
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "prometheus//new-histogram")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.StringType}, "prometheus//new-histogram")
	}
}

// Function to observe a value in a Prometheus histogram
func __prometheus_histogram_observe(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch histogram := arg0.(type) {
	case env.Native:
		if ps.Idx.GetWord(histogram.Kind.Index) != "prometheus-histogram" {
			return MakeBuiltinError(ps, "Expected a prometheus-histogram", "prometheus-histogram//observe")
		}

		switch value := arg1.(type) {
		case env.Integer:
			histogram.Value.(prometheus.Histogram).Observe(float64(value.Value))
			return arg0
		case env.Decimal:
			histogram.Value.(prometheus.Histogram).Observe(value.Value)
			return arg0
		default:
			return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "prometheus-histogram//observe")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus-histogram//observe")
	}
}

// Function to create a new Prometheus summary
func __prometheus_new_summary(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch name := arg0.(type) {
	case env.String:
		switch help := arg1.(type) {
		case env.String:
			// Check if we have objectives
			var objectives map[float64]float64
			if arg2 != nil {
				switch objectivesObj := arg2.(type) {
				case env.Dict:
					objectives = make(map[float64]float64)
					for k, v := range objectivesObj.Data {
						var quantile float64
						var error float64

						// Try to parse the key as a float64 (quantile)
						quantileStr := k
						if q, err := strconv.ParseFloat(quantileStr, 64); err == nil {
							quantile = q
						} else {
							return MakeBuiltinError(ps, "Quantile must be a valid float64", "prometheus//new-summary")
						}

						// Try to parse the value as a float64 (error)
						switch val := v.(type) {
						case env.Integer:
							error = float64(val.Value)
						case env.Decimal:
							error = val.Value
						default:
							return MakeBuiltinError(ps, "Error values must be numbers", "prometheus//new-summary")
						}

						objectives[quantile] = error
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.DictType}, "prometheus//new-summary")
				}
			}

			// Check if we have labels
			var constLabels prometheus.Labels
			if arg3 != nil {
				switch labels := arg3.(type) {
				case env.Dict:
					constLabels = make(prometheus.Labels)
					for k, v := range labels.Data {
						switch val := v.(type) {
						case env.String:
							constLabels[k] = val.Value
						default:
							return MakeBuiltinError(ps, "Label values must be strings", "prometheus//new-summary")
						}
					}
				default:
					return MakeArgError(ps, 4, []env.Type{env.DictType}, "prometheus//new-summary")
				}
			}

			var summary prometheus.Summary
			if objectives != nil {
				summary = prometheus.NewSummary(prometheus.SummaryOpts{
					Name:        name.Value,
					Help:        help.Value,
					Objectives:  objectives,
					ConstLabels: constLabels,
				})
			} else {
				summary = prometheus.NewSummary(prometheus.SummaryOpts{
					Name:        name.Value,
					Help:        help.Value,
					ConstLabels: constLabels,
				})
			}

			// Register the summary
			err := prometheus.Register(summary)
			if err != nil {
				return MakeBuiltinError(ps, fmt.Sprintf("Failed to register summary: %v", err), "prometheus//new-summary")
			}

			return *env.NewNative(ps.Idx, summary, "prometheus-summary")
		default:
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "prometheus//new-summary")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.StringType}, "prometheus//new-summary")
	}
}

// Function to observe a value in a Prometheus summary
func __prometheus_summary_observe(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch summary := arg0.(type) {
	case env.Native:
		if ps.Idx.GetWord(summary.Kind.Index) != "prometheus-summary" {
			return MakeBuiltinError(ps, "Expected a prometheus-summary", "prometheus-summary//observe")
		}

		switch value := arg1.(type) {
		case env.Integer:
			summary.Value.(prometheus.Summary).Observe(float64(value.Value))
			return arg0
		case env.Decimal:
			summary.Value.(prometheus.Summary).Observe(value.Value)
			return arg0
		default:
			return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "prometheus-summary//observe")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus-summary//observe")
	}
}

// Function to register a collector with MustRegister
func __prometheus_must_register(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch collector := arg0.(type) {
	case env.Native:
		// Check if the native object is a Prometheus collector
		kindName := ps.Idx.GetWord(collector.Kind.Index)
		if kindName != "prometheus-counter" &&
			kindName != "prometheus-gauge" &&
			kindName != "prometheus-histogram" &&
			kindName != "prometheus-summary" {
			return MakeBuiltinError(ps, "Expected a Prometheus collector", "prometheus//must-register")
		}

		// Use MustRegister to register the collector
		// This will panic if registration fails
		prometheus.MustRegister(collector.Value.(prometheus.Collector))
		return arg0
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus//must-register")
	}
}

// Function to get a Prometheus HTTP handler
func __prometheus_handler(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	// Return the Prometheus HTTP handler
	handler := promhttp.Handler()
	return *env.NewNative(ps.Idx, handler, "prometheus-handler")
}

// Function to start a Prometheus HTTP server
func __prometheus_start_http_server(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch addr := arg0.(type) {
	case env.String:
		// Set up the HTTP server for Prometheus metrics
		http.Handle("/metrics", promhttp.Handler())

		// Start the server in a goroutine
		go func() {
			if err := http.ListenAndServe(addr.Value, nil); err != nil {
				fmt.Printf("Failed to start Prometheus HTTP server: %v\n", err)
			}
		}()

		return *env.NewInteger(1)
	default:
		return MakeArgError(ps, 1, []env.Type{env.StringType}, "prometheus//start-http-server")
	}
}

// Function to create a mutex to protect metrics from concurrent access
func __prometheus_new_mutex(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	mutex := &sync.Mutex{}
	return *env.NewNative(ps.Idx, mutex, "prometheus-mutex")
}

// Function to lock a mutex
func __prometheus_mutex_lock(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mutex := arg0.(type) {
	case env.Native:
		if ps.Idx.GetWord(mutex.Kind.Index) != "prometheus-mutex" {
			return MakeBuiltinError(ps, "Expected a prometheus-mutex", "prometheus-mutex//lock")
		}

		mutex.Value.(*sync.Mutex).Lock()
		return arg0
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus-mutex//lock")
	}
}

// Function to unlock a mutex
func __prometheus_mutex_unlock(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mutex := arg0.(type) {
	case env.Native:
		if ps.Idx.GetWord(mutex.Kind.Index) != "prometheus-mutex" {
			return MakeBuiltinError(ps, "Expected a prometheus-mutex", "prometheus-mutex//unlock")
		}

		mutex.Value.(*sync.Mutex).Unlock()
		return arg0
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "prometheus-mutex//unlock")
	}
}

// Map of Prometheus builtins
var Builtins_prometheus = map[string]*env.Builtin{
	// Counter functions
	"new-counter": {
		Argsn: 3,
		Doc:   "Creates a new Prometheus counter with the given name, help text, and optional labels.",
		Fn:    __prometheus_new_counter,
	},
	"prometheus-counter//Inc": {
		Argsn: 1,
		Doc:   "Increments a Prometheus counter by 1.",
		Fn:    __prometheus_counter_inc,
	},
	"prometheus-counter//Add": {
		Argsn: 2,
		Doc:   "Adds the given value to a Prometheus counter.",
		Fn:    __prometheus_counter_add,
	},

	// Gauge functions
	"new-gauge": {
		Argsn: 3,
		Doc:   "Creates a new Prometheus gauge with the given name, help text, and optional labels.",
		Fn:    __prometheus_new_gauge,
	},
	"prometheus-gauge//Set": {
		Argsn: 2,
		Doc:   "Sets a Prometheus gauge to the given value.",
		Fn:    __prometheus_gauge_set,
	},
	"prometheus-gauge//Inc": {
		Argsn: 1,
		Doc:   "Increments a Prometheus gauge by 1.",
		Fn:    __prometheus_gauge_inc,
	},
	"prometheus-gauge//Dec": {
		Argsn: 1,
		Doc:   "Decrements a Prometheus gauge by 1.",
		Fn:    __prometheus_gauge_dec,
	},
	"prometheus-gauge//Add": {
		Argsn: 2,
		Doc:   "Adds the given value to a Prometheus gauge.",
		Fn:    __prometheus_gauge_add,
	},
	"prometheus-gauge//Sub": {
		Argsn: 2,
		Doc:   "Subtracts the given value from a Prometheus gauge.",
		Fn:    __prometheus_gauge_sub,
	},

	// Histogram functions
	"new-histogram": {
		Argsn: 4,
		Doc:   "Creates a new Prometheus histogram with the given name, help text, optional buckets, and optional labels.",
		Fn:    __prometheus_new_histogram,
	},
	"prometheus-histogram//Observe": {
		Argsn: 2,
		Doc:   "Observes the given value in a Prometheus histogram.",
		Fn:    __prometheus_histogram_observe,
	},

	// Summary functions
	"new-summary": {
		Argsn: 4,
		Doc:   "Creates a new Prometheus summary with the given name, help text, optional objectives, and optional labels.",
		Fn:    __prometheus_new_summary,
	},
	"prometheus-summary//Observe": {
		Argsn: 2,
		Doc:   "Observes the given value in a Prometheus summary.",
		Fn:    __prometheus_summary_observe,
	},

	// HTTP server functions
	"start-http-server": {
		Argsn: 1,
		Doc:   "Starts a Prometheus HTTP server on the given address (e.g., ':8080').",
		Fn:    __prometheus_start_http_server,
	},

	// Mutex functions for thread safety
	"new-mutex": {
		Argsn: 0,
		Doc:   "Creates a new mutex for protecting metrics from concurrent access.",
		Fn:    __prometheus_new_mutex,
	},
	"prometheus-mutex//Lock": {
		Argsn: 1,
		Doc:   "Locks a mutex.",
		Fn:    __prometheus_mutex_lock,
	},
	"prometheus-mutex//Unlock": {
		Argsn: 1,
		Doc:   "Unlocks a mutex.",
		Fn:    __prometheus_mutex_unlock,
	},

	// Additional functions
	"must-register": {
		Argsn: 1,
		Doc:   "Registers a Prometheus collector using MustRegister. Will panic if registration fails.",
		Fn:    __prometheus_must_register,
	},
	"handler": {
		Argsn: 0,
		Doc:   "Returns a Prometheus HTTP handler for custom HTTP server setups.",
		Fn:    __prometheus_handler,
	},
}
