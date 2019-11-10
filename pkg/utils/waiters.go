package utils

import (
	"context"
	"time"

	"github.com/dimaunx/armada/pkg/cluster"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// WaitForPodsRunning waits for pods to be running
func WaitForPodsRunning(cl *cluster.Cluster, kubeconfigFilePath, namespace, selector string, replicas int) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFilePath)
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	ctx := context.Background()
	podTimeout := 5 * time.Minute
	log.Warnf("Waiting for pods to be running. label: %q, namespace: %q, replicas: %v, duration: %v, cluster: %s.", selector, namespace, replicas, podTimeout, cl.Name)
	podsContext, cancel := context.WithTimeout(ctx, podTimeout)
	wait.Until(func() {
		podList, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{
			LabelSelector: selector,
			FieldSelector: "status.phase=Running",
		})
		if err == nil && len(podList.Items) == replicas {
			log.Infof("✔ All pods with label: %q are running for %s.", selector, cl.Name)
			cancel()
		} else {
			log.Debugf("Still waiting for pods. label: %q, namespace: %q, replicas: %v, duration: %v, cluster: %s.", selector, namespace, replicas, podTimeout, cl.Name)
		}
	}, 30*time.Second, podsContext.Done())

	err = podsContext.Err()
	if err != nil && err != context.Canceled {
		return errors.Wrap(err, "Error waiting for pods to be running.")
	}
	return nil
}

// WaitForDeployment waits for deployment roll out
func WaitForDeployment(cl *cluster.Cluster, kubeconfigFilePath, namespace, deploymentName string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFilePath)
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	ctx := context.Background()
	deploymentTimeout := 5 * time.Minute
	log.Infof("Waiting up to %v for %s deployment roll out %s...", deploymentTimeout, deploymentName, cl.Name)
	deploymentContext, cancel := context.WithTimeout(ctx, deploymentTimeout)
	wait.Until(func() {
		deployment, err := clientSet.AppsV1().Deployments(namespace).Get(deploymentName, metav1.GetOptions{})
		if err == nil && deployment.Status.ReadyReplicas > 0 {
			if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
				log.Infof("✔ %s successfully deployed to %s, ready replicas: %v", deploymentName, cl.Name, deployment.Status.ReadyReplicas)
				cancel()
			} else {
				log.Warnf("Still waiting for %s deployment %s, ready replicas: %v", deploymentName, cl.Name, deployment.Status.ReadyReplicas)
			}
		} else {
			log.Debugf("Still waiting for %s deployment roll out %s...", deploymentName, cl.Name)
		}
	}, 10*time.Second, deploymentContext.Done())
	err = deploymentContext.Err()
	if err != nil && err != context.Canceled {
		return errors.Wrapf(err, "Error waiting for %s deployment roll out.", deploymentName)
	}
	return nil
}

// WaitForDaemonSet waits for daemon set roll out
func WaitForDaemonSet(cl *cluster.Cluster, kubeconfigFilePath, namespace, daemonSetName string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFilePath)
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	nodeList, err := clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	ctx := context.Background()
	daemonSetTimeout := 5 * time.Minute
	log.Infof("Waiting up to %v for %s daemon set roll out %s...", daemonSetTimeout, daemonSetName, cl.Name)
	deploymentContext, cancel := context.WithTimeout(ctx, daemonSetTimeout)
	wait.Until(func() {
		daemonSet, err := clientSet.AppsV1().DaemonSets(namespace).Get(daemonSetName, metav1.GetOptions{})
		if err == nil && daemonSet.Status.NumberReady > 0 {
			if daemonSet.Status.NumberReady == int32(len(nodeList.Items)) {
				log.Infof("✔ %s successfully rolled out to %s, ready replicas: %v", daemonSetName, cl.Name, daemonSet.Status.NumberReady)
				cancel()
			} else {
				log.Warnf("Still waiting for %s daemon set %s roll out, ready replicas: %v", daemonSetName, cl.Name, daemonSet.Status.NumberReady)
			}
		} else {
			log.Debugf("Still waiting for %s daemon set roll out %s...", daemonSetName, cl.Name)
		}
	}, 10*time.Second, deploymentContext.Done())
	err = deploymentContext.Err()
	if err != nil && err != context.Canceled {
		return errors.Wrapf(err, "Error waiting for %s daemon set roll out.", daemonSetName)
	}
	return nil
}
