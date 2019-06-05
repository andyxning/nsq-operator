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
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/andyxning/nsq-operator/cmd/nsq-operator/options"
	"github.com/andyxning/nsq-operator/pkg/apis/nsqio"
	"github.com/andyxning/nsq-operator/pkg/constant"
	"github.com/andyxning/nsq-operator/pkg/generated/informers/externalversions/nsqio/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	apimachinerycache "k8s.io/apimachinery/pkg/util/cache"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	nsqv1alpha1 "github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
	nsqerror "github.com/andyxning/nsq-operator/pkg/error"
	nsqscheme "github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned/scheme"
	listernsqv1alpha1 "github.com/andyxning/nsq-operator/pkg/generated/listers/nsqio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NsqdScaleController is the controller implementation for NsqdScale resources.
type NsqdScaleController struct {
	opts *options.Options

	// kubeClientSet is a standard kubernetes clientset
	kubeClientSet kubernetes.Interface
	// nsqClientSet is a clientset for nsq.io API group
	nsqClientSet nsqclientset.Interface

	nsqdLister listernsqv1alpha1.NsqdLister
	nsqdSynced cache.InformerSynced

	nsqdScaleLister listernsqv1alpha1.NsqdScaleLister
	nsqdScaleSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	updationLRUCache *apimachinerycache.LRUExpireCache
}

// NewNsqdScaleController returns a new NsqdScale controller.
func NewNsqdScaleController(opts *options.Options, kubeClientSet kubernetes.Interface,
	// nsqClientSet is a clientset for nsq.io API group
	nsqClientSet nsqclientset.Interface, nsqdScaleInformer v1alpha1.NsqdScaleInformer,
	nsqdInformer v1alpha1.NsqdInformer) *NsqdScaleController {

	// Create event broadcaster
	// Add nsq-controller types to the default Kubernetes Scheme so Events can be
	// logged for nsq-controller types.
	utilruntime.Must(nsqscheme.AddToScheme(scheme.Scheme))
	klog.Info("Creating event broadcaster for nsqdscale controller")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: constant.NsqdScaleControllerName})

	controller := &NsqdScaleController{
		opts:             opts,
		kubeClientSet:    kubeClientSet,
		nsqClientSet:     nsqClientSet,
		nsqdScaleLister:  nsqdScaleInformer.Lister(),
		nsqdScaleSynced:  nsqdScaleInformer.Informer().HasSynced,
		nsqdLister:       nsqdInformer.Lister(),
		nsqdSynced:       nsqdInformer.Informer().HasSynced,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), nsqio.NsqdScaleKind),
		recorder:         recorder,
		updationLRUCache: apimachinerycache.NewLRUExpireCache(opts.NsqdScaleUpdationLRUCacheSize),
	}

	klog.Info("Setting up event handlers for nsqdscale controller")
	// Set up an event handler for when NsqdScale resources change
	nsqdScaleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueNsqdScale,
		UpdateFunc: func(old, new interface{}) {
			newNDS := new.(*nsqv1alpha1.NsqdScale)
			oldNDS := old.(*nsqv1alpha1.NsqdScale)

			if newNDS.ResourceVersion == oldNDS.ResourceVersion {
				// Periodic re-sync will send update events for all known deployments.
				// Two different versions of the same deployment will always have different RVs.
				return
			}

			if reflect.DeepEqual(newNDS.Spec, oldNDS.Spec) {
				// Status update
				var key string
				var err error
				if key, err = cache.MetaNamespaceKeyFunc(new); err != nil {
					utilruntime.HandleError(err)
					return
				}

				if _, exists := controller.updationLRUCache.Get(key); exists {
					klog.V(2).Infof("Ignore status updation for nsqdscale %s/%s within %v",
						newNDS.Namespace, newNDS.Name, opts.NsqdScaleUpdationLRUCacheTTL)
					return
				}

				controller.updationLRUCache.Add(key, struct{}{}, opts.NsqdScaleUpdationLRUCacheTTL)
			}

			controller.handleObject(new)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (ndsc *NsqdScaleController) Run(threads int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer ndsc.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting NsqdScale controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, ndsc.nsqdScaleSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch workers to process NsqScale resources
	for i := 0; i < threads; i++ {
		go wait.Until(ndsc.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down NsqdScale controller")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (ndsc *NsqdScaleController) runWorker() {
	for ndsc.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (ndsc *NsqdScaleController) processNextWorkItem() bool {
	obj, shutdown := ndsc.workqueue.Get()

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
		defer ndsc.workqueue.Done(obj)
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
			ndsc.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// NsqAdmin resource to be synced.
		if err := ndsc.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			ndsc.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %v, requeuing", key, err)
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		ndsc.workqueue.Forget(obj)
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
// converge the two. It then updates the Status block of the NsqdScale resource
// with the current status of the resource.
func (ndsc *NsqdScaleController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the NsqdScale resource with this namespace/name
	nds, err := ndsc.nsqdScaleLister.NsqdScales(namespace).Get(name)
	if err != nil {
		// The NsqdScale resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("nsqdscale '%s' in work queue no longer exists", key))
			return nil
		}

		klog.Errorf("Get nsqdscale %s/%s error: %v", nds.Namespace, nds.Name, err)
		return err
	}

	replica, operation, err := ndsc.computeReplica(nds)
	if err != nil && err != nsqerror.ErrNoNeedToUpdateNsqdReplica {
		klog.Errorf("Compute replica for nsqd %s/%s error: %v", nds.Namespace, nds.Name, err)
		return err
	}

	if err == nsqerror.ErrNoNeedToUpdateNsqdReplica {
		klog.Infof("No need to update nsqd replica for %s/%s", nds.Namespace, nds.Name)
		return nil
	}

	if operation == constant.ScaleUp {
		if nds.Status.LastScaleUpTime.Add(ndsc.opts.NsqdScaleUpSilenceDuration).Before(time.Now()) {
			err = ndsc.updateNsqdReplicas(nds, replica)
			if err != nil {
				klog.Errorf("Update nsqd %s/%s replica to %d error: %v", nds.Namespace, nds.Name, replica, err)
				return err
			}

			err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				// Retrieve the latest version of nsqdscale before attempting update
				// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
				ndsOld, err := ndsc.nsqClientSet.NsqV1alpha1().NsqdScales(nds.Namespace).Get(nds.Name, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("get nsqdscale %s/%s from apiserver error: %v", nds.Namespace, nds.Name, err)
				}

				// NEVER modify objects from the store. It's a read-only, local cache.
				// You can use DeepCopy() to make a deep copy of original object and modify this copy
				// Or create a copy manually for better performance
				ndsCopy := ndsOld.DeepCopy()
				ndsCopy.Status.LastScaleUpTime = metav1.Now()
				// If the CustomResourceSubresources feature gate is not enabled,
				// we must use Update instead of UpdateStatus to update the Status block of the NsqAdmin resource.
				// UpdateStatus will not allow changes to the Spec of the resource,
				// which is ideal for ensuring nothing other than resource status has been updated.
				_, err = ndsc.nsqClientSet.NsqV1alpha1().NsqdScales(nds.Namespace).Update(ndsCopy)
				return err
			})

			return err
		} else {
			klog.Errorf("Ingore scale up for nsqdscale %s/%s. Last scale up time: %v. Scale up silence duration: %v. Wanted replica: %v",
				nds.Namespace, nds.Name, nds.Status.LastScaleUpTime, ndsc.opts.NsqdScaleUpSilenceDuration, replica)
		}
	}

	if operation == constant.ScaleDown {
		if nds.Status.LastScaleDownTime.Add(ndsc.opts.NsqdScaleDownSilenceDuration).Before(time.Now()) {
			err = ndsc.updateNsqdReplicas(nds, replica)
			if err != nil {
				klog.Errorf("Update nsqd %s/%s replica to %d error: %v", nds.Namespace, nds.Name, replica, err)
				return err
			}

			err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				// Retrieve the latest version of nsqdscale before attempting update
				// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
				ndsOld, err := ndsc.nsqClientSet.NsqV1alpha1().NsqdScales(nds.Namespace).Get(nds.Name, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("get nsqdscale %s/%s from apiserver error: %v", nds.Namespace, nds.Name, err)
				}

				// NEVER modify objects from the store. It's a read-only, local cache.
				// You can use DeepCopy() to make a deep copy of original object and modify this copy
				// Or create a copy manually for better performance
				ndsCopy := ndsOld.DeepCopy()
				ndsCopy.Status.LastScaleDownTime = metav1.Now()
				// If the CustomResourceSubresources feature gate is not enabled,
				// we must use Update instead of UpdateStatus to update the Status block of the NsqAdmin resource.
				// UpdateStatus will not allow changes to the Spec of the resource,
				// which is ideal for ensuring nothing other than resource status has been updated.
				_, err = ndsc.nsqClientSet.NsqV1alpha1().NsqdScales(nds.Namespace).Update(ndsCopy)
				return err
			})

			return err
		} else {
			klog.Errorf("Ingore scale down for nsqdscale %s/%s. Last scale down time: %v. Scale down silence duration: %v. Wanted replica: %v",
				nds.Namespace, nds.Name, nds.Status.LastScaleDownTime, ndsc.opts.NsqdScaleDownSilenceDuration, replica)
		}
	}

	return nil
}

func (ndsc *NsqdScaleController) updateNsqdReplicas(nds *nsqv1alpha1.NsqdScale, replicas int32) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// Retrieve the latest version of nsqd before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		ndOld, err := ndsc.nsqClientSet.NsqV1alpha1().Nsqds(nds.Namespace).Get(nds.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get nsqdscale %s/%s from apiserver error: %v", nds.Namespace, nds.Name, err)
		}

		// NEVER modify objects from the store. It's a read-only, local cache.
		// You can use DeepCopy() to make a deep copy of original object and modify this copy
		// Or create a copy manually for better performance
		ndCopy := ndOld.DeepCopy()
		ndCopy.Spec.Replicas = replicas
		// If the CustomResourceSubresources feature gate is not enabled,
		// we must use Update instead of UpdateStatus to update the Status block of the NsqAdmin resource.
		// UpdateStatus will not allow changes to the Spec of the resource,
		// which is ideal for ensuring nothing other than resource status has been updated.
		_, err = ndsc.nsqClientSet.NsqV1alpha1().Nsqds(nds.Namespace).Update(ndCopy)
		return err
	})

	return err
}

// enqueueNsqdScale takes a NsqdScale resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than NsqdScale.
func (ndsc *NsqdScaleController) enqueueNsqdScale(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	ndsc.workqueue.Add(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Nsqd resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that NsqdScale resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (ndsc *NsqdScaleController) handleObject(obj interface{}) {
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
	klog.V(4).Infof("Processing object %v %s/%s", reflect.TypeOf(object), object.GetNamespace(), object.GetName())
	nds, err := ndsc.nsqdScaleLister.NsqdScales(object.GetNamespace()).Get(object.GetName())
	if err != nil {
		klog.Errorf("Get nsqdscale %s/%s error: %v. Ignore check", object.GetNamespace(), object.GetName(), err)
		return
	}

	ndsc.enqueueNsqdScale(nds)
	return
}

// computeReplica computes the replica of nsqd.
func (ndsc *NsqdScaleController) computeReplica(nds *nsqv1alpha1.NsqdScale) (replica int32, operation constant.NsqdScaleOperation, err error) {
	nd, err := ndsc.nsqdLister.Nsqds(nds.Namespace).Get(nds.Name)
	if err != nil {
		klog.Errorf("Get nsqd %s/%s error: %v", nds.Namespace, nds.Name, err)
		return replica, constant.Unknown, err
	}

	now := time.Now()
	var qpsSum int64
	var validItemCount int64
	for nsqdName, qpses := range nds.Status.Metas {
		for _, qps := range qpses {
			if qps.LastUpdateTime.Add(ndsc.opts.NsqdScaleValidDuration).After(now) {
				klog.V(2).Infof("nsqd %s/%s qps data %+v is valid", nds.Namespace, nsqdName, qps)
				validItemCount += 1
				qpsSum += qps.Qps
				continue
			}
			klog.Warningf("nsqd %s/%s qps data %+v is invalid. Last update time is older than %v from now",
				nds.Namespace, nsqdName, qps, ndsc.opts.NsqdScaleValidDuration)
		}
	}

	if validItemCount == 0 {
		klog.Warningf("No valid nsqd qps data counted for nsqdscale %s/%s", nds.Namespace, nds.Name)
		return replica, constant.Unknown, nsqerror.ErrNoNeedToUpdateNsqdReplica
	}

	qpsAvg := int64(math.Ceil(float64(qpsSum) / float64(validItemCount)))
	qpsDiff := qpsAvg - int64(nds.Spec.QpsThreshold)
	klog.V(2).Infof("Counted qps avg: %d, wanted qps threshold: %d, qps diff: %d", qpsAvg, nds.Spec.QpsThreshold, qpsDiff)
	var replicaDiff int32
	if qpsDiff > 0 {
		replicaDiff = int32(math.Ceil(float64(qpsDiff) / float64(nds.Spec.QpsThreshold)))
	} else {
		replicaDiff = int32(math.Ceil(float64(qpsDiff*int64(nd.Spec.Replicas))) / float64(nds.Spec.QpsThreshold))
	}
	if replicaDiff == 0 {
		klog.Infof("No need to update replica for nsqd %s/%s", nds.Namespace, nds.Name)
		return replica, constant.Unknown, nsqerror.ErrNoNeedToUpdateNsqdReplica
	}

	klog.Infof("Update replica by adding %d for nsqd %s/%s. Current replica is %d",
		replicaDiff, nds.Namespace, nds.Name, nd.Spec.Replicas)
	replica = nd.Spec.Replicas + replicaDiff

	if replica > nds.Spec.Maximum {
		klog.Infof("Align to maximum %d for nsqdscale %s/%s",
			nds.Spec.Maximum, nds.Namespace, nds.Name)
		replica = nds.Spec.Maximum
	}

	if replica < nds.Spec.Minimum {
		klog.Infof("Align to minimum %d for nsqdscale %s/%s",
			nds.Spec.Minimum, nds.Namespace, nds.Name)
		replica = nds.Spec.Minimum
	}

	if replicaDiff > 0 {
		return replica, constant.ScaleUp, nil
	}

	return replica, constant.ScaleDown, nil
}
