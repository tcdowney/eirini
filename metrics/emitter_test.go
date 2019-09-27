package metrics_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/eirini/metrics"
	"code.cloudfoundry.org/eirini/metrics/metricsfakes"
	"code.cloudfoundry.org/eirini/util/utilfakes"
)

var _ = Describe("emitter", func() {

	var (
		emitter   *Emitter
		work      chan []Message
		scheduler *utilfakes.FakeTaskScheduler
		forwarder *metricsfakes.FakeForwarder
		err       error
	)

	BeforeEach(func() {
		work = make(chan []Message, 5)
		scheduler = new(utilfakes.FakeTaskScheduler)
		forwarder = new(metricsfakes.FakeForwarder)
		emitter = NewEmitter(work, scheduler, forwarder)
	})

	Context("when metrics are sent to the channel", func() {

		var (
			message1 Message
			message2 Message
		)

		BeforeEach(func() {
			emitter.Start()

			message1 = Message{
				AppID:   "appid",
				IndexID: "0",
				Metrics: map[string]Measurement{
					CPU:         Measurement{Magnitude: 123.4},
					Memory:      Measurement{Magnitude: 123.4},
					MemoryQuota: Measurement{Magnitude: 1000.4},
					Disk:        Measurement{Magnitude: 10.1},
					DiskQuota:   Measurement{Magnitude: 250.5},
				},
			}
			message2 = Message{
				AppID:   "appid",
				IndexID: "1",
				Metrics: map[string]Measurement{
					CPU:         Measurement{Magnitude: 234.1},
					Memory:      Measurement{Magnitude: 675.4},
					MemoryQuota: Measurement{Magnitude: 1000.4},
					Disk:        Measurement{Magnitude: 10.1},
					DiskQuota:   Measurement{Magnitude: 250.5},
				},
			}

			work <- []Message{message1, message2}
		})

		JustBeforeEach(func() {
			task := scheduler.ScheduleArgsForCall(0)
			err = task()
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should call the forwarder", func() {
			callCount := forwarder.ForwardCallCount()
			Expect(callCount).To(Equal(2))
		})

		It("should forward the first message", func() {
			message := forwarder.ForwardArgsForCall(0)
			Expect(message).To(Equal(message1))
		})

		It("should forward the second message", func() {
			message := forwarder.ForwardArgsForCall(1)
			Expect(message).To(Equal(message2))
		})
	})
})
