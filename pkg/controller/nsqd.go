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
	"math"
	"reflect"
	"time"

	"github.com/andyxning/nsq-operator/cmd/nsq-operator/options"
	"github.com/andyxning/nsq-operator/pkg/apis/nsqio"
	"github.com/andyxning/nsq-operator/pkg/common"
	"github.com/andyxning/nsq-operator/pkg/constant"
	"github.com/andyxning/nsq-operator/pkg/generated/informers/externalversions/nsqio/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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

// NsqdController is the controller implementation for Nsqd resources.
type NsqdController struct {
	opts *options.Options

	// kubeClientSet is a standard kubernetes clientset
	kubeClientSet kubernetes.Interface
	// nsqClientSet is a clientset for nsq.io API group
	nsqClientSet nsqclientset.Interface

	statefulSetsLister listerappsv1.StatefulSetLister
	statefulSetsSynced cache.InformerSynced

	configmapsLister listercorev1.ConfigMapLister
	configmapsSynced cache.InformerSynced

	nsqdsLister listernsqv1alpha1.NsqdLister
	nsqdsSynced cache.InformerSynced

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

// NewNsqdController returns a new Nsqd controller.
func NewNsqdController(opts *options.Options, kubeClientSet kubernetes.Interface,
	// nsqClientSet is a clientset for nsq.io API group
	nsqClientSet nsqclientset.Interface,
	statefulSetInformer informersappsv1.StatefulSetInformer,
	configmapInformer informerscorev1.ConfigMapInformer,
	nsqInformer v1alpha1.NsqdInformer) *NsqdController {

	// Create event broadcaster
	// Add nsq-controller types to the default Kubernetes Scheme so Events can be
	// logged for nsq-controller types.
	utilruntime.Must(nsqscheme.AddToScheme(scheme.Scheme))
	klog.Info("Creating event broadcaster for nsqd controller")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: constant.NsqdControllerName})

	controller := &NsqdController{
		opts:               opts,
		kubeClientSet:      kubeClientSet,
		nsqClientSet:       nsqClientSet,
		statefulSetsLister: statefulSetInformer.Lister(),
		statefulSetsSynced: statefulSetInformer.Informer().HasSynced,
		configmapsLister:   configmapInformer.Lister(),
		configmapsSynced:   configmapInformer.Informer().HasSynced,
		nsqdsLister:        nsqInformer.Lister(),
		nsqdsSynced:        nsqInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), nsqio.NsqdKind),
		recorder:           recorder,
	}

	klog.Info("Setting up event handlers for nsqd controller")
	// Set up an event handler for when Nsqd resources change
	nsqInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueNsqd,
		UpdateFunc: func(old, new interface{}) {
			newNsqd := new.(*nsqv1alpha1.Nsqd)
			oldNsqd := old.(*nsqv1alpha1.Nsqd)
			if newNsqd.ResourceVersion == oldNsqd.ResourceVersion && newNsqd.Spec.Replicas == newNsqd.Status.AvailableReplicas {
				// Periodic resync will send update events for all known StatefulSets.
				// Two different versions of the same StatefulSet will always have different RVs.
				return
			}
			controller.enqueueNsqd(new)
		},
	})
	// Set up an event handler for when StatefulSet resources change. This
	// handler will lookup the owner of the given StatefulSet, and if it is
	// owned by a Nsqd resource will enqueue that Nsqd resource for
	// processing. This way, we don't need to implement custom logic for
	// handling StatefulSet resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	statefulSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newSts := new.(*appsv1.StatefulSet)
			oldSts := old.(*appsv1.StatefulSet)
			if newSts.ResourceVersion == oldSts.ResourceVersion {
				// Periodic resync will send update events for all known StatefulSets.
				// Two different versions of the same StatefulSet will always have different RVs.
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
				// Periodic re-sync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
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
func (ndc *NsqdController) Run(threads int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer ndc.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Nsqd controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, ndc.statefulSetsSynced, ndc.nsqdsSynced, ndc.configmapsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch workers to process Nsqd resources
	for i := 0; i < threads; i++ {
		go wait.Until(ndc.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down Nsqd controller")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (ndc *NsqdController) runWorker() {
	for ndc.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (ndc *NsqdController) processNextWorkItem() bool {
	obj, shutdown := ndc.workqueue.Get()

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
		defer ndc.workqueue.Done(obj)
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
			ndc.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Nsqd resource to be synced.
		if err := ndc.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			ndc.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %v, requeuing", key, err)
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		ndc.workqueue.Forget(obj)
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
// converge the two. It then updates the Status block of the Nsqd resource
// with the current status of the resource.
func (ndc *NsqdController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Nsqd resource with this namespace/name
	nd, err := ndc.nsqdsLister.Nsqds(namespace).Get(name)
	if err != nil {
		// The Nsqd resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("nsqd '%s' in work queue no longer exists", key))
			return nil
		}

		klog.Errorf("Get nsqd %s/%s error: %v", nd.Namespace, nd.Name, err)
		return err
	}

	// Get the configmap with the name derived from nsqd cluster name
	configmap, err := ndc.configmapsLister.ConfigMaps(nd.Namespace).Get(common.NsqdConfigMapName(nd.Name))
	// If the resource doesn't exist, we'll return cause that without the configmap,
	// nsqd can not assemble the command line arguments to start.
	if err != nil {
		klog.Errorf("Get configmap for nsqd %s/%s error: %v", nd.Namespace, nd.Name, err)
		return err
	}

	// Update ownerReferences for nsqd configmap
	err = ndc.updateNsqdConfigMapOwnerReference(nd)
	if err != nil {
		klog.Errorf("Update ownerReferences for nsqd configmap %s/%s error: %v",
			nd.Namespace, common.NsqdConfigMapName(nd.Name), err)
		return err
	}

	configmapHash, err := common.Hash(configmap.Data)
	if err != nil {
		klog.Errorf("Hash configmap data for nsqd %s/%s error: %v", nd.Namespace, nd.Name, err)
		return err
	}

	statefulSetName := common.NsqdStatefulSetName(nd.Name)
	statefulSet, err := ndc.statefulSetsLister.StatefulSets(nd.Namespace).Get(statefulSetName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		klog.Infof("StatefulSet for nsqd %s/%s does not exist. Create it", nd.Namespace, nd.Name)
		statefulSet, err = ndc.kubeClientSet.AppsV1().StatefulSets(nd.Namespace).Create(ndc.newStatefulSet(nd, string(configmapHash)))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		klog.Errorf("Get/Create statefulset for nsqd %s/%s error: %v", nd.Namespace, nd.Name, err)
		return err
	}

	if statefulSet.Status.ReadyReplicas != *statefulSet.Spec.Replicas {
		ctx, cancel := context.WithCancel(context.Background())
		wait.UntilWithContext(ctx, func(ctx context.Context) {
			nsqdSs, err := ndc.kubeClientSet.AppsV1().StatefulSets(nd.Namespace).Get(common.NsqdStatefulSetName(nd.Name), metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Failed to get nsqd %s/%s statefulset. Error: %v", nd.Namespace, nd.Name, err)
				return
			}

			if nsqdSs.Status.ReadyReplicas != *nsqdSs.Spec.Replicas {
				klog.Infof("Waiting for nsqd %s/%s pods ready", nd.Namespace, nd.Name)
				return
			}

			klog.Infof("Nsqd %s/%s reaches it spec", nd.Namespace, nd.Name)
			cancel()
		}, constant.NsqdStatusCheckPeriod)
	}

	// If the StatefulSet is not controlled by this Nsqd resource, we should log
	// a warning to the event recorder and return
	if !metav1.IsControlledBy(statefulSet, nd) {
		statefulSet.GetCreationTimestamp()
		msg := fmt.Sprintf(constant.StatefulSetResourceNotOwnedByNsqd, statefulSet.Name)
		ndc.recorder.Event(nd, corev1.EventTypeWarning, nsqerror.ErrResourceNotOwnedByNsqd, msg)
		return fmt.Errorf(msg)
	}

	klog.V(6).Infof("New configmap hash: %v", configmapHash)
	klog.V(6).Infof("Old configmap hash: %v", statefulSet.Spec.Template.Annotations[constant.NsqConfigMapAnnotationKey])
	klog.V(6).Infof("New configmap data: %v", configmap.Data)
	if statefulSet.Spec.Template.Annotations[constant.NsqConfigMapAnnotationKey] != configmapHash {
		klog.Infof("New configmap detected. New config: %#v", configmap.Data)
		var statefulSetNew *appsv1.StatefulSet
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			// Retrieve the latest version of statefulSet before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			statefulSetOld, err := ndc.kubeClientSet.AppsV1().StatefulSets(nd.Namespace).Get(statefulSetName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get statefulSet %s/%s from apiserver error: %v", nd.Namespace, statefulSetName, err)
			}
			statefulSetCopy := statefulSetOld.DeepCopy()
			statefulSetCopy.Spec.Template.Annotations = map[string]string{
				constant.NsqConfigMapAnnotationKey: configmapHash,
			}

			statefulSetNew, err = ndc.kubeClientSet.AppsV1().StatefulSets(nd.Namespace).Update(statefulSetCopy)
			return err
		})

		// If an error occurs during Update, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		//
		// If no error occurs, just return to give kubernetes some time to make
		// adjustment according to the new spec.
		if err == nil {
			klog.V(6).Infof("New statefulset %v annotation under configmap change: %v", statefulSet.Name, statefulSetNew.Spec.Template.Annotations[constant.NsqConfigMapAnnotationKey])
		}
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		wait.UntilWithContext(ctx, func(ctx context.Context) {
			labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": common.NsqdStatefulSetName(nd.Name)}}
			selector, err := metav1.LabelSelectorAsSelector(labelSelector)
			if err != nil {
				klog.Errorf("Failed to generate label selector for nsqd %s/%s: %v", nd.Namespace, nd.Name, err)
				return
			}

			podList, err := ndc.kubeClientSet.CoreV1().Pods(nd.Namespace).List(metav1.ListOptions{
				LabelSelector: selector.String(),
			})

			if err != nil {
				klog.Errorf("Failed to list nsqd %s/%s pods: %v", nd.Namespace, nd.Name, err)
				return
			}

			for _, pod := range podList.Items {
				if val, exists := pod.GetAnnotations()[constant.NsqConfigMapAnnotationKey]; exists && val != configmapHash {
					klog.Infof("Spec and status signature annotation do not match for nsqd %s/%s. Pod: %v. "+
						"Spec signature annotation: %v, new signature annotation: %v",
						nd.Namespace, nd.Name, pod.Name, pod.GetAnnotations()[constant.NsqConfigMapAnnotationKey], configmapHash)
					return
				}
			}

			nsqdSs, err := ndc.kubeClientSet.AppsV1().StatefulSets(nd.Namespace).Get(common.NsqdStatefulSetName(nd.Name), metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Failed to get nsqd %s/%s statefulset", nd.Namespace, nd.Name)
				return
			}

			if nsqdSs.Status.ReadyReplicas != *nsqdSs.Spec.Replicas {
				klog.Errorf("Waiting for nsqd %s/%s pods ready", nd.Namespace, nd.Name)
				return
			}

			klog.Infof("Nsqd %s/%s configmap change rolling update success", nd.Namespace, nd.Name)
			cancel()
		}, constant.NsqdStatusCheckPeriod)
	}

	// If this number of the replicas on the Nsqd resource is specified, and the
	// number does not equal the current desired replicas on the StatefulSet, we
	// should update the StatefulSet resource. If the image changes, we also should update
	// the deployment resource.
	klog.V(2).Infof("nsqd %s/%s replica: %v, statefulset %s/%s replicas: %v", nd.Namespace, nd.Name, nd.Spec.Replicas,
		statefulSet.Namespace, statefulSet.Name, *statefulSet.Spec.Replicas)
	if (nd.Spec.Replicas != *statefulSet.Spec.Replicas) || (nd.Spec.Image != statefulSet.Spec.Template.Spec.Containers[0].Image) {
		if nd.Spec.Replicas < *statefulSet.Spec.Replicas {
			nds, err := ndc.nsqClientSet.NsqV1alpha1().NsqdScales(nd.Namespace).Get(nd.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get nsqdscale %s/%s error: %v", nd.Namespace, nd.Name, err)
				return err
			}

			if len(nds.Status.Metas) >= int(*statefulSet.Spec.Replicas) {
				for i := *statefulSet.Spec.Replicas - 1; i >= nd.Spec.Replicas; i-- {
					nsqdInstanceName := fmt.Sprintf("%s-%d", common.NsqdStatefulSetName(nd.Name), i)
					if value, exists := nds.Status.Metas[nsqdInstanceName]; exists {
						cleared := true
						for _, meta := range value {
							if !(meta.Qps == 0 && meta.Depth == 0) {
								klog.Warningf("Qps/Depth for %s/%s is not zero: %+v", nd.Namespace, nsqdInstanceName, meta)
								cleared = false
								break
							}
						}

						if !cleared {
							klog.Warningf("Waiting for %s/%s qps/depth reach zero", nd.Namespace, nd.Name)
							return nil
						}

						klog.V(2).Infof("Qps/Depth for %s/%s is zero. Clear", nd.Namespace, nsqdInstanceName)
					} else {
						klog.Warningf("Missing qps/depth for %s/%s", nd.Namespace, nd.Name)
						return nil
					}
				}
			} else {
				klog.Warningf("Some nsqd meta data are not updated to nsqdscale %s/%s", nd.Namespace, nd.Name)
				return nil
			}
		}

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			// Retrieve the latest version of statefulset before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			statefulSetOld, err := ndc.kubeClientSet.AppsV1().StatefulSets(nd.Namespace).Get(statefulSetName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get statefulset %s/%s from apiserver error: %v", nd.Namespace, statefulSetName, err)
			}

			statefulSetCopy := statefulSetOld.DeepCopy()
			statefulSetCopy.Spec.Replicas = &nd.Spec.Replicas
			statefulSetCopy.Spec.Template.Spec.Containers[0].Image = nd.Spec.Image
			klog.Infof("Nsqd %s/%s replicas: %d, image: %v, statefulset replicas: %d, image: %v", namespace, name,
				nd.Spec.Replicas, nd.Spec.Image, *statefulSet.Spec.Replicas, statefulSet.Spec.Template.Spec.Containers[0].Image)
			_, err = ndc.kubeClientSet.AppsV1().StatefulSets(nd.Namespace).Update(statefulSetCopy)
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
			nsqdSs, err := ndc.kubeClientSet.AppsV1().StatefulSets(nd.Namespace).Get(common.NsqdStatefulSetName(nd.Name), metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Failed to get nsqd %s/%s statefulset", nd.Namespace, nd.Name)
				return
			}

			if nsqdSs.Status.ReadyReplicas != *nsqdSs.Spec.Replicas {
				klog.Errorf("Waiting for nsqd %s/%s pods ready", nd.Namespace, nd.Name)
				return
			}

			labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": common.NsqdStatefulSetName(nd.Name)}}
			selector, err := metav1.LabelSelectorAsSelector(labelSelector)
			if err != nil {
				klog.Errorf("Failed to generate label selector for nsqd %s/%s: %v", nd.Namespace, nd.Name, err)
				return
			}

			podList, err := ndc.kubeClientSet.CoreV1().Pods(nd.Namespace).List(metav1.ListOptions{
				LabelSelector: selector.String(),
			})

			if err != nil {
				klog.Errorf("Failed to list nsqd %s/%s pods: %v", nd.Namespace, nd.Name, err)
				return
			}

			for _, pod := range podList.Items {
				if pod.Spec.Containers[0].Image != nd.Spec.Image {
					klog.Infof("Spec and status image does not match for nsqd %s/%s. Pod: %v. "+
						"Spec image: %v, new image: %v",
						nd.Namespace, nd.Name, pod.Name, pod.Spec.Containers[0].Image, nd.Spec.Image)
					return
				}
			}

			klog.Infof("Nsqd %s/%s reaches its replicas or image", nd.Namespace, nd.Name)
			cancel()
		}, constant.NsqdStatusCheckPeriod)

	}

	if !(nd.Spec.MemoryOverBookingPercent == nd.Status.MemoryOverBookingPercent &&
		nd.Spec.MessageAvgSize == nd.Status.MessageAvgSize &&
		nd.Spec.MemoryQueueSize == nd.Status.MemoryQueueSize &&
		nd.Spec.ChannelCount == nd.Status.ChannelCount) {

		memRequest, memLimit := ndc.computeNsqdMemoryResource(nd)
		newResources := corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    ndc.opts.NsqdCPULimitResource,
				corev1.ResourceMemory: memRequest,
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    ndc.opts.NsqdCPURequestResource,
				corev1.ResourceMemory: memLimit,
			},
		}

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			// Retrieve the latest version of statefulset before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			statefulSetOld, err := ndc.kubeClientSet.AppsV1().StatefulSets(nd.Namespace).Get(statefulSetName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get statefulset %s/%s from apiserver error: %v", nd.Namespace, statefulSetName, err)
			}

			statefulSetCopy := statefulSetOld.DeepCopy()
			statefulSetCopy.Spec.Template.Spec.Containers[0].Resources = newResources
			klog.Infof("Nsqd %s/%s statefulset resources: %+v", namespace, name, newResources)
			_, err = ndc.kubeClientSet.AppsV1().StatefulSets(nd.Namespace).Update(statefulSetCopy)
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
			nsqdSs, err := ndc.kubeClientSet.AppsV1().StatefulSets(nd.Namespace).Get(common.NsqdStatefulSetName(nd.Name), metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Failed to get nsqd %s/%s statefulset", nd.Namespace, nd.Name)
				return
			}

			if nsqdSs.Status.ReadyReplicas != *nsqdSs.Spec.Replicas {
				klog.Errorf("Waiting for nsqd %s/%s pods ready", nd.Namespace, nd.Name)
				return
			}

			labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": common.NsqdStatefulSetName(nd.Name)}}
			selector, err := metav1.LabelSelectorAsSelector(labelSelector)
			if err != nil {
				klog.Errorf("Failed to generate label selector for nsqd %s/%s: %v", nd.Namespace, nd.Name, err)
				return
			}

			podList, err := ndc.kubeClientSet.CoreV1().Pods(nd.Namespace).List(metav1.ListOptions{
				LabelSelector: selector.String(),
			})

			if err != nil {
				klog.Errorf("Failed to list nsqd %s/%s pods: %v", nd.Namespace, nd.Name, err)
				return
			}

			for _, pod := range podList.Items {
				if !reflect.DeepEqual(pod.Spec.Containers[0].Resources, newResources) {
					klog.Infof("New resources does not match for nsqd %s/%s. Pod: %v. "+
						"Old resources: %#v, new resources: %#v",
						nd.Namespace, nd.Name, pod.Name, pod.Spec.Containers[0].Resources, newResources)
					return
				}
			}

			klog.Infof("Nsqd %s/%s reaches its replicas or image", nd.Namespace, nd.Name)
			cancel()
		}, constant.NsqdStatusCheckPeriod)

	}

	// Finally, we update the status block of the Nsqd resource to reflect the
	// current state of the world
	err = ndc.updateNsqdStatus(nd, statefulSet)
	if err != nil {
		return err
	}

	return nil
}

func (ndc *NsqdController) updateNsqdStatus(nd *nsqv1alpha1.Nsqd, statefulSet *appsv1.StatefulSet) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// Retrieve the latest version of statefulset before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		ndOld, err := ndc.nsqClientSet.NsqV1alpha1().Nsqds(nd.Namespace).Get(nd.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get nsqd %s/%s from apiserver error: %v", nd.Namespace, nd.Name, err)
		}
		// NEVER modify objects from the store. It's a read-only, local cache.
		// You can use DeepCopy() to make a deep copy of original object and modify this copy
		// Or create a copy manually for better performance
		ndCopy := ndOld.DeepCopy()
		ndCopy.Status.AvailableReplicas = nd.Spec.Replicas
		ndCopy.Status.MemoryOverBookingPercent = nd.Spec.MemoryOverBookingPercent
		ndCopy.Status.MessageAvgSize = nd.Spec.MessageAvgSize
		ndCopy.Status.MemoryQueueSize = nd.Spec.MemoryQueueSize
		ndCopy.Status.ChannelCount = nd.Spec.ChannelCount
		// If the CustomResourceSubresources feature gate is not enabled,
		// we must use Update instead of UpdateStatus to update the Status block of the Nsqd resource.
		// UpdateStatus will not allow changes to the Spec of the resource,
		// which is ideal for ensuring nothing other than resource status has been updated.
		_, err = ndc.nsqClientSet.NsqV1alpha1().Nsqds(nd.Namespace).Update(ndCopy)
		return err
	})

	return err
}

// enqueueNsqd takes a Nsqd resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Nsqd.
func (ndc *NsqdController) enqueueNsqd(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	ndc.workqueue.Add(key)
}

// handleObject will take any resource(except configmap deletion) implementing metav1.Object and attempt
// to find the Nsqd resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Nsqd resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (ndc *NsqdController) handleObject(obj interface{}) {
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
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		klog.V(4).Infof("Processing object %v %s/%s", reflect.TypeOf(object), object.GetNamespace(), object.GetName())
		// If this object is not owned by a Nsqd, we should not do anything more
		// with it.
		if ownerRef.Kind != nsqio.NsqdKind {
			klog.V(4).Infof("Owner reference is not %s. Filter object %v %s/%s", nsqio.NsqdKind,
				reflect.TypeOf(object), object.GetNamespace(), object.GetName())
			return
		}

		nd, err := ndc.nsqdsLister.Nsqds(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("Ignoring orphaned object '%s' of nsqd '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		ndc.enqueueNsqd(nd)
		return
	}
}

// handleConfigMapDeletionObject handles configmap deletion.
func (ndc *NsqdController) handleConfigMapDeletionObject(obj interface{}) {
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
		// If this object is not owned by a Nsqd, we should not do anything more
		// with it.
		if ownerRef.Kind != nsqio.NsqdKind {
			return
		}

		klog.Infof("Delete nsqd configmap %#v", obj)
		return
	}
}

// updateNsqdConfigMapOwnerReference updates the OwnerReferences of nsqd configmap to nsqd.
func (ndc *NsqdController) updateNsqdConfigMapOwnerReference(nd *nsqv1alpha1.Nsqd) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// Retrieve the latest version of configmap before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, err := ndc.kubeClientSet.CoreV1().ConfigMaps(nd.Namespace).Get(common.NsqdConfigMapName(nd.Name), metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get latest version of nsqd configmap %v: %v", common.NsqdConfigMapName(nd.Name), err)
		}

		ownerReferences := []metav1.OwnerReference{
			*metav1.NewControllerRef(nd, schema.GroupVersionKind{
				Group:   nsqv1alpha1.SchemeGroupVersion.Group,
				Version: nsqv1alpha1.SchemeGroupVersion.Version,
				Kind:    nsqio.NsqdKind,
			}),
		}

		if reflect.DeepEqual(result.ObjectMeta.OwnerReferences, ownerReferences) {
			return nil
		}

		newCM := result.DeepCopy()
		newCM.ObjectMeta.OwnerReferences = ownerReferences

		_, err = ndc.kubeClientSet.CoreV1().ConfigMaps(nd.Namespace).Update(newCM)
		return err
	})

	return err
}

// computeNsqdMemoryResource updates the OwnerReferences of nsqd configmap to nsqd.
func (ndc *NsqdController) computeNsqdMemoryResource(nd *nsqv1alpha1.Nsqd) (request resource.Quantity, limit resource.Quantity) {
	count := int64(nd.Spec.ChannelCount + 1) // 1 for topic itself
	singleMemUsage := int64(nd.Spec.MessageAvgSize) * int64(nd.Spec.MemoryQueueSize)

	standardTotalMem := count * singleMemUsage
	AdjustedTotalMem := float64(standardTotalMem) * (1 + float64(nd.Spec.MemoryOverBookingPercent)/100.0)

	request = resource.MustParse(fmt.Sprintf("%vMi", int64(math.Ceil(AdjustedTotalMem/1024.0/1024.0))))
	limit = request

	return request, limit
}

// newStatefulSet creates a new StatefulSet for a Nsqd resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Nsqd resource that 'owns' it.
func (ndc *NsqdController) newStatefulSet(nd *nsqv1alpha1.Nsqd, configMapHash string) *appsv1.StatefulSet {
	memRequest, memLimit := ndc.computeNsqdMemoryResource(nd)

	labels := map[string]string{
		"cluster": common.NsqdStatefulSetName(nd.Name),
	}
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.NsqdStatefulSetName(nd.Name),
			Namespace: nd.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(nd, schema.GroupVersionKind{
					Group:   nsqv1alpha1.SchemeGroupVersion.Group,
					Version: nsqv1alpha1.SchemeGroupVersion.Version,
					Kind:    nsqio.NsqdKind,
				}),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &nd.Spec.Replicas,
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
							Name:  nd.Name,
							Image: nd.Spec.Image,
							Env: []corev1.EnvVar{
								{
									Name: "NODE_IP",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
								{
									Name: "POD_IP",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
								{
									Name:  constant.ClusterNameEnv,
									Value: nd.Name,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      common.NsqdConfigMapName(nd.Name),
									MountPath: constant.NsqConfigMapMountPath,
								},
								{
									Name:      common.NsqdVolumeClaimTemplatesName(nd.Name),
									MountPath: constant.NsqdDataMountPath,
								},
								{
									Name:      constant.LogVolumeName,
									MountPath: common.NsqdLogMountPath(nd.Name),
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    ndc.opts.NsqdCPULimitResource,
									corev1.ResourceMemory: memRequest,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    ndc.opts.NsqdCPURequestResource,
									corev1.ResourceMemory: memLimit,
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/ping",
										Port:   intstr.FromInt(constant.NsqdHttpPort),
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
										Port:   intstr.FromInt(constant.NsqdHttpPort),
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
						{
							Name:  "reporter",
							Image: nd.Spec.Image,
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name:  constant.ClusterNameEnv,
									Value: nd.Name,
								},
							},
							Command: []string{"reporter", "--alsologtostderr=false", "--logtostderr=false", fmt.Sprintf("--log_dir=%s", common.QpsReporterLogDir(nd.Name))},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      constant.LogVolumeName,
									MountPath: common.QpsReporterLogMountPath(nd.Name),
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    ndc.opts.QpsReporterCPULimitResource,
									corev1.ResourceMemory: ndc.opts.QpsReporterMemoryLimitResource,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    ndc.opts.QpsReporterCPURequestResource,
									corev1.ResourceMemory: ndc.opts.QpsReporterMemoryRequestResource,
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(constant.NsqdScaleHttpPort),
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
										Path:   "/healthz",
										Port:   intstr.FromInt(constant.NsqdScaleHttpPort),
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
							Name: common.NsqdConfigMapName(nd.Name),
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: common.NsqdConfigMapName(nd.Name),
									},
								},
							},
						},
						{
							Name: constant.LogVolumeName,
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: nd.Spec.LogMappingDir,
								},
							},
						},
					},
					TerminationGracePeriodSeconds: &ndc.opts.NsqdTerminationGracePeriodSeconds,
				},
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      common.NsqdVolumeClaimTemplatesName(nd.Name),
						Namespace: nd.Namespace,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						StorageClassName: &nd.Spec.StorageClassName,
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceStorage: ndc.opts.NsqdPVCStorageResource,
							},
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: ndc.opts.NsqdPVCStorageResource,
							},
						},
					},
				},
			},
		},
	}
}
