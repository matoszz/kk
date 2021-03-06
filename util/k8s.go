package util

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mateo1647/kk/internal/options"
	"github.com/mateo1647/kk/pkg/client"
)

var (
	clientset = client.InitClient()
)

// setOptions - set common options for clientset
func SetOptions(opt *options.SearchOptions) (string, *metav1.ListOptions) {
	// set default namespace as "default"
	namespace := "default"

	// override `namespace` if `--all-namespaces` exist
	if opt.AllNamespaces {
		namespace = ""
	} else {
		if len(opt.Namespace) > 0 {
			namespace = opt.Namespace
		} else {
			ns, _, err := client.ClientConfig().Namespace()
			if err != nil {
				log.WithFields(log.Fields{
					"err": err.Error(),
				}).Debug("Failed to resolve namespace")
			} else {
				namespace = ns
			}
		}
	}

	// retrieve listOptions from meta
	listOptions := &metav1.ListOptions{
		LabelSelector: opt.Selector,
		FieldSelector: opt.FieldSelector,
	}
	return namespace, listOptions
}

// DaemonsetList - return a list of DaemonSet(s)
func DaemonsetList(opt *options.SearchOptions) *appsv1.DaemonSetList {
	ns, o := SetOptions(opt)
	list, err := clientset.AppsV1().DaemonSets(ns).List(*o)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Debug("Unable to get DaemonSet List")
	}
	return list
}

// DeploymentList - return a list of Deployment(s)
func DeploymentList(opt *options.SearchOptions) *appsv1.DeploymentList {
	ns, o := SetOptions(opt)
	list, err := clientset.AppsV1().Deployments(ns).List(*o)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Debug("Unable to get Deployment List")
	}
	return list
}

// PodList - return a list of Pod(s)
func PodList(opt *options.SearchOptions) *corev1.PodList {
	ns, o := SetOptions(opt)
	list, err := clientset.CoreV1().Pods(ns).List(*o)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Debug("Unable to get Pod List")
	}
	return list
}

// NodeList - return a list of Node(s)
func NodeList(opt *options.SearchOptions) *corev1.NodeList {
	_, o := SetOptions(opt)
	list, err := clientset.CoreV1().Nodes().List(*o)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Debug("Unable to get Node List")
	}
	return list
}

// ConfigMapList - return a list of ConfigMap(s)
func ConfigMapList(opt *options.SearchOptions) *corev1.ConfigMapList {
	ns, o := SetOptions(opt)
	list, err := clientset.CoreV1().ConfigMaps(ns).List(*o)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Debug("Unable to get ConfigMap List")
	}
	return list
}

// SecretList - return a list of Secret(s)
func SecretList(opt *options.SearchOptions) *corev1.SecretList {
	ns, o := SetOptions(opt)
	list, err := clientset.CoreV1().Secrets(ns).List(*o)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Debug("Unable to get Secret List")
	}
	return list
}

// StatefulSetList - return a list of StatefulSets
func StatefulSetList(opt *options.SearchOptions) *appsv1.StatefulSetList {
	ns, o := SetOptions(opt)
	list, err := clientset.AppsV1().StatefulSets(ns).List(*o)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Debug("Unable to get .StatefulSet List")
	}
	return list
}

func ServiceList(opt *options.SearchOptions) *corev1.ServiceList {
	ns, o := SetOptions(opt)
	list, err := clientset.CoreV1().Services(ns).List(*o)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Debug("Unable to get .Services List")
	}
	return list
}

// TrimQuoteAndSpace - remove Spaces, Tabs, SingleQuotes, DoubleQuites
func TrimQuoteAndSpace(input string) string {
	if len(input) >= 2 {
		if input[0] == '"' && input[len(input)-1] == '"' {
			return input[1 : len(input)-1]
		}
		if input[0] == '\'' && input[len(input)-1] == '\'' {
			return input[1 : len(input)-1]
		}
	}
	return strings.TrimSpace(input)
}

// GetAge - return human readable time expression
func GetAge(d time.Duration) string {
	return duration.HumanDuration(d)
}

func GetDefaultNamespace() string {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	if ns, _, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).Namespace(); err == nil {
		return ns
	}
	return v1.NamespaceDefault
}
func KeysString(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k, v := range m {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(keys, ",")
}

func RunCommand(name string, args ...string) []string {
	//fmt.Printf("%v %v\n", name, args)
	cmd := exec.Command(name, args...)

	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("error running command: %v %v\n", name, args)
	}
	output := strings.Split(string(cmdOut), "\n")

	return output
}

func RawK8sOutput(namespace string, context string, labels string, args ...string) []string {
	cmdArgs := K8sCommandArgs(args, namespace, context, labels)
	output := RunCommand("kubectl", cmdArgs...)
	return output
}

func K8sCommandArgs(args []string, namespace string, context string, labels string) []string {
	if namespace != "" {
		args = append(args, fmt.Sprintf("--namespace=%v", namespace))
	}
	if context != "" {
		args = append(args, fmt.Sprintf("--context=%v", context))
	}
	if labels != "" {
		args = append(args, fmt.Sprintf("--selector=%v", labels))
	}
	return args
}
