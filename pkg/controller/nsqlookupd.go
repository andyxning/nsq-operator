/*
Copyright 2019 The NSQ-Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/andyxning/nsq-operator/cmd/nsq-operator/options"
	"github.com/andyxning/nsq-operator/pkg/apis/nsqio"
	"github.com/andyxning/nsq-operator/pkg/common"
	"github.com/andyxning/nsq-operator/pkg/constant"
	"github.com/andyxning/nsq-operator/pkg/generated/informers/externalversions/nsqio/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	nsqclientset "github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	listerappsv1 "k8s.io/client-go/listers/apps/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"

	nsqv1alpha1 "github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
	nsqerror "github.com/andyxning/nsq-operator/pkg/error"
	nsqscheme "github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned/scheme"
	listernsqv1alpha1 "github.com/andyxning/nsq-operator/pkg/generated/listers/nsqio/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informersappsv1 "k8s.io/client-go/informers/apps/v1"
	informerscorev1 "k8s.io/client-go/informers/core/v1"
)

// NsqLookupdController is the reconcile implementation for NsqLookupd resources.
type NsqLookupdController struct {
	opts *options.Options

	// kubeClientSet is a standard kubernetes clientset
	kubeClientSet kubernetes.Interface
	// nsqClientSet is a clientset for nsq.io API group
	nsqClientSet nsqclientset.Interface

	deploymentsLister listerappsv1.DeploymentLister
	deploymentsSynced cache.InformerSynced

	configmapsLister listercorev1.ConfigMapLister
	configmapsSynced cache.InformerSynced

	nsqLookupdsLister listernsqv1alpha1.NsqLookupdLister
	nsqLookupdsSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewNsqLookupdController returns a NsqLookupd controller.
func NewNsqLookupdController(opts *options.Options, kubeClientSet kubernetes.Interface,
	// nsqClientSet is a clientset for nsq.io API group
	nsqClientSet nsqclientset.Interface,
	deploymentInformer informersappsv1.DeploymentInformer,
	configmapInformer informerscorev1.ConfigMapInformer,
	nsqInformer v1alpha1.NsqLookupdInformer) *NsqLookupdController {

	// Create event broadcaster
	// Add nsq-controller types to the default Kubernetes Scheme so Events can be
	// logged for nsq-controller types.
	utilruntime.Must(nsqscheme.AddToScheme(scheme.Scheme))
	klog.Info("Creating event broadcaster for nsqlookupd controller")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: constant.NsqLookupdControllerName})

	controller := &NsqLookupdController{
		opts:              opts,
		kubeClientSet:     kubeClientSet,
		nsqClientSet:      nsqClientSet,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		configmapsLister:  configmapInformer.Lister(),
		configmapsSynced:  configmapInformer.Informer().HasSynced,
		nsqLookupdsLister: nsqInformer.Lister(),
		nsqLookupdsSynced: nsqInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), nsqio.NsqLookupdKind),
		recorder:          recorder,
	}

	klog.Info("Setting up event handlers for nsqlookupd controller")
	// Set up an event handler for when NsqLookupd resources change
	nsqInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueNsqLookupd,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueNsqLookupd(new)
		},
	})
	// Set up an event handler for when deployment resources change. This
	// handler will lookup the owner of the given deployment, and if it is
	// owned by a NsqLookupd resource will enqueue that NsqLookupd resource for
	// processing. This way, we don't need to implement custom logic for
	// handling deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known deployments.
				// Two different versions of the same deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	configmapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newCM := new.(*corev1.ConfigMap)
			oldCM := old.(*corev1.ConfigMap)

			oldDataHash, err := common.Hash(oldCM.Data)
			if err != nil {
				klog.Warningf("Compute old configmap data hash error: %v. Skipping this event", err)
				return
			}

			newDataHash, err := common.Hash(newCM.Data)
			if err != nil {
				klog.Warningf("Compute new configmap data hash error: %v. Skipping this event", err)
				return
			}

			if newCM.ResourceVersion == oldCM.ResourceVersion || oldDataHash == newDataHash {
				// Periodic re-sync will send update events for all known deployments.
				// Two different versions of the same deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleConfigMapDeletionObject,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (nlc *NsqLookupdController) Run(threads int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer nlc.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting NsqLookupd controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, nlc.deploymentsSynced, nlc.nsqLookupdsSynced, nlc.configmapsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch workers to process NsqLookupd resources
	for i := 0; i < threads; i++ {
		go wait.Until(nlc.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down NsqLookupd controller")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (nlc *NsqLookupdController) runWorker() {
	for nlc.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (nlc *NsqLookupdController) processNextWorkItem() bool {
	obj, shutdown := nlc.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer nlc.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			nlc.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// NsqLookupd resource to be synced.
		if err := nlc.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			nlc.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %v, requeuing", key, err)
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		nlc.workqueue.Forget(obj)
		klog.V(2).Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the NsqLookupd resource
// with the current status of the resource.
func (nlc *NsqLookupdController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the NsqLookupd resource with this namespace/name
	nl, err := nlc.nsqLookupdsLister.NsqLookupds(namespace).Get(name)
	if err != nil {
		// The NsqLookupd resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("NsqLookupd '%s' in work queue no longer exists", key))
			return nil
		}

		klog.Errorf("Get nsqlookupd %s/%s error: %v", nl.Namespace, nl.Name, err)
		return err
	}

	// Get the configmap with the name derived from NsqLookupd cluster name
	configmap, err := nlc.configmapsLister.ConfigMaps(nl.Namespace).Get(common.NsqLookupdConfigMapName(nl.Name))
	// If the resource doesn't exist, we'll return cause that without the configmap,
	// NsqLookupd can not assemble the command line arguments to start.
	if err != nil {
		klog.Errorf("Get configmap for nsqlookupd %s/%s error: %v", nl.Namespace, nl.Name, err)
		return err
	}

	// Update ownerReferences for nsqlookupd configmap
	err = nlc.updateNsqLookupdConfigMapOwnerReference(nl)
	if err != nil {
		klog.Errorf("Update ownerReferences for nsqlookupd configmap %s/%s error: %v",
			nl.Namespace, common.NsqLookupdConfigMapName(nl.Name), err)
		return err
	}

	configmapHash, err := common.Hash(configmap.Data)
	if err != nil {
		klog.Errorf("Hash configmap data for nsqlookupd %s/%s error: %v", nl.Namespace, nl.Name, err)
		return err
	}

	deploymentName := common.NsqLookupdDeploymentName(nl.Name)
	deployment, err := nlc.deploymentsLister.Deployments(nl.Namespace).Get(deploymentName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		klog.Infof("deployment for nsqlookupd %s/%s does not exist. Create it", nl.Namespace, nl.Name)
		deployment, err = nlc.kubeClientSet.AppsV1().Deployments(nl.Namespace).Create(nlc.newDeployment(nl, string(configmapHash)))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		klog.Errorf("Get/Create deployment for nsqlookupd %s/%s error: %v", nl.Namespace, nl.Name, err)
		return err
	}

	if !(deployment.Status.ReadyReplicas == *deployment.Spec.Replicas && deployment.Status.Replicas == deployment.Status.ReadyReplicas) {
		ctx, cancel := context.WithCancel(context.Background())
		wait.UntilWithContext(ctx, func(ctx context.Context) {
			nsqLookupdDep, err := nlc.kubeClientSet.AppsV1().Deployments(nl.Namespace).Get(common.NsqLookupdDeploymentName(nl.Name), metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Failed to get nsqlookupd %s/%s deployment", nl.Namespace, nl.Name)
				return
			}

			if !(nsqLookupdDep.Status.ReadyReplicas == *nsqLookupdDep.Spec.Replicas && nsqLookupdDep.Status.Replicas == nsqLookupdDep.Status.ReadyReplicas) {
				klog.Infof("Waiting for nsqlookupd %s/%s pods ready", nl.Namespace, nl.Name)
				return
			}

			klog.Infof("NsqLookupd %s/%s reaches it spec", nl.Namespace, nl.Name)
			cancel()
		}, constant.NsqLookupdStatusCheckPeriod)
	}

	// If the deployment is not controlled by this NsqLookupd resource, we should log
	// a warning to the event recorder and return
	if !metav1.IsControlledBy(deployment, nl) {
		deployment.GetCreationTimestamp()
		msg := fmt.Sprintf(constant.DeploymentResourceNotOwnedByNsqLookupd, deployment.Name)
		nlc.recorder.Event(nl, corev1.EventTypeWarning, nsqerror.ErrResourceNotOwnedByNsqLookupd, msg)
		return fmt.Errorf(msg)
	}

	nsqLookupdInstancesChanged := false
	klog.V(6).Infof("New configmap hash: %v", configmapHash)
	klog.V(6).Infof("Old configmap hash: %v", deployment.Spec.Template.Annotations[constant.NsqConfigMapAnnotationKey])
	klog.V(6).Infof("New configmap data: %v", configmap.Data)
	if deployment.Spec.Template.Annotations[constant.NsqConfigMapAnnotationKey] != configmapHash {
		nsqLookupdInstancesChanged = true
		klog.Infof("New configmap detected. New config: %v", configmap.Data)
		var deploymentNew *appsv1.Deployment
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			// Retrieve the latest version of deployment before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			deploymentOld, err := nlc.kubeClientSet.AppsV1().Deployments(nl.Namespace).Get(deploymentName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get deployment %s/%s from apiserver error: %v", nl.Namespace, deploymentName, err)
			}
			deploymentCopy := deploymentOld.DeepCopy()
			deploymentCopy.Spec.Template.Annotations = map[string]string{
				constant.NsqConfigMapAnnotationKey: configmapHash,
			}
			deploymentNew, err = nlc.kubeClientSet.AppsV1().Deployments(nl.Namespace).Update(deploymentCopy)
			return err
		})

		// If an error occurs during Update, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		//
		// If no error occurs, just return to give kubernetes some time to make
		// adjustment according to the new spec.
		if err == nil {
			klog.V(6).Infof("New deployment %v annotation under configmap change: %v", deployment.Name, deploymentNew.Spec.Template.Annotations[constant.NsqConfigMapAnnotationKey])
		}

		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		wait.UntilWithContext(ctx, func(ctx context.Context) {
			labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": common.NsqLookupdDeploymentName(nl.Name)}}
			selector, err := metav1.LabelSelectorAsSelector(labelSelector)
			if err != nil {
				klog.Errorf("Failed to generate label selector for nsqlookupd %s/%s: %v", nl.Namespace, nl.Name, err)
				return
			}

			podList, err := nlc.kubeClientSet.CoreV1().Pods(nl.Namespace).List(metav1.ListOptions{
				LabelSelector: selector.String(),
			})

			if err != nil {
				klog.Errorf("Failed to list nsqlookupd %s/%s pods: %v", nl.Namespace, nl.Name, err)
				return
			}

			for _, pod := range podList.Items {
				if val, exists := pod.GetAnnotations()[constant.NsqConfigMapAnnotationKey]; exists && val != configmapHash {
					klog.Infof("Spec and status signature annotation do not match for nsqlookupd %s/%s. Pod: %v. "+
						"Spec signature annotation: %v, new signature annotation: %v",
						nl.Namespace, nl.Name, pod.Name, pod.GetAnnotations()[constant.NsqConfigMapAnnotationKey], configmapHash)
					return
				}
			}

			nsqLookupdDep, err := nlc.kubeClientSet.AppsV1().Deployments(nl.Namespace).Get(common.NsqAdminDeploymentName(nl.Name), metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Failed to get nsqlookupd %s/%s deployment", nl.Namespace, nl.Name)
				return
			}

			if !(nsqLookupdDep.Status.ReadyReplicas == *nsqLookupdDep.Spec.Replicas && nsqLookupdDep.Status.Replicas == nsqLookupdDep.Status.ReadyReplicas) {
				klog.Errorf("Waiting for nsqlookupd %s/%s pods ready", nl.Namespace, nl.Name)
				return
			}

			klog.Infof("NsqLookupd %s/%s configmap change rolling update success", nl.Namespace, nl.Name)
			cancel()
		}, constant.NsqLookupdStatusCheckPeriod)
	}

	// If this number of the replicas on the NsqLookupd resource is specified, and the
	// number does not equal the current desired replicas on the deployment, we
	// should update the deployment resource. If the image changes, we also should update
	// the deployment resource.
	if (nl.Spec.Replicas != *deployment.Spec.Replicas) || (nl.Spec.Image != deployment.Spec.Template.Spec.Containers[0].Image) {
		nsqLookupdInstancesChanged = true
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			// Retrieve the latest version of deployment before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			deploymentOld, err := nlc.kubeClientSet.AppsV1().Deployments(nl.Namespace).Get(deploymentName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get deployment %s/%s from apiserver error: %v", nl.Namespace, deploymentName, err)
			}
			deploymentCopy := deploymentOld.DeepCopy()
			deploymentCopy.Spec.Replicas = &nl.Spec.Replicas
			deploymentCopy.Spec.Template.Spec.Containers[0].Image = nl.Spec.Image
			klog.Infof("NsqLookupd %s/%s replicas: %d, image: %v, deployment replicas: %d, image: %v", namespace, name,
				nl.Spec.Replicas, nl.Spec.Image, *deployment.Spec.Replicas, deployment.Spec.Template.Spec.Containers[0].Image)
			_, err = nlc.kubeClientSet.AppsV1().Deployments(nl.Namespace).Update(deploymentCopy)
			return err
		})

		// If an error occurs during Update, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		wait.UntilWithContext(ctx, func(ctx context.Context) {
			nsqLookupdDep, err := nlc.kubeClientSet.AppsV1().Deployments(nl.Namespace).Get(common.NsqAdminDeploymentName(nl.Name), metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Failed to get nsqlookupd %s/%s deployment", nl.Namespace, nl.Name)
				return
			}

			if !(nsqLookupdDep.Status.ReadyReplicas == *nsqLookupdDep.Spec.Replicas && nsqLookupdDep.Status.Replicas == nsqLookupdDep.Status.ReadyReplicas) {
				klog.Errorf("Waiting for nsqlookupd %s/%s pods ready", nl.Namespace, nl.Name)
				return
			}

			labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": common.NsqLookupdDeploymentName(nl.Name)}}
			selector, err := metav1.LabelSelectorAsSelector(labelSelector)
			if err != nil {
				klog.Errorf("Failed to generate label selector for nsqlookupd %s/%s: %v", nl.Namespace, nl.Name, err)
				return
			}

			podList, err := nlc.kubeClientSet.CoreV1().Pods(nl.Namespace).List(metav1.ListOptions{
				LabelSelector: selector.String(),
			})

			if err != nil {
				klog.Errorf("Failed to list nsqlookupd %s/%s pods: %v", nl.Namespace, nl.Name, err)
				return
			}

			for _, pod := range podList.Items {
				if pod.Spec.Containers[0].Image != nl.Spec.Image {
					klog.Infof("Spec and status image does not match for nsqlookupd %s/%s. Pod: %v. "+
						"Spec image: %v, new image: %v",
						nl.Namespace, nl.Name, pod.Name, pod.Spec.Containers[0].Image, nl.Spec.Image)
					return
				}
			}

			klog.Infof("NsqLookupd %s/%s reaches its replicas or image", nl.Namespace, nl.Name)
			cancel()
		}, constant.NsqLookupdStatusCheckPeriod)
	}

	// Finally, we update the status block of the NsqLookupd resource to reflect the
	// current state of the world
	err = nlc.updateNsqLookupdStatus(nl, deployment)
	if err != nil {
		return err
	}

	nsqAdminConfigMapName := common.NsqAdminConfigMapName(nl.Name)
	nsqAdminConfigMap, err := nlc.configmapsLister.ConfigMaps(nl.Namespace).Get(nsqAdminConfigMapName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": common.NsqLookupdDeploymentName(nl.Name)}}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return fmt.Errorf("failed to generate label selector for nsqlookupd %s/%s: %v", nl.Namespace, nl.Name, err)
		}

		podList, err := nlc.kubeClientSet.CoreV1().Pods(nl.Namespace).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})

		if err != nil {
			return fmt.Errorf("failed to list nsqlookupd %s/%s pods: %v", nl.Namespace, nl.Name, err)
		}

		klog.Infof("Configmap for nsqadmin %s/%s does not exist. Create it", nl.Namespace, nl.Name)
		nsqAdminConfigMap, err = nlc.newNsqAdminConfigMap(nl, podList)
		if err != nil {
			klog.Infof("Gen nsqadmin configmap %s/%s error: %v", nl.Namespace, nl.Name, err)
			return err
		}
		nsqAdminConfigMap, err = nlc.kubeClientSet.CoreV1().ConfigMaps(nl.Namespace).Create(nsqAdminConfigMap)
		if err != nil {
			klog.Errorf("Create nsqadmin %s/%s configmap error: %v", nl.Namespace, nl.Name, err)
			return err
		}
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if nsqLookupdInstancesChanged {
		klog.Infof("Nsqlookupd %s/%s instances change. Update nsqadmin and nsqd configmap", nl.Namespace, nl.Name)
		labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": common.NsqLookupdDeploymentName(nl.Name)}}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return fmt.Errorf("failed to generate label selector for nsqlookupd %s/%s: %v", nl.Namespace, nl.Name, err)
		}

		podList, err := nlc.kubeClientSet.CoreV1().Pods(nl.Namespace).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})

		if err != nil {
			return fmt.Errorf("failed to list nsqlookupd %s/%s pods: %v", nl.Namespace, nl.Name, err)
		}

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			// Retrieve the latest version of configmap before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			result, err := nlc.kubeClientSet.CoreV1().ConfigMaps(nl.Namespace).Get(nsqAdminConfigMapName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get latest version of nsqadmin configmap %v: %v", common.NsqAdminConfigMapName(nl.Name), err)
			}

			data, err := nlc.assembleNsqAdminConfigMapData(nl, podList)
			if err != nil {
				return fmt.Errorf("failed to assemble nsqadmin configmap %s/%s data: %v", nl.Namespace, common.NsqAdminConfigMapName(nl.Name), err)
			}

			newCM := result.DeepCopy()
			newCM.Data = data

			_, err = nlc.kubeClientSet.CoreV1().ConfigMaps(nl.Namespace).Update(newCM)
			return err
		})

		// If an error occurs during Get/Create, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return err
		}

		klog.Infof("Nsqlookupd %s/%s instances change. Update nsqadmin configmap %s/%s success", nl.Namespace, nl.Name,
			nl.Namespace, common.NsqAdminConfigMapName(nl.Name))

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			// Retrieve the latest version of configmap before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			result, err := nlc.kubeClientSet.CoreV1().ConfigMaps(nl.Namespace).Get(common.NsqdConfigMapName(nl.Name), metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get latest version of nsqd configmap %v: %v", common.NsqdConfigMapName(nl.Name), err)
			}

			newCM := result.DeepCopy()
			err = nlc.updateNsqdConfigMapData(newCM, nl, podList)
			if err != nil {
				return fmt.Errorf("failed to update nsqd configmap %s/%s data: %v", nl.Namespace, common.NsqdConfigMapName(nl.Name), err)
			}

			_, err = nlc.kubeClientSet.CoreV1().ConfigMaps(nl.Namespace).Update(newCM)
			return err
		})

		// If an error occurs during Get/Create, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return err
		}

		klog.Infof("Nsqlookupd %s/%s instances change. Update nsqd configmap %s/%s success", nl.Namespace, nl.Name,
			nl.Namespace, common.NsqdConfigMapName(nl.Name))
	}

	// If the configmap is not controlled by this NsqLookupd resource, we should log
	// a warning to the event recorder and return
	if !metav1.IsControlledBy(nsqAdminConfigMap, nl) {
		nsqAdminConfigMap.GetCreationTimestamp()
		msg := fmt.Sprintf(constant.ConfigMapResourceNotOwnedByNsqLookupd, nsqAdminConfigMap.Name)
		nlc.recorder.Event(nl, corev1.EventTypeWarning, nsqerror.ErrResourceNotOwnedByNsqLookupd, msg)
		return fmt.Errorf(msg)
	}

	return nil
}

func (nlc *NsqLookupdController) updateNsqLookupdStatus(nl *nsqv1alpha1.NsqLookupd, deployment *appsv1.Deployment) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// Retrieve the latest version of deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		nlOld, err := nlc.nsqClientSet.NsqV1alpha1().NsqLookupds(nl.Namespace).Get(nl.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get nsqlookupd %s/%s from apiserver error: %v", nl.Namespace, nl.Name, err)
		}
		// NEVER modify objects from the store. It's a read-only, local cache.
		// You can use DeepCopy() to make a deep copy of original object and modify this copy
		// Or create a copy manually for better performance
		nlCopy := nlOld.DeepCopy()
		nlCopy.Status.AvailableReplicas = nlOld.Spec.Replicas
		// If the CustomResourceSubresources feature gate is not enabled,
		// we must use Update instead of UpdateStatus to update the Status block of the NsqLookupd resource.
		// UpdateStatus will not allow changes to the Spec of the resource,
		// which is ideal for ensuring nothing other than resource status has been updated.
		_, err = nlc.nsqClientSet.NsqV1alpha1().NsqLookupds(nl.Namespace).Update(nlCopy)
		return err
	})
	return err
}

// enqueueNsqLookupd takes a NsqLookupd resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than NsqLookupd.
func (nlc *NsqLookupdController) enqueueNsqLookupd(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	nlc.workqueue.Add(key)
}

// handleObject will take any resource(except configmap deletion) implementing metav1.Object and attempt
// to find the NsqLookupd resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that NsqLookupd resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (nlc *NsqLookupdController) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s/%s' from tombstone", object.GetNamespace(), object.GetName())
	}
	klog.V(4).Infof("Processing object %s/%s", object.GetNamespace(), object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a NsqLookupd, we should not do anything more
		// with it.
		if ownerRef.Kind != nsqio.NsqLookupdKind {
			return
		}

		nl, err := nlc.nsqLookupdsLister.NsqLookupds(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("Ignoring orphaned object '%s' of nsqlookupd '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		nlc.enqueueNsqLookupd(nl)
		return
	}
}

// handleConfigMapDeletionObject handles configmap deletion.
func (nlc *NsqLookupdController) handleConfigMapDeletionObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s/%s' from tombstone", object.GetNamespace(), object.GetName())
	}
	klog.V(4).Infof("Processing object %s/%s", object.GetNamespace(), object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a NsqLookupd, we should not do anything more
		// with it.
		if ownerRef.Kind != nsqio.NsqLookupdKind {
			return
		}

		klog.Infof("Delete nsqlookup configmap %#v", obj)
		return
	}
}

// assembleNsqAdminConfigMapData returns nsqadmin configmap data
func (nlc *NsqLookupdController) assembleNsqAdminConfigMapData(nl *nsqv1alpha1.NsqLookupd, podList *corev1.PodList) (map[string]string, error) {
	var addresses []string
	for _, pod := range podList.Items {
		addresses = append(addresses, fmt.Sprintf("%s:%v", pod.Status.PodIP, constant.NsqLookupdHttpPort))
	}

	return map[string]string{
		"nsqadmin": fmt.Sprintf("%s=%q\n%s=%q\n",
			string(constant.NsqAdminCommandArguments), fmt.Sprintf("-http-address=0.0.0.0:%v", constant.NsqAdminHttpPort),
			string(constant.NsqAdminLookupdHttpAddress), common.AssembleNsqLookupdAddresses(addresses)),
	}, nil
}

// updateNsqdConfigMapData returns nsqd configmap data
func (nlc *NsqLookupdController) updateNsqdConfigMapData(newCM *corev1.ConfigMap, nl *nsqv1alpha1.NsqLookupd, podList *corev1.PodList) error {
	var addresses []string
	for _, pod := range podList.Items {
		addresses = append(addresses, fmt.Sprintf("%s:%v", pod.Status.PodIP, constant.NsqLookupdTcpPort))
	}

	nsqdCommandArguments := strings.Split(newCM.Data["nsqd"], "\n")
	if len(nsqdCommandArguments) <= 1 {
		return fmt.Errorf("command arguments for nsqd %s/%s error: there should be at least two lines", nl.Namespace, nl.Name)
	}

	newCM.Data = map[string]string{
		"nsqd": fmt.Sprintf("%s\n%s=%q\n",
			nsqdCommandArguments[0],
			constant.NsqdLookupdTcpAddress, common.AssembleNsqLookupdAddresses(addresses)),
	}
	return nil
}

// updateNsqLookupdConfigMapOwnerReference updates the OwnerReferences of nsqlookupd configmap to nsqlookupd.
func (nlc *NsqLookupdController) updateNsqLookupdConfigMapOwnerReference(nl *nsqv1alpha1.NsqLookupd) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// Retrieve the latest version of configmap before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, err := nlc.kubeClientSet.CoreV1().ConfigMaps(nl.Namespace).Get(common.NsqLookupdConfigMapName(nl.Name), metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get latest version of nsqlookupd configmap %v: %v", common.NsqLookupdConfigMapName(nl.Name), err)
		}

		ownerReferences := []metav1.OwnerReference{
			*metav1.NewControllerRef(nl, schema.GroupVersionKind{
				Group:   nsqv1alpha1.SchemeGroupVersion.Group,
				Version: nsqv1alpha1.SchemeGroupVersion.Version,
				Kind:    nsqio.NsqLookupdKind,
			}),
		}

		if reflect.DeepEqual(result.ObjectMeta.OwnerReferences, ownerReferences) {
			return nil
		}

		newCM := result.DeepCopy()
		newCM.ObjectMeta.OwnerReferences = ownerReferences

		_, err = nlc.kubeClientSet.CoreV1().ConfigMaps(nl.Namespace).Update(newCM)
		return err
	})

	return err
}

// newNsqAdminConfigMap creates a configmap for a NsqAdmin resource.
func (nlc *NsqLookupdController) newNsqAdminConfigMap(nl *nsqv1alpha1.NsqLookupd, podList *corev1.PodList) (*corev1.ConfigMap, error) {
	data, err := nlc.assembleNsqAdminConfigMapData(nl, podList)
	if err != nil {
		return nil, err
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.NsqAdminConfigMapName(nl.Name),
			Namespace: nl.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(nl, schema.GroupVersionKind{
					Group:   nsqv1alpha1.SchemeGroupVersion.Group,
					Version: nsqv1alpha1.SchemeGroupVersion.Version,
					Kind:    nsqio.NsqLookupdKind,
				}),
			},
		},
		Data: data,
	}, nil
}

// newDeployment creates a new deployment for a NsqLookupd resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the NsqLookupd resource that 'owns' it.
func (nlc *NsqLookupdController) newDeployment(nl *nsqv1alpha1.NsqLookupd, configMapHash string) *appsv1.Deployment {
	labels := map[string]string{
		"cluster": common.NsqLookupdDeploymentName(nl.Name),
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.NsqLookupdDeploymentName(nl.Name),
			Namespace: nl.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(nl, schema.GroupVersionKind{
					Group:   nsqv1alpha1.SchemeGroupVersion.Group,
					Version: nsqv1alpha1.SchemeGroupVersion.Version,
					Kind:    nsqio.NsqLookupdKind,
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &nl.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						constant.NsqConfigMapAnnotationKey: configMapHash,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  nl.Name,
							Image: nl.Spec.Image,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      common.NsqLookupdConfigMapName(nl.Name),
									MountPath: constant.NsqConfigMapMountPath,
								},
								{
									Name:      constant.LogVolumeName,
									MountPath: common.NsqLookupdLogMountPath(nl.Name),
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    nlc.opts.NsqLookupdCPULimitResource,
									corev1.ResourceMemory: nlc.opts.NsqLookupdMemoryLimitResource,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    nlc.opts.NsqLookupdCPURequestResource,
									corev1.ResourceMemory: nlc.opts.NsqLookupdMemoryRequestResource,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  constant.ClusterNameEnv,
									Value: nl.Name,
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/ping",
										Port:   intstr.FromInt(constant.NsqLookupdHttpPort),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 3,
								TimeoutSeconds:      2,
								PeriodSeconds:       30,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/ping",
										Port:   intstr.FromInt(constant.NsqLookupdHttpPort),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 3,
								TimeoutSeconds:      2,
								PeriodSeconds:       30,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: common.NsqLookupdConfigMapName(nl.Name),
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: common.NsqLookupdConfigMapName(nl.Name),
									},
								},
							},
						},
						{
							Name: constant.LogVolumeName,
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: nl.Spec.LogMappingDir,
								},
							},
						},
					},
					TerminationGracePeriodSeconds: &nlc.opts.NsqLookupdTerminationGracePeriodSeconds,
				},
			},
		},
	}
}
