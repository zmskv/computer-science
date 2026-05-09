package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCollectorGatherIncludesCustomMetrics(t *testing.T) {
	collector := NewCollector(200)
	collector.startedAt = time.Unix(1_700_000_000, 0)
	collector.now = func() time.Time {
		return collector.startedAt.Add(3 * time.Second)
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}

	required := map[string]bool{
		"go_analyse_process_uptime_seconds":        false,
		"go_analyse_gc_configured_percent":         false,
		"go_analyse_gc_last_run_timestamp_seconds": false,
		"go_analyse_gc_last_run_age_seconds":       false,
	}

	for _, family := range families {
		if _, ok := required[family.GetName()]; ok {
			required[family.GetName()] = true
		}
	}

	for name, present := range required {
		if !present {
			t.Fatalf("Gather() missing metric family %q", name)
		}
	}
}

func TestNewRegistryIncludesGoCollectorMetrics(t *testing.T) {
	registry := NewRegistry(100)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}

	var foundGoCollector bool
	for _, family := range families {
		if family.GetName() == "go_goroutines" {
			foundGoCollector = true
			break
		}
	}

	if !foundGoCollector {
		t.Fatal("Gather() does not include built-in Go collector metrics")
	}
}

func TestSecondsSinceLastGCWithoutGC(t *testing.T) {
	if got := secondsSinceLastGC(0, time.Now()); got != 0 {
		t.Fatalf("secondsSinceLastGC() = %v, want 0", got)
	}
}
