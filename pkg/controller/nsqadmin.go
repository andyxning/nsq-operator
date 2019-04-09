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
	"time"

	"github.com/andyxning/nsq-operator/pkg/apis/nsqio"
	"github.com/andyxning/nsq-operator/pkg/common"
	"github.com/andyxning/nsq-operator/pkg/constant"
	"github.com/andyxning/nsq-operator/pkg/generated/informers/externalversions/nsqio/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
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

// NsqAdminController is the controller implementation for NsqAdmin resources.
type NsqAdminController struct {
	// kubeClientSet is a standard kubernetes clientset
	kubeClientSet kubernetes.Interface
	// nsqClientSet is a clientset for nsq.io API group
	nsqClientSet nsqclientset.Interface

	deploymentsLister listerappsv1.DeploymentLister
	deploymentsSynced cache.InformerSynced

	configmapsLister listercorev1.ConfigMapLister
	configmapsSynced cache.InformerSynced

	nsqAdminsLister listernsqv1alpha1.NsqAdminLister
	nsqAdminsSynced cache.InformerSynced

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

// NewNsqAdminController returns a new NsqAdmin controller.
func NewNsqAdminController(kubeClientSet kubernetes.Interface,
	// nsqClientSet is a clientset for nsq.io API group
	nsqClientSet nsqclientset.Interface,
	deploymentInformer informersappsv1.DeploymentInformer,
	configmapInformer informerscorev1.ConfigMapInformer,
	nsqInformer v1alpha1.NsqAdminInformer) *NsqAdminController {

	// Create event broadcaster
	// Add nsq-controller types to the default Kubernetes Scheme so Events can be
	// logged for nsq-controller types.
	utilruntime.Must(nsqscheme.AddToScheme(scheme.Scheme))
	klog.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: constant.NsqAdminControllerName})

	controller := &NsqAdminController{
		kubeClientSet:     kubeClientSet,
		nsqClientSet:      nsqClientSet,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		configmapsLister:  configmapInformer.Lister(),
		configmapsSynced:  configmapInformer.Informer().HasSynced,
		nsqAdminsLister:   nsqInformer.Lister(),
		nsqAdminsSynced:   nsqInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), nsqio.NsqAdminKind),
		recorder:          recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when NsqAdmin resources change
	nsqInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueNsqAdmin,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueNsqAdmin(new)
		},
	})
	// Set up an event handler for when Deployment resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a NsqAdmin resource will enqueue that NsqAdmin resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
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
		DeleteFunc: controller.handleObject,
	})

	return controller

}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (nac *NsqAdminController) Run(threads int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer nac.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting NsqAdmin controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, nac.deploymentsSynced, nac.nsqAdminsSynced, nac.configmapsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch workers to process NsqAdmin resources
	for i := 0; i < threads; i++ {
		go wait.Until(nac.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down NsqAdmin controller")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (nac *NsqAdminController) runWorker() {
	for nac.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (nac *NsqAdminController) processNextWorkItem() bool {
	obj, shutdown := nac.workqueue.Get()

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
		defer nac.workqueue.Done(obj)
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
			nac.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// NsqAdmin resource to be synced.
		if err := nac.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			nac.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %v, requeuing", key, err)
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		nac.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the NsqAdmin resource
// with the current status of the resource.
func (nac *NsqAdminController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the NsqAdmin resource with this namespace/name
	na, err := nac.nsqAdminsLister.NsqAdmins(namespace).Get(name)
	if err != nil {
		// The NsqAdmin resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("nsqadmin '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	// Get the configmap with the name derived from nsqadmin cluster name
	configmap, err := nac.configmapsLister.ConfigMaps(na.Namespace).Get(common.NsqAdminConfigMapName(na.Name))
	// If the resource doesn't exist, we'll return cause that without the configmap,
	// nsqadmin can not assemble the command line arguments to start.
	if err != nil {
		return err
	}

	configmapHash, err := common.Hash(configmap.Data)
	if err != nil {
		klog.Errorf("Hash configmap data for nsqadmin %v error: %v", na.Name, err)
		return err
	}

	deploymentName := common.NsqAdminDeploymentName(na.Name)
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: deployment name must be non empty", key))
		return nil
	}

	deployment, err := nac.deploymentsLister.Deployments(na.Namespace).Get(deploymentName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		klog.Infof("Deployment for nsqadmin %v does not exist. Create it", na.Name)
		deployment, err = nac.kubeClientSet.AppsV1().Deployments(na.Namespace).Create(nac.newDeployment(na, string(configmapHash)))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If the Deployment is not controlled by this NsqAdmin resource, we should log
	// a warning to the event recorder and return
	if !metav1.IsControlledBy(deployment, na) {
		deployment.GetCreationTimestamp()
		msg := fmt.Sprintf(constant.MessageResourceExists, deployment.Name)
		nac.recorder.Event(na, corev1.EventTypeWarning, nsqerror.ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	klog.V(6).Infof("New configmap hash: %v", string(configmapHash))
	klog.V(6).Infof("Old configmap hash: %v", deployment.Spec.Template.Annotations[constant.NsqConfigMapAnnotationKey])
	klog.V(6).Infof("New configmap data: %v", configmap.Data)
	if deployment.Spec.Template.Annotations[constant.NsqConfigMapAnnotationKey] != string(configmapHash) {
		klog.Infof("New configmap detected. New config: %v", configmap.Data)
		deploymentCopy := deployment.DeepCopy()
		deploymentCopy.Spec.Template.Annotations = map[string]string{
			constant.NsqConfigMapAnnotationKey: string(configmapHash),
		}
		deploymentNew, err := nac.kubeClientSet.AppsV1().Deployments(na.Namespace).Update(deploymentCopy)

		// If an error occurs during Update, we'll requeue the item so we can
		// attempt processing again later. THis could have been caused by a
		// temporary network failure, or any other transient reason.
		//
		// If no error occurs, just return to give kubernetes some time to make
		// adjustment according to the new spec.
		klog.V(6).Infof("Update deployment %v under configmap change error: %v", deploymentCopy.Name, err)
		klog.V(6).Infof("New deployment %v annotation under configmap change: %v", deploymentCopy.Name, []byte(deploymentNew.Spec.Template.Annotations[constant.NsqConfigMapAnnotationKey]))
		return err
	}

	// If this number of the replicas on the NsqAdmin resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	if na.Spec.Replicas != nil && *na.Spec.Replicas != *deployment.Spec.Replicas {
		deploymentCopy := deployment.DeepCopy()
		deploymentCopy.Spec.Replicas = na.Spec.Replicas
		klog.Infof("NsqAdmin %s replicas: %d, deployment replicas: %d", name, *na.Spec.Replicas, *deployment.Spec.Replicas)
		deployment, err = nac.kubeClientSet.AppsV1().Deployments(na.Namespace).Update(deploymentCopy)
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the NsqAdmin resource to reflect the
	// current state of the world
	err = nac.updateNsqAdminStatus(na, deployment)
	if err != nil {
		return err
	}

	return nil
}

func (nac *NsqAdminController) updateNsqAdminStatus(na *nsqv1alpha1.NsqAdmin, deployment *appsv1.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	naCopy := na.DeepCopy()
	naCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the NsqAdmin resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := nac.nsqClientSet.NsqV1alpha1().NsqAdmins(na.Namespace).Update(naCopy)
	return err
}

// enqueueNsqAdmin takes a NsqAdmin resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than NsqAdmin.
func (nac *NsqAdminController) enqueueNsqAdmin(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	nac.workqueue.Add(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the NsqAdmin resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that NsqAdmin resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *NsqAdminController) handleObject(obj interface{}) {
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
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object(%v): %s", object.GetSelfLink(), object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a NsqAdmin, we should not do anything more
		// with it.
		if ownerRef.Kind != nsqio.NsqAdminKind {
			return
		}

		nd, err := c.nsqAdminsLister.NsqAdmins(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueNsqAdmin(nd)
		return
	}
}

// newDeployment creates a new Deployment for a NsqAdmin resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the NsqAdmin resource that 'owns' it.
func (nac *NsqAdminController) newDeployment(na *nsqv1alpha1.NsqAdmin, cfs string) *appsv1.Deployment {
	labels := map[string]string{
		"cluster": na.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.NsqAdminDeploymentName(na.Name),
			Namespace: na.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(na, schema.GroupVersionKind{
					Group:   nsqv1alpha1.SchemeGroupVersion.Group,
					Version: nsqv1alpha1.SchemeGroupVersion.Version,
					Kind:    nsqio.NsqAdminKind,
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: na.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						constant.NsqConfigMapAnnotationKey: cfs,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  na.Name,
							Image: na.Spec.Image,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      common.NsqAdminConfigMapName(na.Name),
									MountPath: constant.NsqConfigMapMountPath,
								},
							},
							ImagePullPolicy: corev1.PullAlways,
						},
					},
					Volumes: []corev1.Volume{{
						Name: common.NsqAdminConfigMapName(na.Name),
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: common.NsqAdminConfigMapName(na.Name),
								},
							},
						},
					},
					},
				},
			},
		},
	}
}