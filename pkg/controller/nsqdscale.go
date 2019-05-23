///*
//Copyright 2019 The NSQ-Operator Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//*/
//
package controller

//
//import (
//	"fmt"
//	"math"
//	"time"
//
//	"github.com/andyxning/nsq-operator/cmd/nsq-operator/options"
//	"github.com/andyxning/nsq-operator/pkg/apis/nsqio"
//	"github.com/andyxning/nsq-operator/pkg/constant"
//	"github.com/andyxning/nsq-operator/pkg/generated/informers/externalversions/nsqio/v1alpha1"
//	"k8s.io/apimachinery/pkg/api/errors"
//	"k8s.io/apimachinery/pkg/api/resource"
//	"k8s.io/client-go/kubernetes"
//	"k8s.io/client-go/kubernetes/scheme"
//	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
//	"k8s.io/client-go/tools/cache"
//	"k8s.io/client-go/tools/record"
//	"k8s.io/client-go/util/retry"
//	"k8s.io/client-go/util/workqueue"
//	"k8s.io/klog"
//
//	nsqclientset "github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned"
//	corev1 "k8s.io/api/core/v1"
//	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
//	"k8s.io/apimachinery/pkg/util/wait"
//
//	nsqv1alpha1 "github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
//	nsqscheme "github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned/scheme"
//	listernsqv1alpha1 "github.com/andyxning/nsq-operator/pkg/generated/listers/nsqio/v1alpha1"
//	appsv1 "k8s.io/api/apps/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//)
//
//// NsqdScaleController is the controller implementation for NsqdScale resources.
//type NsqdScaleController struct {
//	opts *options.Options
//
//	// kubeClientSet is a standard kubernetes clientset
//	kubeClientSet kubernetes.Interface
//	// nsqClientSet is a clientset for nsq.io API group
//	nsqClientSet nsqclientset.Interface
//
//	nsqdsLister listernsqv1alpha1.NsqdLister
//	nsqdsSynced cache.InformerSynced
//
//	// workqueue is a rate limited work queue. This is used to queue work to be
//	// processed instead of performing it as soon as a change happens. This
//	// means we can ensure we only process a fixed amount of resources at a
//	// time, and makes it easy to ensure we are never processing the same item
//	// simultaneously in two different workers.
//	workqueue workqueue.RateLimitingInterface
//	// recorder is an event recorder for recording Event resources to the
//	// Kubernetes API.
//	recorder record.EventRecorder
//}
//
//// NewNsqdScaleController returns a new NsqdScale controller.
//func NewNsqdScaleController(opts *options.Options, kubeClientSet kubernetes.Interface,
//	// nsqClientSet is a clientset for nsq.io API group
//	nsqClientSet nsqclientset.Interface, nsqInformer v1alpha1.NsqdInformer) *NsqdScaleController {
//
//	// Create event broadcaster
//	// Add nsq-controller types to the default Kubernetes Scheme so Events can be
//	// logged for nsq-controller types.
//	utilruntime.Must(nsqscheme.AddToScheme(scheme.Scheme))
//	klog.Info("Creating event broadcaster for nsqdscale controller")
//	eventBroadcaster := record.NewBroadcaster()
//	eventBroadcaster.StartLogging(klog.Infof)
//	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
//	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: constant.NsqdScaleControllerName})
//
//	controller := &NsqdScaleController{
//		opts:          opts,
//		kubeClientSet: kubeClientSet,
//		nsqClientSet:  nsqClientSet,
//		nsqdsLister:   nsqInformer.Lister(),
//		nsqdsSynced:   nsqInformer.Informer().HasSynced,
//		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), nsqio.NsqdScaleKind),
//		recorder:      recorder,
//	}
//
//	klog.Info("Setting up event handlers for nsqdscale controller")
//	// Set up an event handler for when NsqdScale resources change
//	nsqInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
//		AddFunc: controller.enqueueNsqdScale,
//		UpdateFunc: func(old, new interface{}) {
//			controller.enqueueNsqdScale(new)
//		},
//	})
//
//	return controller
//}
//
//// Run will set up the event handlers for types we are interested in, as well
//// as syncing informer caches and starting workers. It will block until stopCh
//// is closed, at which point it will shutdown the workqueue and wait for
//// workers to finish processing their current work items.
//func (ndsc *NsqdScaleController) Run(threads int, stopCh <-chan struct{}) error {
//	defer utilruntime.HandleCrash()
//	defer ndsc.workqueue.ShutDown()
//
//	// Start the informer factories to begin populating the informer caches
//	klog.Info("Starting NsqdScale controller")
//
//	// Wait for the caches to be synced before starting workers
//	klog.Info("Waiting for informer caches to sync")
//	if ok := cache.WaitForCacheSync(stopCh, ndsc.nsqdsSynced); !ok {
//		return fmt.Errorf("failed to wait for caches to sync")
//	}
//
//	klog.Info("Starting workers")
//	// Launch workers to process NsqScale resources
//	for i := 0; i < threads; i++ {
//		go wait.Until(ndsc.runWorker, time.Second, stopCh)
//	}
//
//	klog.Info("Started workers")
//	<-stopCh
//	klog.Info("Shutting down NsqdScale controller")
//
//	return nil
//}
//
//// runWorker is a long-running function that will continually call the
//// processNextWorkItem function in order to read and process a message on the
//// workqueue.
//func (ndsc *NsqdScaleController) runWorker() {
//	for ndsc.processNextWorkItem() {
//	}
//}
//
//// processNextWorkItem will read a single work item off the workqueue and
//// attempt to process it, by calling the syncHandler.
//func (ndsc *NsqdScaleController) processNextWorkItem() bool {
//	obj, shutdown := ndsc.workqueue.Get()
//
//	if shutdown {
//		return false
//	}
//
//	// We wrap this block in a func so we can defer c.workqueue.Done.
//	err := func(obj interface{}) error {
//		// We call Done here so the workqueue knows we have finished
//		// processing this item. We also must remember to call Forget if we
//		// do not want this work item being re-queued. For example, we do
//		// not call Forget if a transient error occurs, instead the item is
//		// put back on the workqueue and attempted again after a back-off
//		// period.
//		defer ndsc.workqueue.Done(obj)
//		var key string
//		var ok bool
//		// We expect strings to come off the workqueue. These are of the
//		// form namespace/name. We do this as the delayed nature of the
//		// workqueue means the items in the informer cache may actually be
//		// more up to date that when the item was initially put onto the
//		// workqueue.
//		if key, ok = obj.(string); !ok {
//			// As the item in the workqueue is actually invalid, we call
//			// Forget here else we'd go into a loop of attempting to
//			// process a work item that is invalid.
//			ndsc.workqueue.Forget(obj)
//			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
//			return nil
//		}
//		// Run the syncHandler, passing it the namespace/name string of the
//		// NsqAdmin resource to be synced.
//		if err := ndsc.syncHandler(key); err != nil {
//			// Put the item back on the workqueue to handle any transient errors.
//			ndsc.workqueue.AddRateLimited(key)
//			return fmt.Errorf("error syncing '%s': %v, requeuing", key, err)
//		}
//		// Finally, if no error occurs we Forget this item so it does not
//		// get queued again until another change happens.
//		ndsc.workqueue.Forget(obj)
//		klog.V(2).Infof("Successfully synced '%s'", key)
//		return nil
//	}(obj)
//
//	if err != nil {
//		utilruntime.HandleError(err)
//		return true
//	}
//
//	return true
//}
//
//// syncHandler compares the actual state with the desired, and attempts to
//// converge the two. It then updates the Status block of the NsqdScale resource
//// with the current status of the resource.
//func (ndsc *NsqdScaleController) syncHandler(key string) error {
//	// Convert the namespace/name string into a distinct namespace and name
//	namespace, name, err := cache.SplitMetaNamespaceKey(key)
//	if err != nil {
//		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
//		return nil
//	}
//
//	// Get the NsqdScale resource with this namespace/name
//	nd, err := ndsc.nsqdsLister.Nsqds(namespace).Get(name)
//	if err != nil {
//		// The Nsqd resource may no longer exist, in which case we stop
//		// processing.
//		if errors.IsNotFound(err) {
//			utilruntime.HandleError(fmt.Errorf("nsqd '%s' in work queue no longer exists", key))
//			return nil
//		}
//
//		klog.Errorf("Get nsqd %s/%s error: %v", nd.Namespace, nd.Name, err)
//		return err
//	}
//
//	// Finally, we update the status block of the Nsqd resource to reflect the
//	// current state of the world
//	err = ndsc.updateNsqdReplicas(nd)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (ndsc *NsqdScaleController) updateNsqdReplicas(nd *nsqv1alpha1.Nsqd, statefulSet *appsv1.StatefulSet) error {
//	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
//		// Retrieve the latest version of statefulset before attempting update
//		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
//		ndOld, err := ndc.nsqClientSet.NsqV1alpha1().Nsqds(nd.Namespace).Get(nd.Name, metav1.GetOptions{})
//		if err != nil {
//			return fmt.Errorf("get nsqd %s/%s from apiserver error: %v", nd.Namespace, nd.Name, err)
//		}
//		// NEVER modify objects from the store. It's a read-only, local cache.
//		// You can use DeepCopy() to make a deep copy of original object and modify this copy
//		// Or create a copy manually for better performance
//		ndCopy := ndOld.DeepCopy()
//		ndCopy.Status.AvailableReplicas = nd.Spec.Replicas
//		ndCopy.Status.MemoryOverSalePercent = nd.Spec.MemoryOverSalePercent
//		ndCopy.Status.MessageAvgSize = nd.Spec.MessageAvgSize
//		ndCopy.Status.MemoryQueueSize = nd.Spec.MemoryQueueSize
//		ndCopy.Status.ChannelCount = nd.Spec.ChannelCount
//		// If the CustomResourceSubresources feature gate is not enabled,
//		// we must use Update instead of UpdateStatus to update the Status block of the NsqAdmin resource.
//		// UpdateStatus will not allow changes to the Spec of the resource,
//		// which is ideal for ensuring nothing other than resource status has been updated.
//		_, err = ndc.nsqClientSet.NsqV1alpha1().Nsqds(nd.Namespace).Update(ndCopy)
//		return err
//	})
//
//	return err
//}
//
//// enqueueNsqdScale takes a NsqdScale resource and converts it into a namespace/name
//// string which is then put onto the work queue. This method should *not* be
//// passed resources of any type other than NsqdScale.
//func (ndsc *NsqdScaleController) enqueueNsqdScale(obj interface{}) {
//	var key string
//	var err error
//	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
//		utilruntime.HandleError(err)
//		return
//	}
//	ndsc.workqueue.Add(key)
//}
//
//// handleObject will take any resource implementing metav1.Object and attempt
//// to find the Nsqd resource that 'owns' it. It does this by looking at the
//// objects metadata.ownerReferences field for an appropriate OwnerReference.
//// It then enqueues that NsqdScale resource to be processed. If the object does not
//// have an appropriate OwnerReference, it will simply be skipped.
//func (ndsc *NsqdScaleController) handleObject(obj interface{}) {
//	var object metav1.Object
//	var ok bool
//	if object, ok = obj.(metav1.Object); !ok {
//		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
//		if !ok {
//			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
//			return
//		}
//		object, ok = tombstone.Obj.(metav1.Object)
//		if !ok {
//			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
//			return
//		}
//		klog.V(4).Infof("Recovered deleted object '%s/%s' from tombstone", object.GetNamespace(), object.GetName())
//	}
//	klog.V(4).Infof("Processing object %s/%s", object.GetNamespace(), object.GetName())
//	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
//		// If this object is not owned by a Nsqd, we should not do anything more
//		// with it.
//		if ownerRef.Kind != nsqio.NsqdKind {
//			return
//		}
//
//		nd, err := ndsc.nsqdsLister.Nsqds(object.GetNamespace()).Get(ownerRef.Name)
//		if err != nil {
//			klog.V(4).Infof("Ignoring orphaned object '%s' of nsqd '%s'", object.GetSelfLink(), ownerRef.Name)
//			return
//		}
//
//		ndsc.enqueueNsqd(nd)
//		return
//	}
//}
//
//// computeNsqdScale updates the OwnerReferences of nsqd configmap to nsqd.
//func (ndc *NsqdController) computeNsqdMemoryResource(nd *nsqv1alpha1.Nsqd) (request resource.Quantity, limit resource.Quantity) {
//	count := int64(nd.Spec.ChannelCount + 1) // 1 for topic itself
//	singleMemUsage := int64(nd.Spec.MessageAvgSize) * int64(nd.Spec.MemoryQueueSize)
//
//	standardTotalMem := count * singleMemUsage
//	AdjustedTotalMem := float64(standardTotalMem) * (1 + float64(nd.Spec.MemoryOverSalePercent)/100.0)
//
//	request = resource.MustParse(fmt.Sprintf("%vMi", int64(math.Ceil(AdjustedTotalMem/1024.0/1024.0))))
//	limit = request
//
//	return request, limit
//}
