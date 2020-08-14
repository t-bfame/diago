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

func createPodConfig(name string, image string, env map[string]string, labels map[string]string) (podConfig *v1.Pod, err error) {
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
