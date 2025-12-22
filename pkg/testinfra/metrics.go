package testinfra

import (
	"bytes"
	"fmt"
	"sort"

	dto "github.com/prometheus/client_model/go"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type metricControllerNameKey struct {
	metric     string
	controller string
	name       string
}

func (k *metricControllerNameKey) Key() string {
	return fmt.Sprintf("%s/%s/%s", k.metric, k.controller, k.name)
}

// reconcileObject ============================================================

func newReconcileObject(key metricControllerNameKey) *reconcileObject {
	return &reconcileObject{
		metricControllerNameKey: key,
		values:                  map[string]float64{},
	}
}

type reconcileObject struct {
	metricControllerNameKey
	values map[string]float64
	total  float64
}

func (o *reconcileObject) Add(metric *reconcileMetricRaw) bool {
	if metric.Key() != o.Key() {
		return false
	}
	if o.values == nil {
		o.values = make(map[string]float64)
	}
	o.values[metric.result] = metric.val
	o.total += metric.val
	return true
}

func (o *reconcileObject) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s, %s, %s, %v\n", o.metric, o.controller, o.name, o.total))
	for k, v := range o.values {
		buf.WriteString(fmt.Sprintf("                                    , , %s, %v\n", k, v))
	}
	return buf.String()
}

// reconcileMetricRaw ============================================================

func newReconcileMetricRaw(family *dto.MetricFamily, metric *dto.Metric) *reconcileMetricRaw {
	key := metricControllerNameKey{
		metric: family.GetName(),
	}
	var result string
	for _, l := range metric.Label {
		switch l.GetName() {
		case "controller":
			key.controller = l.GetValue()
		case "name":
			key.name = l.GetValue()
		case "result":
			result = l.GetValue()
		}
	}
	var value float64
	if metric.Counter != nil {
		value = metric.Counter.GetValue()
	}
	return &reconcileMetricRaw{
		metricControllerNameKey: key,
		result:                  result,
		val:                     value,
	}
}

type reconcileMetricRaw struct {
	metricControllerNameKey
	result string
	val    float64
}

func (d *reconcileMetricRaw) Key() string {
	return fmt.Sprintf("%s/%s/%s", d.metric, d.controller, d.name)
}

// =======================================================

func PrintMetrics() error {
	// there's no much use of metrics as printed now
	// they should be collected after each scenario, remember first occurrence and reported how much it increases at the end
	// that would expose which resources were reconciled after the test that created them
	if true {
		return nil
	}

	fmt.Println("")
	fmt.Println("Controller metrics")
	fmt.Println("")

	families, err := metrics.Registry.Gather()
	if err != nil {
		return err
	}

	var data []*reconcileObject
	for _, family := range families {
		if family.GetName() == "cloud_manager_reconcile" {
			for _, metric := range family.Metric {
				raw := newReconcileMetricRaw(family, metric)
				var obj *reconcileObject
				for _, x := range data {
					if x.Key() == raw.Key() {
						obj = x
						break
					}
				}
				if obj == nil {
					obj = newReconcileObject(raw.metricControllerNameKey)
					data = append(data, obj)
				}
				obj.Add(raw)
			}
		}
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].total > data[j].total
	})
	for _, row := range data {
		fmt.Print(row.String())
	}

	fmt.Println("")

	return nil
}
