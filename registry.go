package gangliamr

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/daaku/go.ganglia/gmetric"
)

type metric interface {
	writeMeta(c *gmetric.Client)
	writeValue(c *gmetric.Client)
	register(r *Registry)
}

// Registry provides a metrics Registry that will report the registered metrics
// to Ganglia based on the configuration.
type Registry struct {
	Prefix        string
	NameSeparator string // Defaults to a dot "."
	Client        *gmetric.Client
	TickInterval  time.Duration
	startOnce     sync.Once
	metrics       []metric
	mutex         sync.Mutex
}

func (r *Registry) start() {
	go func() {
		for {
			metrics := r.registered()
			for _, m := range metrics {
				m.writeMeta(r.Client)
				m.writeValue(r.Client)
			}
			time.Sleep(r.TickInterval)
		}
	}()
}

// Register a metric.
func (r *Registry) Register(m interface{}) {
	r.startOnce.Do(r.start)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	v, ok := m.(metric)
	if !ok {
		panic(fmt.Sprintf("unknown metric type: %T", m))
	}
	v.register(r)
	r.metrics = append(r.metrics, v)
}

func (r *Registry) registered() []metric {
	metrics := make([]metric, len(r.metrics))
	r.mutex.Lock()
	defer r.mutex.Unlock()
	copy(metrics, r.metrics)
	return metrics
}

func (r *Registry) makeName(parts ...string) string {
	var nonempty []string
	sep := r.NameSeparator
	if sep == "" {
		sep = "."
	}
	if r.Prefix != "" {
		nonempty = append(nonempty, r.Prefix)
	}
	for _, p := range parts {
		if p != "" {
			nonempty = append(nonempty, p)
		}
	}
	return strings.Join(nonempty, sep)
}