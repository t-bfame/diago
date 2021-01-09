package scheduler

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/t-bfame/diago/api/v1alpha1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

type SchedulerModel struct {
	client *v1alpha1.DiagoV1Alpha1Client
}

func (sm SchedulerModel) createContainerSpec(name string, image string, env map[string]string) (containers []v1.Container) {
	envVars := []v1.EnvVar{}

	for envName, envVal := range env {
		envVars = append(envVars, v1.EnvVar{
			Name:  envName,
			Value: envVal,
		})
	}

	container := v1.Container{
		Name:            name,
		Image:           image,
		Env:             envVars,
		ImagePullPolicy: v1.PullNever,
	}

	return []v1.Container{container}
}

func (sm SchedulerModel) getEnvs(group string, instance InstanceID) map[string]string {
	workerConfig, err := sm.client.WorkerGroups("default").Get(group)

	if err != nil {
		log.WithField("group", group).Error("Unable to find config for worker")
		return nil
	}

	envs := map[string]string{
		"DIAGO_WORKER_GROUP":                group,
		"DIAGO_WORKER_GROUP_INSTANCE":       string(instance),
		"DIAGO_LEADER_HOST":                 os.Getenv("GRPC_HOST"),
		"DIAGO_LEADER_PORT":                 os.Getenv("GRPC_PORT"),
		"ALLOWED_INACTIVITY_PERIOD_SECONDS": fmt.Sprintf("%d", workerConfig.Spec.AllowedInactivityPeriod),
	}

	return envs
}

func (sm SchedulerModel) getLabels(group string, instance InstanceID) map[string]string {
	labels := map[string]string{
		"group":    group,
		"instance": string(instance),
	}

	return labels
}

func (sm SchedulerModel) getConfigs(group string, instance InstanceID) (image string, env map[string]string, labels map[string]string, err error) {
	workerConfig, err := sm.client.WorkerGroups("default").Get(group)

	if err != nil {
		log.WithField("group", group).Error("Unable to find config for worker")
		return "", nil, nil, err
	}

	image = workerConfig.Spec.Image
	return image, sm.getEnvs(group, instance), sm.getLabels(group, instance), nil
}

func (sm SchedulerModel) createPodConfig(group string, instance InstanceID) (podConfig *v1.Pod, err error) {
	name := group + "-" + string(instance)
	image, env, labels, err := sm.getConfigs(group, instance)
	var gracePeriod int64 = 0

	if err != nil {
		return nil, err
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},

		Spec: v1.PodSpec{
			Containers:                    sm.createContainerSpec(name, image, env),
			RestartPolicy:                 v1.RestartPolicyNever,
			TerminationGracePeriodSeconds: &gracePeriod,
		},
	}

	return pod, nil
}

func (sm SchedulerModel) getCapacity(group string) (uint64, error) {

	workerConfig, err := sm.client.WorkerGroups("default").Get(group)

	if err != nil {
		log.WithField("group", group).Error("Unable to find capacity for worker")
		return strconv.ParseUint(os.Getenv("DEFAULT_GROUP_CAPACITY"), 10, 64)
	}

	return uint64(workerConfig.Spec.Capacity), nil
}

func NewSchedulerModel(config *rest.Config) (*SchedulerModel, error) {
	crdclient, err := v1alpha1.NewClient(config)

	if err != nil {
		return nil, errors.New("Unable to initialize custom CRD client")
	}

	return &SchedulerModel{crdclient}, nil
}
