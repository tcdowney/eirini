package k8s

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate counterfeiter . EventLister
type EventLister interface {
	List(opts meta.ListOptions) (*v1.EventList, error)
}

func MapToEnvVar(env map[string]string) []v1.EnvVar {
	envVars := []v1.EnvVar{}
	for k, v := range env {
		envVar := v1.EnvVar{Name: k, Value: v}
		envVars = append(envVars, envVar)
	}
	return envVars
}

func int32ptr(i int) *int32 {
	u := int32(i)
	return &u
}

func int64ptr(i int) *int64 {
	u := int64(i)
	return &u
}

func GetEvents(client EventLister, pod v1.Pod) (*v1.EventList, error) {
	return client.List(meta.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.namespace=%s,involvedObject.uid=%s,involvedObject.name=%s", pod.Namespace, string(pod.UID), pod.Name)})
}

func IsStopped(eventList *v1.EventList) bool {
	events := eventList.Items

	if len(events) == 0 {
		return false
	}

	event := events[len(events)-1]
	return event.Reason == eventKilling
}
