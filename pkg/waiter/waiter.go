package waiter

import (
	"context"
	"time"

	"github.com/dimaunx/armada/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// WaitForPodsRunning waits for pods to be running
func WaitForPodsRunning(cl *config.Cluster, c kubernetes.Interface, namespace, selector string, replicas int) error {
	ctx := context.Background()
	log.Warnf("Waiting for pods to be running. label: %q, namespace: %q, replicas: %v, duration: %v, *types: %s.", selector, namespace, replicas, config.WaitDurationResources, cl.Name)
	podsContext, cancel := context.WithTimeout(ctx, config.WaitDurationResources)
	wait.Until(func() {
		podList, err := c.CoreV1().Pods(namespace).List(metav1.ListOptions{
			LabelSelector: selector,
			FieldSelector: "status.phase=Running",
		})
		if err == nil && len(podList.Items) == replicas {
			log.Infof("✔ All pods with label: %q are running for %s.", selector, cl.Name)
			cancel()
		} else {
			log.Debugf("Still waiting for pods. label: %q, namespace: %q, replicas: %v, duration: %v, cluster: %s.", selector, namespace, replicas, config.WaitDurationResources, cl.Name)
		}
	}, 30*time.Second, podsContext.Done())

	err := podsContext.Err()
	if err != nil && err != context.Canceled {
		return errors.Wrap(err, "Error waiting for pods to be running.")
	}
	return nil
}

// WaitForDeployment waits for deployment roll out
func WaitForDeployment(cl *config.Cluster, c kubernetes.Interface, namespace, deploymentName string) error {
	ctx := context.Background()
	log.Debugf("Waiting up to %v for %s deployment roll out %s...", config.WaitDurationResources, deploymentName, cl.Name)
	deploymentContext, cancel := context.WithTimeout(ctx, config.WaitDurationResources)
	wait.Until(func() {
		deployment, err := c.AppsV1().Deployments(namespace).Get(deploymentName, metav1.GetOptions{})
		if err == nil {
			if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
				log.Debugf("✔ %s successfully deployed to %s, ready replicas: %v", deploymentName, cl.Name, deployment.Status.ReadyReplicas)
				cancel()
			} else {
				log.Debugf("Still waiting for %s deployment for %s, ready replicas: %v", deploymentName, cl.Name, deployment.Status.ReadyReplicas)
			}
		} else {
			log.Debugf("Still waiting for %s deployment roll out %s...", deploymentName, cl.Name)
		}
	}, 10*time.Second, deploymentContext.Done())
	err := deploymentContext.Err()
	if err != nil && err != context.Canceled {
		return errors.Wrapf(err, "Error waiting for %s deployment roll out.", deploymentName)
	}
	return nil
}

// WaitForDaemonSet waits for daemon set roll out
func WaitForDaemonSet(cl *config.Cluster, c kubernetes.Interface, namespace, daemonSetName string) error {
	ctx := context.Background()
	log.Debugf("Waiting up to %v for %s daemon set roll out %s...", config.WaitDurationResources, daemonSetName, cl.Name)
	deploymentContext, cancel := context.WithTimeout(ctx, config.WaitDurationResources)
	wait.Until(func() {
		daemonSet, err := c.AppsV1().DaemonSets(namespace).Get(daemonSetName, metav1.GetOptions{})
		if err == nil && daemonSet.Status.CurrentNumberScheduled > 0 {
			if daemonSet.Status.NumberReady == daemonSet.Status.DesiredNumberScheduled {
				log.Debugf("✔ %s successfully rolled out to %s, ready replicas: %v", daemonSetName, cl.Name, daemonSet.Status.NumberReady)
				cancel()
			} else {
				log.Debugf("Still waiting for %s daemon set roll out for %s, ready replicas: %v", daemonSetName, cl.Name, daemonSet.Status.NumberReady)
			}
		} else {
			log.Debugf("Still waiting for %s daemon set roll out %s...", daemonSetName, cl.Name)
		}
	}, 10*time.Second, deploymentContext.Done())
	err := deploymentContext.Err()
	if err != nil && err != context.Canceled {
		return errors.Wrapf(err, "Error waiting for %s daemon set roll out.", daemonSetName)
	}
	return nil
}
