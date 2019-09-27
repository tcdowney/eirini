package metrics_test

import (
	"time"

	"code.cloudfoundry.org/eirini/metrics"
	"code.cloudfoundry.org/eirini/metrics/metricsfakes"
	loggregator "code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Actual things, don't commit", func() {
	It("is real", func() {
		tlsConfig, err := loggregator.NewIngressTLSConfig(
			"/Users/eirini/workspace/eirini/ca.pem",
			"/Users/eirini/workspace/eirini/loggregator-cert.pem",
			"/Users/eirini/workspace/eirini/loggregator-key.pem",
		)
		Expect(err).ToNot(HaveOccurred())

		loggregatorClient, err := loggregator.NewIngressClient(
			tlsConfig,
			loggregator.WithAddr(cfg.Properties.LoggregatorAddress),
		)

	})
})

var _ = Describe("Forwarder", func() {
	It("should forward source info to loggregator", func() {
		fakeClient := new(metricsfakes.FakeLoggregatorClient)
		forwarder := metrics.NewLoggregatorForwarder(fakeClient)
		envelope := newEnvelope()

		msg := metrics.Message{
			AppID:   "amazing-app-id",
			IndexID: "best-index-id",
			Metrics: map[string]metrics.Measurement{
				metrics.CPU:    {Magnitude: 3, Unit: metrics.CPUUnit},
				metrics.Memory: {Magnitude: 300000, Unit: metrics.MemoryUnit},
			},
		}
		forwarder.Forward(msg)
		Expect(fakeClient.EmitGaugeCallCount()).To(Equal(1))

		emitGaugeOpts := fakeClient.EmitGaugeArgsForCall(0)
		for _, g := range emitGaugeOpts {
			g(envelope)
		}

		Expect(envelope.SourceId).To(Equal(msg.AppID))
		Expect(envelope.InstanceId).To(Equal(msg.IndexID))
		expectedMetrics := map[string]*loggregator_v2.GaugeValue{
			metrics.CPU: {
				Unit:  metrics.CPUUnit,
				Value: 3,
			},
			metrics.Memory: {
				Unit:  metrics.MemoryUnit,
				Value: 300000,
			},
		}

		Expect(envelope.GetGauge().Metrics).To(Equal(expectedMetrics))
	})
})

func newEnvelope() *loggregator_v2.Envelope {
	return &loggregator_v2.Envelope{
		Timestamp: time.Now().UnixNano(),
		Message: &loggregator_v2.Envelope_Gauge{
			Gauge: &loggregator_v2.Gauge{
				Metrics: make(map[string]*loggregator_v2.GaugeValue),
			},
		},
		Tags: make(map[string]string),
	}
}
