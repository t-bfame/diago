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

func getEnvs(group string, instance InstanceID) map[string]string {
	// TODO: Use configs to get the correct IP addresses for master
	envs := map[string]string{
		"DIAGO_IMAGE_GROUP":    group,
		"DIAGO_IMAGE_GROUP_ID": string(instance),
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

func getConfigs(group string, instance InstanceID) (image string, env map[string]string, labels map[string]string) {
	image = "hello-world"

	return image, getEnvs(group, instance), getLabels(group, instance)
}

func createPodConfig(group string, instance InstanceID) (podConfig *v1.Pod, err error) {
	// TODO: Talk to storage to get configs
	name := group + "-" + string(instance)
	image, env, labels := getConfigs(group, instance)

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
