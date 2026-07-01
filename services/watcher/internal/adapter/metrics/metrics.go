package metrics

import (
	"github.com/LiquidCats/paw/watcher/internal/app/domain/entities"
	"github.com/prometheus/client_golang/prometheus"
)

type RequestsToNodeCount struct {
	internal *prometheus.CounterVec
}

func NewRequestsToNodeCount(namespace string) *RequestsToNodeCount {
	counterVec := prometheus.V2.NewCounterVec(prometheus.CounterVecOpts{
		CounterOpts: prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "node_request_count",
		},
		VariableLabels: prometheus.ConstrainedLabels{
			{
				Name:       "chain",
				Constraint: nil,
			},
		},
	})

	prometheus.MustRegister(counterVec)

	return &RequestsToNodeCount{
		internal: counterVec,
	}
}

func (r *RequestsToNodeCount) Inc(chain entities.Chain) {
	r.internal.WithLabelValues(string(chain)).Inc()
}
