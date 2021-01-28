package chaosmgr

import (
	"errors"
	"fmt"
	"sync"
	"time"

	m "github.com/t-bfame/diago/internal/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

// ChaosManager controls chaos
type ChaosManager struct {
	clientset *kubernetes.Clientset

	podMap map[string]bool
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

func (cm *ChaosManager) relevantPodNames(instance *m.ChaosInstance) ([]string, error) {
	pods, err := cm.clientset.CoreV1().Pods(instance.Namespace).List(metav1.ListOptions{
		LabelSelector: getLabelString(instance.Selectors),
	})

	if err != nil {
		log.WithError(err).Error("Unable to get pods from namespace")
		return nil, err
	}

	var podNames []string

	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	return podNames, nil
}

func (cm *ChaosManager) deletePod(name string, namespace string, timeout uint64, wg *sync.WaitGroup, podCh chan error) {
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		deletePolicy := metav1.DeletePropagationForeground

		if err := cm.clientset.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}); err != nil {
			log.WithError(err).WithField("podName", name).WithField("namespace", namespace).Error("Encountered error while pod deletion")
		}

		log.WithField("podName", name).WithField("namespace", namespace).Info("Removed pod")

	case <-podCh:
		log.WithField("name", name).WithField("namespace", namespace).Info("Received chaos interrupt, terminating simulation")
	}

	cm.pmux.Lock()
	defer cm.pmux.Unlock()

	delete(cm.podMap, fmt.Sprintf("%s-%s", name, namespace))
	wg.Done()
}

// Simulate simulates disaster!!
func (cm *ChaosManager) Simulate(instance *m.ChaosInstance) (chan error, error) {
	// Fetch names of pods that we can simulate disaster for
	p, err := cm.relevantPodNames(instance)

	if err != nil {
		return nil, err
	} else if len(p) == 0 {
		return nil, errors.New("No pods found for simulating disaster, recheck pod label selectors")
	}

	// Check if pod is in a test, else register it
	cm.pmux.Lock()
	defer cm.pmux.Unlock()

	var allowedPodNames []string
	for _, pod := range p {
		if _, ok := cm.podMap[fmt.Sprintf("%s-%s", pod, instance.Namespace)]; !ok {
			allowedPodNames = append(allowedPodNames, pod)
		}
	}

	if len(allowedPodNames) == 0 {
		return nil, errors.New("All pods are currently occupied in other tests, disaster will not be simulated")
	}

	// If duration of test is less than the pod death timeout
	if instance.Duration <= instance.Timeout {
		return nil, errors.New("Pod time of death is after end of test, disaster will not be simulated")
	}

	if instance.Count > len(allowedPodNames) {
		return nil, errors.New("Not enough pods avaialble for deletion, disaster will not be simulated")
	}

	var wg sync.WaitGroup
	wg.Add(instance.Count)

	funnelCh := make(chan error)

	for i := 0; i < instance.Count; i++ {
		podName := allowedPodNames[i]
		cm.podMap[fmt.Sprintf("%s-%s", podName, instance.Namespace)] = true

		go cm.deletePod(podName, instance.Namespace, instance.Timeout, &wg, funnelCh)
	}

	go func() {
		wg.Wait()

		select {
		case _, ok := <-funnelCh:
			if !ok {
				log.Debug("Funnel channel was already closed i.e. simulation was terminated")
			}
		default:
			close(funnelCh)
		}
	}()

	log.WithField("pods", allowedPodNames).Info("Pods selected for deletion during chaos")
	return funnelCh, nil
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
	cm.podMap = make(map[string]bool)

	if err != nil {
		panic(err.Error())
	}

	return cm
}
