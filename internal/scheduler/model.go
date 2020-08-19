package scheduler

import (
	"errors"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodConfig TODO: Move out to Storage later
type PodConfig struct {
	Image                         string
	Capacity                      uint64
	TerminationGracePeriodSeconds float32
}

var storage map[string]PodConfig = map[string]PodConfig{
	"diago-worker": PodConfig{Image: "diago-worker", Capacity: 5, TerminationGracePeriodSeconds: 30},
}

func createContainerSpec(name string, image string, env map[string]string) (containers []v1.Container) {
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

func getEnvs(group string, instance InstanceID) map[string]string {
	dat, ok := storage[group]
	if !ok {
		return nil
	}

	envs := map[string]string{
		"DIAGO_WORKER_GROUP":               group,
		"DIAGO_WORKER_GROUP_INSTANCE":      string(instance),
		"DIAGO_LEADER_HOST":                os.Getenv("GRPC_HOST"),
		"DIAGO_LEADER_PORT":                os.Getenv("GRPC_PORT"),
		"TERMINATION_GRACE_PERIOD_SECONDS": fmt.Sprintf("%f", dat.TerminationGracePeriodSeconds),
	}

	return envs
}

func getLabels(group string, instance InstanceID) map[string]string {
	labels := map[string]string{
		"group":    group,
		"instance": string(instance),
	}

	return labels
}

func getConfigs(group string, instance InstanceID) (image string, env map[string]string, labels map[string]string, err error) {
	dat, ok := storage[group]
	if !ok {
		return "", nil, nil, errors.New("Could not find image for specified group")
	}

	image = dat.Image
	return image, getEnvs(group, instance), getLabels(group, instance), nil
}

func createPodConfig(group string, instance InstanceID) (podConfig *v1.Pod, err error) {
	// TODO: Talk to storage to get configs
	name := group + "-" + string(instance)
	image, env, labels, err := getConfigs(group, instance)
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
			Containers:                    createContainerSpec(name, image, env),
			RestartPolicy:                 v1.RestartPolicyNever,
			TerminationGracePeriodSeconds: &gracePeriod,
		},
	}

	return pod, nil
}

func getCapacity(group string) (uint64, error) {
	dat, ok := storage[group]
	if !ok {
		return 0, errors.New("Could not find capacity for specified group")
	}

	return dat.Capacity, nil
}
