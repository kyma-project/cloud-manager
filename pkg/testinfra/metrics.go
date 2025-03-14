package testinfra

import (
	"fmt"
	"github.com/elliotchance/pie/v2"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

func PrintMetrics() error {
	fmt.Println("")
	fmt.Println("Controller metrics")
	fmt.Println("")

	families, err := metrics.Registry.Gather()
	if err != nil {
		return err
	}

	for _, family := range families {
		if ptr.Deref(family.Name, "") == "controller_runtime_reconcile_total" {
			for _, metric := range family.Metric {
				val := ptr.Deref(metric.Counter.Value, float64(0))
				if val == 0 {
					continue
				}
				labels := pie.Map(metric.Label, func(x *io_prometheus_client.LabelPair) string {
					return fmt.Sprintf("%s: %s", ptr.Deref(x.Name, ""), ptr.Deref(x.Value, ""))
				})
				fmt.Printf("%v: %v\n", labels, val)
			}
		}
	}

	fmt.Println("")

	return nil
}
