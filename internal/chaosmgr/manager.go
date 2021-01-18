package chaosmgr

import (
	"fmt"
	"sync"

	m "github.com/t-bfame/diago/internal/model"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

type ChaosManager struct {
	clientset *kubernetes.Clientset

	podMap map[string]*sync.Mutex
	pmux   sync.Mutex
}

func getLabelString(lm map[string]string) (labels string) {
	for k, v := range lm {
		labels = labels + fmt.Sprintf("%s=%s,", k, v)
	}

	// Removes last ',' from labels string
	labels = labels[:len(labels)-1]

	return labels
}

func (cm *ChaosManager) relevantPodNames(instance *m.ChaosInstance) []string {
	pods, err := cm.clientset.CoreV1().Pods(instance.Namespace).List(meta_v1.ListOptions{
		LabelSelector: getLabelString(instance.Selectors),
	})

	if err != nil {
		log.Error(err)
	}

	var podNames []string

	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	return podNames
}

func (cm *ChaosManager) Simulate(instance *m.ChaosInstance) {
	cm.pmux.Lock()

	p := cm.relevantPodNames(instance)
	log.WithField("Podname", p).Info("Result")
}

// NewChaosManager laalala
func NewChaosManager() *ChaosManager {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	cm := new(ChaosManager)
	cm.clientset = clientset

	if err != nil {
		panic(err.Error())
	}

	return cm
}
