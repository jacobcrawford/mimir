// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/cortexproject/cortex/blob/master/pkg/compactor/syncer_metrics_test.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Cortex Authors.

package compactor

import (
	"bytes"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"

	"github.com/grafana/mimir/pkg/util/test"
)

func TestSyncerMetrics(t *testing.T) {
	logger := test.NewTestingLogger(t)
	reg := prometheus.NewPedanticRegistry()

	sm := newAggregatedSyncerMetrics(reg)
	sm.gatherThanosSyncerMetrics(generateTestData(12345), logger)
	sm.gatherThanosSyncerMetrics(generateTestData(76543), logger)
	sm.gatherThanosSyncerMetrics(generateTestData(22222), logger)
	// total base = 111110

	err := testutil.GatherAndCompare(reg, bytes.NewBufferString(`
			# HELP cortex_compactor_meta_blocks_synced Number of block metadata synced
			# TYPE cortex_compactor_meta_blocks_synced gauge
			cortex_compactor_meta_blocks_synced{state="loaded"} 30

			# HELP cortex_compactor_meta_syncs_total Total blocks metadata synchronization attempts.
			# TYPE cortex_compactor_meta_syncs_total counter
			cortex_compactor_meta_syncs_total 111110

			# HELP cortex_compactor_meta_sync_failures_total Total blocks metadata synchronization failures.
			# TYPE cortex_compactor_meta_sync_failures_total counter
			cortex_compactor_meta_sync_failures_total 222220

			# HELP cortex_compactor_meta_sync_duration_seconds Duration of the blocks metadata synchronization in seconds.
			# TYPE cortex_compactor_meta_sync_duration_seconds histogram
			# Observed values: 3.7035, 22.9629, 6.6666 (seconds)
			cortex_compactor_meta_sync_duration_seconds_bucket{le="0.01"} 0
			cortex_compactor_meta_sync_duration_seconds_bucket{le="0.1"} 0
			cortex_compactor_meta_sync_duration_seconds_bucket{le="0.3"} 0
			cortex_compactor_meta_sync_duration_seconds_bucket{le="0.6"} 0
			cortex_compactor_meta_sync_duration_seconds_bucket{le="1"} 0
			cortex_compactor_meta_sync_duration_seconds_bucket{le="3"} 0
			cortex_compactor_meta_sync_duration_seconds_bucket{le="6"} 1
			cortex_compactor_meta_sync_duration_seconds_bucket{le="9"} 2
			cortex_compactor_meta_sync_duration_seconds_bucket{le="20"} 2
			cortex_compactor_meta_sync_duration_seconds_bucket{le="30"} 3
			cortex_compactor_meta_sync_duration_seconds_bucket{le="60"} 3
			cortex_compactor_meta_sync_duration_seconds_bucket{le="90"} 3
			cortex_compactor_meta_sync_duration_seconds_bucket{le="120"} 3
			cortex_compactor_meta_sync_duration_seconds_bucket{le="240"} 3
			cortex_compactor_meta_sync_duration_seconds_bucket{le="360"} 3
			cortex_compactor_meta_sync_duration_seconds_bucket{le="720"} 3
			cortex_compactor_meta_sync_duration_seconds_bucket{le="+Inf"} 3
			# rounding error
			cortex_compactor_meta_sync_duration_seconds_sum 33.333000000000006
			cortex_compactor_meta_sync_duration_seconds_count 3

			# HELP cortex_compactor_garbage_collection_total Total number of garbage collection operations.
			# TYPE cortex_compactor_garbage_collection_total counter
			cortex_compactor_garbage_collection_total 555550

			# HELP cortex_compactor_garbage_collection_failures_total Total number of failed garbage collection operations.
			# TYPE cortex_compactor_garbage_collection_failures_total counter
			cortex_compactor_garbage_collection_failures_total 666660

			# HELP cortex_compactor_garbage_collection_duration_seconds Time it took to perform garbage collection iteration.
			# TYPE cortex_compactor_garbage_collection_duration_seconds histogram
			# Observed values: 8.6415, 53.5801, 15.5554
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="0.01"} 0
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="0.1"} 0
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="0.3"} 0
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="0.6"} 0
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="1"} 0
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="3"} 0
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="6"} 0
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="9"} 1
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="20"} 2
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="30"} 2
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="60"} 3
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="90"} 3
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="120"} 3
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="240"} 3
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="360"} 3
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="720"} 3
			cortex_compactor_garbage_collection_duration_seconds_bucket{le="+Inf"} 3
			cortex_compactor_garbage_collection_duration_seconds_sum 77.777
			cortex_compactor_garbage_collection_duration_seconds_count 3
	`))
	require.NoError(t, err)
}

func generateTestData(base float64) *prometheus.Registry {
	r := prometheus.NewRegistry()
	m := newTestSyncerMetrics(r)
	m.metaSync.Add(1 * base)
	m.metaSyncFailures.Add(2 * base)
	m.metaSyncDuration.Observe(3 * base / 10000)
	m.metaBlocksSynced.WithLabelValues("loaded").Add(10)
	m.garbageCollections.Add(5 * base)
	m.garbageCollectionFailures.Add(6 * base)
	m.garbageCollectionDuration.Observe(7 * base / 10000)
	return r
}

// directly copied from Thanos (and renamed syncerMetrics to testSyncerMetrics to avoid conflict)
type testSyncerMetrics struct {
	metaSync                  prometheus.Counter
	metaSyncFailures          prometheus.Counter
	metaSyncDuration          prometheus.Histogram
	metaBlocksSynced          *prometheus.GaugeVec
	garbageCollections        prometheus.Counter
	garbageCollectionFailures prometheus.Counter
	garbageCollectionDuration prometheus.Histogram
}

func newTestSyncerMetrics(reg prometheus.Registerer) *testSyncerMetrics {
	var m testSyncerMetrics

	m.metaSync = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "blocks_meta_syncs_total",
		Help: "Total blocks metadata synchronization attempts.",
	})
	m.metaSyncFailures = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "blocks_meta_sync_failures_total",
		Help: "Total blocks metadata synchronization failures.",
	})
	m.metaSyncDuration = promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
		Name:    "blocks_meta_sync_duration_seconds",
		Help:    "Duration of the blocks metadata synchronization in seconds.",
		Buckets: []float64{0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120, 240, 360, 720},
	})

	m.metaBlocksSynced = promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
		Name: "blocks_meta_synced",
		Help: "Number of block metadata synced",
	}, []string{"state"})

	m.garbageCollections = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "thanos_compact_garbage_collection_total",
		Help: "Total number of garbage collection operations.",
	})
	m.garbageCollectionFailures = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "thanos_compact_garbage_collection_failures_total",
		Help: "Total number of failed garbage collection operations.",
	})
	m.garbageCollectionDuration = promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
		Name:    "thanos_compact_garbage_collection_duration_seconds",
		Help:    "Time it took to perform garbage collection iteration.",
		Buckets: []float64{0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120, 240, 360, 720},
	})

	return &m
}
