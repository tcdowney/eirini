package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/eirini/bifrost"
	"code.cloudfoundry.org/eirini/cmd"
	typesv1 "code.cloudfoundry.org/eirini/crd/eirini/v1"
	"code.cloudfoundry.org/eirini/handler"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
	eiriniv1 "code.cloudfoundry.org/eirini/pkg/client/clientset/versioned/typed/eirini/v1"
	"code.cloudfoundry.org/lager"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var crdCmd = &cobra.Command{
	Use:   "crd",
	Short: "Simulate pre-defined state of the Kubernetes cluster.",
	Run:   runCrd,
}

func initCrd() {
	crdCmd.Flags().StringP("config", "c", "", "Path to the Eirini config file")
}

func runCrd(command *cobra.Command, _ []string) {
	path, err := command.Flags().GetString("config")
	cmd.ExitWithError(err)

	if path == "" {
		cmd.ExitWithError(errors.New("--config is missing"))
	}

	cfg := setConfigFromFile(path)
	handlerLogger := lager.NewLogger("handler-crd-logger")
	handlerLogger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	syncLogger := lager.NewLogger("sync-crd-logger")
	syncLogger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	client := cmd.CreateEiriniClient("/Users/eirini/.kube/config")
	registryIP := cfg.Properties.RegistryAddress
	converter := bifrost.NewConverter(syncLogger, registryIP)

	bifrost := &bifrost.Bifrost{
		Converter: converter,
		Desirer: &Desirer{
			Client:    client.EiriniV1(),
			Namespace: "default",
		},
	}

	stager := &Stagercrd{}
	handler := handler.New(bifrost, stager, handlerLogger)

	log.Fatal(http.ListenAndServe("127.0.0.1:8085", handler))
}

type Desirer struct {
	Client    eiriniv1.EiriniV1Interface
	Namespace string
}

type LRP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	LRP               opi.LRP `json:"lrp"`
}

type LRPList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of pod metrics.
	Items []opi.LRP `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (l *LRPList) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

func (l *LRPList) DeepCopyObject() runtime.Object {
	return l
}

// func (obj *LRPList) GetObjectKind() schema.ObjectKind { return obj }
// func (obj *LRPList) SetGroupVersionKind(gvk schema.GroupVersionKind) {
// 	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
// }
// func (obj *LRPList) GroupVersionKind() schema.GroupVersionKind {
// 	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
// }

func (d *Desirer) Desire(lrp *opi.LRP) error {
	fmt.Printf("%#v\n", lrp)
	l := typesv1.LRP{
		LRP: *lrp,
		ObjectMeta: metav1.ObjectMeta{
			Name: lrp.ProcessGUID(),
		},
	}
	_, err := d.Client.LRPs(d.Namespace).Create(&l)
	fmt.Printf("%#v\n", l)
	return err
}

func (d *Desirer) List() ([]*opi.LRP, error) {
	return []*opi.LRP{}, nil
}

func (d *Desirer) Get(identifier opi.LRPIdentifier) (*opi.LRP, error) {
	return nil, nil
}

func (d *Desirer) GetInstances(identifier opi.LRPIdentifier) ([]*opi.Instance, error) {
	return []*opi.Instance{}, nil
}

func (d *Desirer) Update(lrp *opi.LRP) error {
	return nil
}

func (d *Desirer) Stop(identifier opi.LRPIdentifier) error {
	return nil
}

func (d *Desirer) StopInstance(identifier opi.LRPIdentifier, index uint) error {
	return nil
}

type Convertercrd struct{}

func (c *Convertercrd) Convert(request cf.DesireLRPRequest) (opi.LRP, error) {
	return opi.LRP{}, nil
}

type Stagercrd struct{}

func (s *Stagercrd) Stage(stagingGUID string, request cf.StagingRequest) error {
	return nil
}

func (s *Stagercrd) CompleteStaging(task *models.TaskCallbackResponse) error {
	return nil
}
