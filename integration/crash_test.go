package statefulsets_test

import (
	"encoding/json"
	"sync"

	"code.cloudfoundry.org/eirini/events"
	"code.cloudfoundry.org/eirini/k8s"
	"code.cloudfoundry.org/eirini/k8s/informers/event"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Crashes", func() {

	var (
		reportChan      chan events.CrashReport
		informerStopper chan struct{}
		logger          *lagertest.TestLogger
		informerWG      sync.WaitGroup

		reportGenerator event.DefaultCrashReportGenerator

		desirer opi.Desirer
		odinLRP *opi.LRP
	)

	BeforeEach(func() {
		reportGenerator = event.DefaultCrashReportGenerator{}
		reportChan = make(chan events.CrashReport)
		informerStopper = make(chan struct{})

		logger = lagertest.NewTestLogger("crash-event-logger-test")

		informerWG = sync.WaitGroup{}
		informerWG.Add(1)
		crashInformer := event.NewCrashInformer(clientset, 0, namespace, reportChan, informerStopper, logger, reportGenerator)
		go func() {
			crashInformer.Start()
			informerWG.Done()
		}()

		desirer = k8s.NewStatefulSetDesirer(
			clientset,
			namespace,
			"registry-secret",
			"rootfsversion",
			logger,
		)
	})

	AfterEach(func() {
		close(informerStopper)
		informerWG.Wait()
	})

	Context("When an app crashes", func() {

		BeforeEach(func() {
			odinLRP = createCrashingLRP("ödin")
			err := desirer.Desire(odinLRP)
			Expect(err).ToNot(HaveOccurred())
		})

		FIt("generates crash reports", func() {
			Eventually(reportChan, timeout).Should(Receive(Equal(events.CrashReport{
				ProcessGUID: odinLRP.LRPIdentifier.ProcessGUID(),
			})))
		})
	})
})

func createCrashingLRP(name string) *opi.LRP {
	guid := randomString()
	routes, err := json.Marshal([]cf.Route{{Hostname: "foo.example.com", Port: 8080}})
	Expect(err).ToNot(HaveOccurred())
	return &opi.LRP{
		Command: []string{
			"/bin/sh",
			"-c",
			"exit 1",
		},
		AppName:         name,
		SpaceName:       "space-foo",
		TargetInstances: 1,
		Image:           "alpine",
		Metadata: map[string]string{
			cf.ProcessGUID: name,
			cf.VcapAppUris: string(routes),
		},
		LRPIdentifier: opi.LRPIdentifier{GUID: guid, Version: "version_" + guid},
		LRP:           "metadata",
		DiskMB:        2047,
	}
}
