package chaosmgr

import (
	"errors"
	"fmt"
	"sync"
	"time"

	m "github.com/t-bfame/diago/pkg/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

// ChaosManager Responsible for simulating Choas by interacting with the Kubernetes API
type ChaosManager struct {
	// The kubernetes client
	clientset *kubernetes.Clientset

	// A map from podName to boolean to keep track of all pods under simulation
	podMap map[string]bool
	// A mutex to protect the podMap
	pmux   sync.Mutex

	// map from test template Id to stop channel
	chMap map[string] chan error
}

// Returns a string of comma separated label queries 
// for labels in the provided map
func getLabelString(lm map[string]string) (labels string) {
	for k, v := range lm {
		labels = labels + fmt.Sprintf("%s=%s,", k, v)
	}

	// Removes last ',' from labels string
	labels = labels[:len(labels)-1]

	return labels
}

// Uses a ChaosInstance's Namespace and Selector fields to find
// pods in the running kubernetes cluster
// Returns a list of strings of pod names
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

// Deletes a pod with the provided name and namespace in the kubernetes cluster
// Deletes the pod after timeout amount of time in seconds
// If podCh receives a message, then deletion operation is interrupted and stopped
// Calls Done() on the waitGroup at the end to indicate operation completion
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

// Simulate simulates chaos based on the provided ChaosInstance
// testId is the test template ID for which chaos is being simulated
// testDuration is the duration of the entire test's load
// Returns a chan used to receive errors, 
// an array of pods being deleted as a part of the chaos
// OR an error indicating if there was an error starting the simulation
func (cm *ChaosManager) Simulate(testId m.TestID, instance *m.ChaosInstance, testDuration uint64) (chan error, []string, error) {
	// Fetch names of pods that we can simulate disaster for
	p, err := cm.relevantPodNames(instance)

	if err != nil {
		return nil, nil, err
	// If no pods were found, then maybe disaster cannot be simulated
	} else if len(p) == 0 {
		return nil, nil, errors.New("No pods found for simulating disaster, recheck pod label selectors")
	}

	// Check if pod is in a test, else register it
	cm.pmux.Lock()
	defer cm.pmux.Unlock()

	// Ensure that pods that were picked are not a past of a simulation already
	var allowedPodNames []string
	for _, pod := range p {
		if _, ok := cm.podMap[fmt.Sprintf("%s-%s", pod, instance.Namespace)]; !ok {
			allowedPodNames = append(allowedPodNames, pod)
		}
	}

	// If no pods remain to pick from then return an error
	if len(allowedPodNames) == 0 {
		return nil, nil, errors.New("All pods are currently occupied in other tests, disaster will not be simulated")
	}

	// If duration of test is less than the pod death timeout
	if testDuration <= instance.Timeout {
		return nil, nil, errors.New("Pod time of death is after end of test, disaster will not be simulated")
	}

	// If number of pods required for simulation is less than available pods
	if instance.Count > len(allowedPodNames) {
		return nil, nil, errors.New("Not enough pods avaialble for deletion, disaster will not be simulated")
	}

	// Create a wait group to keep track of each pods deletion operation
	var wg sync.WaitGroup
	wg.Add(instance.Count)

	// Channel to communicate error messages back to the client
	funnelCh := make(chan error)
	var deletedPodNames []string

	for i := 0; i < instance.Count; i++ {
		podName := allowedPodNames[i]
		deletedPodNames = append(deletedPodNames, podName)
		cm.podMap[fmt.Sprintf("%s-%s", podName, instance.Namespace)] = true

		// spin up a groutine to delete pod
		go cm.deletePod(podName, instance.Namespace, instance.Timeout, &wg, funnelCh)
	}

	// another goroutine to close the returned error channel once simulation is complete
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

	// map from test template id to error channel
	cm.chMap[fmt.Sprintf("%s-%s", testId, instance.ID)] = funnelCh

	log.WithField("pods", allowedPodNames).Info("Pods selected for deletion during chaos")
	return funnelCh, deletedPodNames, nil
}

// For the provided test template ID and chaosID
// stop the load test simulation by closing the rror channel
func (cm *ChaosManager) Stop(testId m.TestID, instanceID m.ChaosID) {
	funnelCh := cm.chMap[fmt.Sprintf("%s-%s", testId, instanceID)]

	// remove the channel from map 
	delete(cm.chMap, fmt.Sprintf("%s-%s", testId, instanceID))

	close(funnelCh)
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
	cm.chMap = make(map[string]chan error)

	if err != nil {
		panic(err.Error())
	}

	return cm
}
