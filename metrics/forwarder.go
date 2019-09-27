package metrics

import (
	loggregator "code.cloudfoundry.org/go-loggregator"
)

const (
	CPUUnit    = "percentage"
	MemoryUnit = "bytes"
	DiskUnit   = "bytes"

	CPU         = "cpu"
	Memory      = "memory"
	Disk        = "disk"
	MemoryQuota = "memory_quota"
	DiskQuota   = "disk_quota"
)

//go:generate counterfeiter . LoggregatorClient
type LoggregatorClient interface {
	EmitGauge(...loggregator.EmitGaugeOption)
}

type LoggregatorForwarder struct {
	client LoggregatorClient
}

func NewLoggregatorForwarder(client LoggregatorClient) *LoggregatorForwarder {
	return &LoggregatorForwarder{
		client: client,
	}
}

func (l *LoggregatorForwarder) Forward(msg Message) {
	gaugeOptions := []loggregator.EmitGaugeOption{loggregator.WithGaugeSourceInfo(msg.AppID, msg.IndexID)}
	for metricName, metricValue := range msg.Metrics {
		gaugeOptions = append(gaugeOptions, loggregator.WithGaugeValue(metricName, metricValue.Magnitude, metricValue.Unit))
	}

	l.client.EmitGauge(gaugeOptions...)
}
