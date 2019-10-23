package statefulsets_test

import (
	"code.cloudfoundry.org/eirini/k8s"
	informerroute "code.cloudfoundry.org/eirini/k8s/informers/route"
	"code.cloudfoundry.org/eirini/opi"
	"code.cloudfoundry.org/eirini/route"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Routes", func() {

	var (
		desirer opi.Desirer
		odinLRP *opi.LRP
	)

	AfterEach(func() {
		cleanupStatefulSet(odinLRP)
		Eventually(func() []appsv1.StatefulSet {
			return listAllStatefulSets(odinLRP, odinLRP)
		}, timeout).Should(BeEmpty())
	})

	BeforeEach(func() {
		odinLRP = createLRP("ödin")
		logger := lagertest.NewTestLogger("test")
		desirer = k8s.NewStatefulSetDesirer(
			clientset,
			namespace,
			"registry-secret",
			"rootfsversion",
			logger,
		)
	})

	Context("When creating a StatefulSet", func() {

		It("sends register routes message TODO", func() {
			err := desirer.Desire(odinLRP)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				pods := listPods(odinLRP.LRPIdentifier)
				if len(pods) < 1 {
					return false
				}
				return podReady(pods[0])
			}, timeout).Should(BeTrue())

			logger := lagertest.NewTestLogger("test")
			collector := k8s.NewRouteCollector(clientset, namespace, logger)
			routes, err := collector.Collect()
			Expect(err).ToNot(HaveOccurred())
			pods := listPods(odinLRP.LRPIdentifier)
			Expect(routes).To(ContainElement(route.Message{
				InstanceID: pods[0].Name,
				Name:       odinLRP.GUID,
				Address:    pods[0].Status.PodIP,
				Port:       8080,
				TLSPort:    0,
				Routes: route.Routes{
					RegisteredRoutes: []string{"foo.example.com"},
				},
			}))
		})

		Context("When one of the instances if failing", func() {
			BeforeEach(func() {
				odinLRP = createLRP("odin")
				odinLRP.Health = opi.Healtcheck{
					Type: "port",
					Port: 3000,
				}
				odinLRP.Command = []string{
					"/bin/sh",
					"-c",
					`if [ $(echo $HOSTNAME | sed 's|.*-\(.*\)|\1|') -eq 1 ]; then
	exit;
else
	while true; do
		nc -lk -p 3000 -e echo just a server;
	done;
fi;`,
				}
			})
		})
	})

	Context("When deleting a LRP", func() {

		var (
			pods     []corev1.Pod
			workChan chan *route.Message
			stopChan chan struct{}
		)

		BeforeEach(func() {
			odinLRP.TargetInstances = 2
			err := desirer.Desire(odinLRP)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				pods := listPods(odinLRP.LRPIdentifier)
				if len(pods) < 2 {
					return false
				}
				return podReady(pods[0]) && podReady(pods[1])
			}, timeout).Should(BeTrue())

			stopChan = make(chan struct{})
			workChan = make(chan *route.Message, 5)

			logger := lagertest.NewTestLogger("instance-informer-test")

			informer := &informerroute.InstanceChangeInformer{
				Client:    clientset,
				Cancel:    stopChan,
				Namespace: namespace,
				Logger:    logger,
			}

			go informer.Start(workChan)
		})

		AfterEach(func() {
			close(stopChan)
		})

		Context("when the app is scaled down", func() {
			It("sends unregister routes message", func() {
				pods = listPods(odinLRP.LRPIdentifier)
				odinLRP.TargetInstances = 1
				err := desirer.Update(odinLRP)

				Expect(err).ToNot(HaveOccurred())
				Eventually(workChan, timeout).Should(Receive(Equal(&route.Message{
					Routes: route.Routes{
						UnregisteredRoutes: []string{"foo.example.com"},
					},
					InstanceID: pods[1].Name,
					Name:       odinLRP.GUID,
					Address:    pods[1].Status.PodIP,
					Port:       8080,
					TLSPort:    0,
				})))
			})
		})

		Context("when an app instance is stopped", func() {
			It("sends unregister routes message", func() {
				pods = listPods(odinLRP.LRPIdentifier)
				err := desirer.StopInstance(odinLRP.LRPIdentifier, 0)

				Expect(err).ToNot(HaveOccurred())
				Eventually(workChan, timeout).Should(Receive(Equal(&route.Message{
					Routes: route.Routes{
						UnregisteredRoutes: []string{"foo.example.com"},
					},
					InstanceID: pods[0].Name,
					Name:       odinLRP.GUID,
					Address:    pods[0].Status.PodIP,
					Port:       8080,
					TLSPort:    0,
				})))
			})
		})
	})

})

func podReady(pod corev1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}
