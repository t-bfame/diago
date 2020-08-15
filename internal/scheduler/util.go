package scheduler

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createContainerSpec(name string, image string, env map[string]string) (containers []v1.Container) {
	envVars := []v1.EnvVar{}

	for envName, envVal := range env {
		envVars = append(envVars, v1.EnvVar{
			Name:  envName,
			Value: envVal,
		})
	}

	container := v1.Container{
		Name:  name,
		Image: image,
		Env:   envVars,
	}

	return []v1.Container{container}
}

func getEnvs(group string, instanceCount int) map[string]string {
	// TODO: Use configs to get the correct IP addresses for master
	envs := map[string]string{
		"DIAGO_IMAGE_GROUP":    group,
		"DIAGO_IMAGE_GROUP_ID": string(instanceCount),
	}

	return envs
}

func getLabels(group string, instanceCount int) map[string]string {
	labels := map[string]string{
		"group":    group,
		"instance": string(instanceCount),
	}

	return labels
}

func getConfigs(group string, instanceCount int) (image string, env map[string]string, labels map[string]string) {
	image = "hello-world"

	return image, getEnvs(group, instanceCount), getLabels(group, instanceCount)
}

func createPodConfig(group string, instanceCount int) (podConfig *v1.Pod, err error) {
	// TODO: Talk to storage to get configs
	name := group + "-" + string(instanceCount)
	image, env, labels := getConfigs(group, instanceCount)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},

		Spec: v1.PodSpec{
			Containers:    createContainerSpec(name, image, env),
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	return pod, nil
}
