package k8s

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"confirmate.io/collectors/cloud/internal/logconfig"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var log *slog.Logger

func init() {
	log = logconfig.GetLogger().With("component", "k8s-collector")
}

type k8sCollector struct {
	intf kubernetes.Interface
	ctID string
}

func (d *k8sCollector) TargetOfEvaluationID() string {
	return d.ctID
}

func AuthFromKubeConfig() (intf kubernetes.Interface, err error) {
	var kubeconfig string

	// TODO(oxisto): this crashes if called twice
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		return nil, errors.New("could not find kubeconfig")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("could not read kubeconfig: %w", err)
	}

	// create the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return client, nil
}
