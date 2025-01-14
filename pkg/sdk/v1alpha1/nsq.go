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

package v1alpha1

import (
	"context"
	"fmt"

	"github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
	pkgcommon "github.com/andyxning/nsq-operator/pkg/common"
	"github.com/andyxning/nsq-operator/pkg/constant"
	"github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned"
	"github.com/andyxning/nsq-operator/pkg/sdk/v1alpha1/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

func CreateCluster(kubeClient *kubernetes.Clientset, nsqClient *versioned.Clientset, ncr *types.NsqCreateRequest) error {
	// Create nsqlookupd configmap
	nsqLookupdCM := ncr.AssembleNsqLookupdConfigMap()
	klog.Infof("Create nsqlookupd configmap %s/%s", nsqLookupdCM.Namespace, nsqLookupdCM.Name)
	_, err := kubeClient.CoreV1().ConfigMaps(ncr.NsqdConfig.Namespace).Create(nsqLookupdCM)
	if errors.IsAlreadyExists(err) {
		klog.Infof("Nsqlookupd configmap %s/%s exists. Update it", nsqLookupdCM.Namespace, nsqLookupdCM.Name)
		var old *corev1.ConfigMap
		if klog.V(4) {
			old, err = kubeClient.CoreV1().ConfigMaps(ncr.NsqdConfig.Namespace).Get(nsqLookupdCM.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get old nsqlookupd configmap %s/%s error: %v", nsqLookupdCM.Namespace, nsqLookupdCM.Name, err)
				return err
			}

			klog.V(4).Infof("Old nsqlookupd configmap %s/%s: %#v", nsqLookupdCM.Namespace, nsqLookupdCM.Name, *old)
		}

		oldCopy := old.DeepCopy()
		oldCopy.Data = nsqLookupdCM.Data
		klog.V(4).Infof("New nsqlookupd configmap %s/%s: %#v", nsqLookupdCM.Namespace, nsqLookupdCM.Name, *oldCopy)
		_, err = kubeClient.CoreV1().ConfigMaps(ncr.NsqdConfig.Namespace).Update(oldCopy)
	}

	if err != nil {
		klog.Errorf("Create nsqlookupd configmap error: %v", err)
		return err
	}

	// Create nsqlookupd
	nsqLookupd := ncr.AssembleNsqLookupd()
	klog.Infof("Create nsqlookupd %s/%s", nsqLookupd.Namespace, nsqLookupd.Name)
	_, err = nsqClient.NsqV1alpha1().NsqLookupds(ncr.NsqdConfig.Namespace).Create(nsqLookupd)
	if errors.IsAlreadyExists(err) {
		klog.Infof("Nsqlookupd %s/%s exists. Update it", nsqLookupd.Namespace, nsqLookupd.Name)
		var old *v1alpha1.NsqLookupd
		if klog.V(4) {
			old, err = nsqClient.NsqV1alpha1().NsqLookupds(ncr.NsqdConfig.Namespace).Get(nsqLookupd.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get old nsqlookupd %s/%s error: %v", nsqLookupd.Namespace, nsqLookupd.Name, err)
				return err
			}

			klog.V(4).Infof("Old nsqlookupd %s/%s: %#v", nsqLookupd.Namespace, nsqLookupd.Name, *old)
		}

		oldCopy := old.DeepCopy()
		oldCopy.Spec = nsqLookupd.Spec
		klog.V(4).Infof("New nsqlookupd %s/%s: %#v", nsqLookupd.Namespace, nsqLookupd.Name, *oldCopy)
		_, err = nsqClient.NsqV1alpha1().NsqLookupds(ncr.NsqdConfig.Namespace).Update(oldCopy)
	}

	if err != nil {
		klog.Errorf("Create nsqlookupd error: %v", err)
		return err
	}

	// Wait for nsqlookupd to reach its spec
	klog.V(2).Infof("Waiting for nsqlookupd %s/%s reach its spec", nsqLookupd.Namespace, nsqLookupd.Name)
	ctx, cancel := context.WithTimeout(context.Background(), *ncr.NsqdConfig.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqLookupd, err := nsqClient.NsqV1alpha1().NsqLookupds(ncr.NsqdConfig.Namespace).Get(ncr.NsqdConfig.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqlookupd error: %v", err)
			return
		}

		if nsqLookupd.Spec.Replicas != nsqLookupd.Status.AvailableReplicas {
			klog.V(2).Infof("Spec and status does not match for nsqlookupd %s/%s. Spec: %#v, Status: %#v",
				ncr.NsqdConfig.Namespace, ncr.NsqdConfig.Name, nsqLookupd.Spec, nsqLookupd.Status)
			return
		}

		// Spec and status matches
		cancel()
	}, constant.NsqLookupdStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqlookupd resource to reach its spec")
		return ctx.Err()
	}

	// Check for existence of nsqadmin configmap
	klog.V(2).Infof("Waiting for nsqadmin configmap %s/%s reach its spec", ncr.NsqdConfig.Namespace, pkgcommon.NsqAdminConfigMapName(ncr.NsqdConfig.Name))
	ctx, cancel = context.WithTimeout(context.Background(), *ncr.NsqdConfig.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqadminCM, err := kubeClient.CoreV1().ConfigMaps(ncr.NsqdConfig.Namespace).Get(pkgcommon.NsqAdminConfigMapName(ncr.NsqdConfig.Name), metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqadmin configmap error: %v", err)
			return
		}

		if nsqadminCM.Data != nil && len(nsqadminCM.Data) != 0 {
			// nsqadmin configmap exists
			cancel()
		}
	}, constant.ConfigMapStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqadmin configmap to be ready")
		return ctx.Err()
	}

	// Create nsqadmin
	nsqadmin := ncr.AssembleNsqAdmin()
	klog.Infof("Create nsqadmin %s/%s", nsqadmin.Namespace, nsqadmin.Name)
	_, err = nsqClient.NsqV1alpha1().NsqAdmins(ncr.NsqdConfig.Namespace).Create(nsqadmin)
	if errors.IsAlreadyExists(err) {
		klog.Infof("Nsqadmin %s/%s exists. Update it", nsqadmin.Namespace, nsqadmin.Name)
		var old *v1alpha1.NsqAdmin
		if klog.V(4) {
			old, err = nsqClient.NsqV1alpha1().NsqAdmins(ncr.NsqdConfig.Namespace).Get(nsqadmin.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get old nsqadmin %s/%s error: %v", nsqadmin.Namespace, nsqadmin.Name, err)
				return err
			}

			klog.V(4).Infof("Old nsqadmin %s/%s: %#v", nsqadmin.Namespace, nsqadmin.Name, *old)
		}

		oldCopy := old.DeepCopy()
		oldCopy.Spec = nsqadmin.Spec
		klog.V(4).Infof("New nsqadmin %s/%s: %#v", nsqadmin.Namespace, nsqadmin.Name, *oldCopy)
		_, err = nsqClient.NsqV1alpha1().NsqAdmins(ncr.NsqdConfig.Namespace).Update(oldCopy)
	}

	if err != nil {
		klog.Errorf("Create nsqadmin error: %v", err)
		return err
	}

	// Wait for nsqadmin to reach its spec
	klog.V(2).Infof("Waiting for nsqadmin %s/%s reach its spec", nsqadmin.Namespace, nsqadmin.Name)
	ctx, cancel = context.WithTimeout(context.Background(), *ncr.NsqdConfig.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqAdmin, err := nsqClient.NsqV1alpha1().NsqAdmins(ncr.NsqdConfig.Namespace).Get(ncr.NsqdConfig.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqadmin error: %v", err)
			return
		}

		if nsqAdmin.Spec.Replicas != nsqAdmin.Status.AvailableReplicas {
			klog.V(2).Infof("Spec and status does not match for nsqadmin %s/%s. Spec: %#v, Status: %#v",
				ncr.NsqdConfig.Namespace, ncr.NsqdConfig.Name, nsqAdmin.Spec, nsqAdmin.Status)
			return
		}

		// Spec and status matches
		cancel()
	}, constant.NsqAdminStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqadmin resource to reach its spec")
		return ctx.Err()
	}

	// Create nsqd configmap
	labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": pkgcommon.NsqLookupdDeploymentName(ncr.NsqdConfig.Name)}}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		klog.Errorf("Gen label selector for nsqlookupd %s/%s pods error: %v", ncr.NsqdConfig.Namespace, ncr.NsqdConfig.Name, err)
		return err
	}

	podList, err := kubeClient.CoreV1().Pods(ncr.NsqdConfig.Namespace).List(metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		klog.Errorf("List nsqlookupd %s/%s pods error: %v", ncr.NsqdConfig.Namespace, ncr.NsqdConfig.Name, err)
		return err
	}

	var addresses []string
	for _, pod := range podList.Items {
		addresses = append(addresses, fmt.Sprintf("%s:%v", pod.Status.PodIP, constant.NsqLookupdTcpPort))
	}

	nsqdCM := ncr.NsqdConfig.AssembleNsqdConfigMap(addresses)
	klog.Infof("Create nsqd configmap %s/%s", nsqdCM.Namespace, nsqdCM.Name)
	_, err = kubeClient.CoreV1().ConfigMaps(ncr.NsqdConfig.Namespace).Create(nsqdCM)
	if errors.IsAlreadyExists(err) {
		klog.Infof("Nsqd configmap %s/%s exists. Update it", nsqdCM.Namespace, nsqdCM.Name)
		var old *corev1.ConfigMap
		if klog.V(4) {
			old, err = kubeClient.CoreV1().ConfigMaps(ncr.NsqdConfig.Namespace).Get(nsqdCM.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get old nsqd configmap %s/%s error: %v", nsqdCM.Namespace, nsqdCM.Name, err)
				return err
			}

			klog.V(4).Infof("Old nsqd configmap %s/%s: %#v", nsqdCM.Namespace, nsqdCM.Name, *old)
		}

		oldCopy := old.DeepCopy()
		oldCopy.Data = nsqdCM.Data
		klog.V(4).Infof("New nsqd configmap %s/%s: %#v", nsqdCM.Namespace, nsqdCM.Name, *oldCopy)
		_, err = kubeClient.CoreV1().ConfigMaps(ncr.NsqdConfig.Namespace).Update(oldCopy)
	}

	if err != nil {
		klog.Errorf("Create nsqd configmap error: %v", err)
		return err
	}

	// Create nsqd
	nsqd := ncr.AssembleNsqd()
	klog.Infof("Create nsqd %s/%s", nsqd.Namespace, nsqd.Name)
	_, err = nsqClient.NsqV1alpha1().Nsqds(ncr.NsqdConfig.Namespace).Create(nsqd)
	if errors.IsAlreadyExists(err) {
		klog.Infof("Nsqd %s/%s exists. Update it", nsqd.Namespace, nsqd.Name)
		var old *v1alpha1.Nsqd
		if klog.V(4) {
			old, err = nsqClient.NsqV1alpha1().Nsqds(ncr.NsqdConfig.Namespace).Get(nsqd.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get old nsqd %s/%s error: %v", nsqd.Namespace, nsqd.Name, err)
				return err
			}

			klog.V(4).Infof("Old nsqd %s/%s: %#v", nsqd.Namespace, nsqd.Name, *old)
		}

		oldCopy := old.DeepCopy()
		oldCopy.Spec = nsqd.Spec
		klog.V(4).Infof("New nsqd %s/%s: %#v", nsqd.Namespace, nsqd.Name, *oldCopy)
		_, err = nsqClient.NsqV1alpha1().Nsqds(ncr.NsqdConfig.Namespace).Update(oldCopy)
	}

	if err != nil {
		klog.Errorf("Create nsqd error: %v", err)
		return err
	}

	// Wait for nsqd to reach its spec
	klog.V(2).Infof("Waiting for nsqd %s/%s reach its spec", nsqd.Namespace, nsqd.Name)
	ctx, cancel = context.WithTimeout(context.Background(), *ncr.NsqdConfig.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqd, err := nsqClient.NsqV1alpha1().Nsqds(ncr.NsqdConfig.Namespace).Get(ncr.NsqdConfig.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqd error: %v", err)
			return
		}

		if nsqd.Spec.Replicas != nsqd.Status.AvailableReplicas {
			klog.V(2).Infof("Spec and status does not match for nsqd %s/%s. Spec: %#v, Status: %#v",
				ncr.NsqdConfig.Namespace, ncr.NsqdConfig.Name, nsqd.Spec, nsqd.Status)
			return
		}

		// Spec and status matches
		cancel()
	}, constant.NsqdStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqd resource to reach its spec")
		return ctx.Err()
	}

	// Create nsqdscale
	nsqdScale := ncr.AssembleNsqdScale()
	klog.Infof("Create nsqdscale %s/%s: %+v", nsqdScale.Namespace, nsqdScale.Name, nsqdScale)
	_, err = nsqClient.NsqV1alpha1().NsqdScales(nsqdScale.Namespace).Create(nsqdScale)
	if errors.IsAlreadyExists(err) {
		klog.Infof("Nsqdscale %s/%s exists. Update it", nsqdScale.Namespace, nsqdScale.Name)
		var old *v1alpha1.NsqdScale
		old, err = nsqClient.NsqV1alpha1().NsqdScales(nsqdScale.Namespace).Get(nsqdScale.Name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get old nsqdscale %s/%s error: %v", nsqdScale.Namespace, nsqdScale.Name, err)
			return err
		}
		klog.V(4).Infof("Old nsqdscale %s/%s: %#v", nsqdScale.Namespace, nsqdScale.Name, *old)

		oldCopy := old.DeepCopy()
		oldCopy.Spec = nsqdScale.Spec
		klog.V(4).Infof("New nsqdscale %s/%s: %#v", nsqdScale.Namespace, nsqdScale.Name, *oldCopy)
		_, err = nsqClient.NsqV1alpha1().NsqdScales(nsqdScale.Namespace).Update(oldCopy)
	}

	if err != nil {
		klog.Errorf("Create nsqdscale error: %v", err)
		return err
	}

	return nil
}

func DeleteCluster(kubeClient *kubernetes.Clientset, nsqClient *versioned.Clientset, ndr *types.NsqDeleteRequest) error {
	err := nsqClient.NsqV1alpha1().NsqLookupds(ndr.Namespace).Delete(ndr.Name, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Warningf("Delete nsqlookupd %s/%s error: %v", ndr.Namespace, ndr.Name, err)
		return err
	}

	klog.Infof("Waiting for nsqlookupd %s/%s to be deleted", ndr.Namespace, ndr.Name)
	ctx, cancel := context.WithTimeout(context.Background(), *ndr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": pkgcommon.NsqLookupdDeploymentName(ndr.Name)}}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			klog.Warningf("Assemble nsqlookupd %s/%s selector error: %v", ndr.Namespace, ndr.Name, err)
			return
		}
		pl, err := kubeClient.CoreV1().Pods(ndr.Namespace).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})

		if err != nil {
			klog.Warningf("List nsqlookupd %s/%s pods error: %v", ndr.Namespace, ndr.Name, err)
			return
		}

		if len(pl.Items) != 0 {
			klog.Infof("There are still %d nsqlookupd for %s/%s", len(pl.Items), ndr.Namespace, ndr.Name)
			return
		}

		cancel()
		klog.Infof("Delete nsqlookupd %s/%s success", ndr.Namespace, ndr.Name)
	}, constant.NsqLookupdStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Delete nsqlookupd %s/%s error: %v", ndr.Namespace, ndr.Name, ctx.Err())
		return ctx.Err()
	}

	err = nsqClient.NsqV1alpha1().NsqAdmins(ndr.Namespace).Delete(ndr.Name, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("Delete nsqadmin %s/%s error: %v", ndr.Namespace, ndr.Name, err)
		return err
	}

	klog.Infof("Waiting for nsqadmin %s/%s to be deleted", ndr.Namespace, ndr.Name)
	ctx, cancel = context.WithTimeout(context.Background(), *ndr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": pkgcommon.NsqAdminDeploymentName(ndr.Name)}}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			klog.Warningf("Assemble nsqadmin %s/%s selector error: %v", ndr.Namespace, ndr.Name, err)
			return
		}
		pl, err := kubeClient.CoreV1().Pods(ndr.Namespace).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})

		if err != nil {
			klog.Warningf("List nsqadmin %s/%s pods error: %v", ndr.Namespace, ndr.Name, err)
			return
		}

		if len(pl.Items) != 0 {
			klog.Infof("There are still %d nsqadmin for %s/%s", len(pl.Items), ndr.Namespace, ndr.Name)
			return
		}

		cancel()
		klog.Infof("Delete nsqadmin %s/%s success", ndr.Namespace, ndr.Name)
	}, constant.NsqAdminStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Delete nsqadmin %s/%s error: %v", ndr.Namespace, ndr.Name, ctx.Err())
		return ctx.Err()
	}

	err = nsqClient.NsqV1alpha1().Nsqds(ndr.Namespace).Delete(ndr.Name, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("Delete nsqd %s/%s error: %v", ndr.Namespace, ndr.Name, err)
		return err
	}

	klog.Infof("Waiting for nsqd %s/%s to be deleted", ndr.Namespace, ndr.Name)
	ctx, cancel = context.WithTimeout(context.Background(), *ndr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": pkgcommon.NsqdStatefulSetName(ndr.Name)}}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			klog.Warningf("Assemble nsqd %s/%s selector error: %v", ndr.Namespace, ndr.Name, err)
			return
		}
		pl, err := kubeClient.CoreV1().Pods(ndr.Namespace).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})

		if err != nil {
			klog.Warningf("List nsqd %s/%s pods error: %v", ndr.Namespace, ndr.Name, err)
			return
		}

		if len(pl.Items) != 0 {
			klog.Infof("There are still %d nsqd for %s/%s", len(pl.Items), ndr.Namespace, ndr.Name)
			return
		}

		cancel()
		klog.Infof("Delete nsqd %s/%s success", ndr.Namespace, ndr.Name)
	}, constant.NsqdStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Delete nsqd %s/%s error: %v", ndr.Namespace, ndr.Name, ctx.Err())
		return ctx.Err()
	}

	err = kubeClient.CoreV1().ConfigMaps(ndr.Namespace).Delete(pkgcommon.NsqLookupdConfigMapName(ndr.Name), &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("Delete nsqlookupd configmap %s/%s error: %v", ndr.Namespace, pkgcommon.NsqLookupdConfigMapName(ndr.Name), err)
		return err
	}

	klog.Infof("Waiting for nsqlookupd configmap %s/%s to be deleted", ndr.Namespace, pkgcommon.NsqLookupdConfigMapName(ndr.Name))
	ctx, cancel = context.WithTimeout(context.Background(), *ndr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		_, err := kubeClient.CoreV1().ConfigMaps(ndr.Namespace).Get(pkgcommon.NsqLookupdConfigMapName(ndr.Name), metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			klog.Warningf("Get nsqlookupd configmap %s/%s error: %v", ndr.Namespace, pkgcommon.NsqLookupdConfigMapName(ndr.Name), err)
			return
		}

		cancel()
		klog.Infof("Delete nsqlookupd configmap %s/%s success", ndr.Namespace, pkgcommon.NsqLookupdConfigMapName(ndr.Name))
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Delete nsqlookupd configmap %s/%s error: %v", ndr.Namespace, pkgcommon.NsqLookupdConfigMapName(ndr.Name), ctx.Err())
		return ctx.Err()
	}

	err = kubeClient.CoreV1().ConfigMaps(ndr.Namespace).Delete(pkgcommon.NsqAdminConfigMapName(ndr.Name), &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("Delete nsqadmin configmap %s/%s error: %v", ndr.Namespace, pkgcommon.NsqAdminConfigMapName(ndr.Name), err)
		return err
	}

	klog.Infof("Waiting for nsqadmin configmap %s/%s to be deleted", ndr.Namespace, pkgcommon.NsqAdminConfigMapName(ndr.Name))
	ctx, cancel = context.WithTimeout(context.Background(), *ndr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		_, err := kubeClient.CoreV1().ConfigMaps(ndr.Namespace).Get(pkgcommon.NsqAdminConfigMapName(ndr.Name), metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			klog.Warningf("Get nsqadmin configmap %s/%s error: %v", ndr.Namespace, pkgcommon.NsqAdminConfigMapName(ndr.Name), err)
			return
		}

		cancel()
		klog.Infof("Delete nsqadmin configmap %s/%s success", ndr.Namespace, pkgcommon.NsqAdminConfigMapName(ndr.Name))
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Delete nsqadmin configmap %s/%s error: %v", ndr.Namespace, pkgcommon.NsqAdminConfigMapName(ndr.Name), ctx.Err())
		return ctx.Err()
	}

	err = kubeClient.CoreV1().ConfigMaps(ndr.Namespace).Delete(pkgcommon.NsqdConfigMapName(ndr.Name), &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("Delete nsqd configmap %s/%s error: %v", ndr.Namespace, pkgcommon.NsqdConfigMapName(ndr.Name), err)
		return err
	}

	klog.Infof("Waiting for nsqd configmap %s/%s to be deleted", ndr.Namespace, pkgcommon.NsqdConfigMapName(ndr.Name))
	ctx, cancel = context.WithTimeout(context.Background(), *ndr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		_, err := kubeClient.CoreV1().ConfigMaps(ndr.Namespace).Get(pkgcommon.NsqdConfigMapName(ndr.Name), metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			klog.Warningf("Get nsqd configmap %s/%s error: %v", ndr.Namespace, pkgcommon.NsqdConfigMapName(ndr.Name), err)
			return
		}

		cancel()
		klog.Infof("Delete nsqd configmap %s/%s success", ndr.Namespace, pkgcommon.NsqdConfigMapName(ndr.Name))
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Delete nsqd configmap %s/%s error: %v", ndr.Namespace, pkgcommon.NsqdConfigMapName(ndr.Name), ctx.Err())
		return ctx.Err()
	}

	err = nsqClient.NsqV1alpha1().NsqdScales(ndr.Namespace).Delete(ndr.Name, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("Delete nsqdscale %s/%s error: %v", ndr.Namespace, ndr.Name, err)
		return err
	}

	klog.Infof("Waiting for nsqdscale %s/%s to be deleted", ndr.Namespace, ndr.Name)
	ctx, cancel = context.WithTimeout(context.Background(), *ndr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		_, err := nsqClient.NsqV1alpha1().NsqdScales(ndr.Namespace).Get(ndr.Name, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			klog.Warningf("Get nsqdscale %s/%s error: %v", ndr.Namespace, ndr.Name, err)
			return
		}

		cancel()
		klog.Infof("Delete nsqdscale %s/%s success", ndr.Namespace, ndr.Name)
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Delete nsqdscale %s/%s error: %v", ndr.Namespace, ndr.Name, ctx.Err())
		return ctx.Err()
	}

	return nil
}

func ScaleNsqAdmin(nsqClient *versioned.Clientset, nasr *types.NsqAdminReplicaUpdateRequest) error {
	klog.Infof("Set nsqadmin %s/%s replicas to %d", nasr.Namespace, nasr.Name, nasr.Replicas)
	ctx, cancel := context.WithTimeout(context.Background(), *nasr.WaitTimeout)

	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqAdmin, err := nsqClient.NsqV1alpha1().NsqAdmins(nasr.Namespace).Get(nasr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqadmin %s/%s error: %v", nasr.Namespace, nasr.Name, err)
			return
		}

		nsqAdminCopy := nsqAdmin.DeepCopy()
		nsqAdminCopy.Spec.Replicas = nasr.Replicas
		_, err = nsqClient.NsqV1alpha1().NsqAdmins(nasr.Namespace).Update(nsqAdminCopy)
		if err != nil {
			klog.Errorf("Update nsqadmin %s/%s error: %v", nasr.Namespace, nasr.Name, err)
			return
		}

		cancel()
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for setting nsqadmin %s/%s replicas", nasr.Namespace, nasr.Name)
		return ctx.Err()
	}

	klog.Infof("Waiting for nsqadmin %s/%s replicas to %d", nasr.Namespace, nasr.Name, nasr.Replicas)
	ctx, cancel = context.WithTimeout(context.Background(), *nasr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqAdmin, err := nsqClient.NsqV1alpha1().NsqAdmins(nasr.Namespace).Get(nasr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqadmin %s/%s error: %v", nasr.Namespace, nasr.Name, err)
			return
		}

		if nsqAdmin.Status.AvailableReplicas != nasr.Replicas {
			klog.V(2).Infof("Spec and status replicas does not match for nsqadmin %s/%s. Old replicas: %v, New replicas: %v",
				nasr.Namespace, nasr.Name, nsqAdmin.Status.AvailableReplicas, nasr.Replicas)
			return
		}

		cancel()
	}, constant.NsqAdminStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqadmin %s/%s reach its replicas %d", nasr.Namespace, nasr.Name, nasr.Replicas)
		return ctx.Err()
	}

	return nil
}

func ScaleNsqLookupd(nsqClient *versioned.Clientset, nlsr *types.NsqLookupdReplicaUpdateRequest) error {
	klog.Infof("Set nsqlookupd %s/%s replicas to %d", nlsr.Namespace, nlsr.Name, nlsr.Replicas)
	ctx, cancel := context.WithTimeout(context.Background(), *nlsr.WaitTimeout)

	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqLookupd, err := nsqClient.NsqV1alpha1().NsqLookupds(nlsr.Namespace).Get(nlsr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqLookupd %s/%s error: %v", nlsr.Namespace, nlsr.Name, err)
			return
		}

		nsqLookupdCopy := nsqLookupd.DeepCopy()
		nsqLookupdCopy.Spec.Replicas = nlsr.Replicas
		_, err = nsqClient.NsqV1alpha1().NsqLookupds(nlsr.Namespace).Update(nsqLookupdCopy)
		if err != nil {
			klog.Errorf("Update nsqlookupd %s/%s error: %v", nlsr.Namespace, nlsr.Name, err)
			return
		}

		cancel()
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for setting nsqlookupd %s/%s replicas", nlsr.Namespace, nlsr.Name)
		return ctx.Err()
	}

	klog.Infof("Waiting for nsqlookupd %s/%s replicas to %d", nlsr.Namespace, nlsr.Name, nlsr.Replicas)
	ctx, cancel = context.WithTimeout(context.Background(), *nlsr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqLookupd, err := nsqClient.NsqV1alpha1().NsqLookupds(nlsr.Namespace).Get(nlsr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqlookupd %s/%s error: %v", nlsr.Namespace, nlsr.Name, err)
			return
		}

		if nsqLookupd.Status.AvailableReplicas != nlsr.Replicas {
			klog.V(2).Infof("Spec and status replicas does not match for nsqlookupd %s/%s. Old replicas: %v, New replicas: %v",
				nlsr.Namespace, nlsr.Name, nsqLookupd.Status.AvailableReplicas, nlsr.Replicas)
			return
		}

		cancel()
	}, constant.NsqLookupdStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqlookupd %s/%s reach its replicas %d", nlsr.Namespace, nlsr.Name, nlsr.Replicas)
		return ctx.Err()
	}

	return nil
}

func ScaleNsqd(nsqClient *versioned.Clientset, ndsr *types.NsqdReplicaUpdateRequest) error {
	klog.Infof("Set nsqd %s/%s replicas to %d", ndsr.Namespace, ndsr.Name, ndsr.Replicas)
	ctx, cancel := context.WithTimeout(context.Background(), *ndsr.WaitTimeout)

	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqd, err := nsqClient.NsqV1alpha1().Nsqds(ndsr.Namespace).Get(ndsr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqd %s/%s error: %v", ndsr.Namespace, ndsr.Name, err)
			return
		}

		nsqdCopy := nsqd.DeepCopy()
		nsqdCopy.Spec.Replicas = ndsr.Replicas
		_, err = nsqClient.NsqV1alpha1().Nsqds(ndsr.Namespace).Update(nsqdCopy)
		if err != nil {
			klog.Errorf("Update nsqd %s/%s error: %v", ndsr.Namespace, ndsr.Name, err)
			return
		}

		cancel()
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for setting nsqd %s/%s replicas", ndsr.Namespace, ndsr.Name)
		return ctx.Err()
	}

	klog.Infof("Waiting for nsqd %s/%s replicas to %d", ndsr.Namespace, ndsr.Name, ndsr.Replicas)
	ctx, cancel = context.WithTimeout(context.Background(), *ndsr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqd, err := nsqClient.NsqV1alpha1().Nsqds(ndsr.Namespace).Get(ndsr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqd %s/%s error: %v", ndsr.Namespace, ndsr.Name, err)
			return
		}

		if nsqd.Status.AvailableReplicas != ndsr.Replicas {
			klog.V(2).Infof("Spec and status replicas does not match for nsqd %s/%s. Old replicas: %v, New replicas: %v",
				ndsr.Namespace, ndsr.Name, nsqd.Status.AvailableReplicas, ndsr.Replicas)
			return
		}

		cancel()
	}, constant.NsqdStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqd %s/%s reach its replicas %d", ndsr.Namespace, ndsr.Name, ndsr.Replicas)
		return ctx.Err()
	}

	return nil
}

func UpdateNsqAdminImage(kubeClient *kubernetes.Clientset, nsqClient *versioned.Clientset, nauir *types.NsqAdminUpdateImageRequest) error {
	klog.Infof("Set nsqadmin %s/%s image to %s", nauir.Namespace, nauir.Name, nauir.Image)
	ctx, cancel := context.WithTimeout(context.Background(), *nauir.WaitTimeout)

	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqAdmin, err := nsqClient.NsqV1alpha1().NsqAdmins(nauir.Namespace).Get(nauir.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqadmin %s/%s error: %v", nauir.Namespace, nauir.Name, err)
			return
		}

		nsqAdminCopy := nsqAdmin.DeepCopy()
		nsqAdminCopy.Spec.Image = nauir.Image
		_, err = nsqClient.NsqV1alpha1().NsqAdmins(nauir.Namespace).Update(nsqAdminCopy)
		if err != nil {
			klog.Errorf("Update nsqadmin %s/%s error: %v", nauir.Namespace, nauir.Name, err)
			return
		}

		cancel()
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for setting nsqadmin %s/%s image", nauir.Namespace, nauir.Name)
		return ctx.Err()
	}

	klog.Infof("Waiting for nsqadmin %s/%s image to %s", nauir.Namespace, nauir.Name, nauir.Image)
	ctx, cancel = context.WithTimeout(context.Background(), *nauir.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": pkgcommon.NsqAdminDeploymentName(nauir.Name)}}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			klog.Errorf("Failed to generate label selector for nsqadmin %s/%s: %v", nauir.Namespace, nauir.Name, err)
			return
		}

		podList, err := kubeClient.CoreV1().Pods(nauir.Namespace).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})

		if err != nil {
			klog.Errorf("Failed to list nsqadmin %s/%s pods: %v", nauir.Namespace, nauir.Name, err)
			return
		}

		for _, pod := range podList.Items {
			if pod.Spec.Containers[0].Image != nauir.Image {
				klog.Infof("Spec and status image does not match for nsqadmin %s/%s. Pod: %v. Spec image: %v, new image: %v",
					nauir.Namespace, nauir.Name, pod.Name, pod.Spec.Containers[0].Image, nauir.Image)
				return
			}
		}

		nsqAdminDep, err := kubeClient.AppsV1().Deployments(nauir.Namespace).Get(pkgcommon.NsqAdminDeploymentName(nauir.Name), metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get nsqadmin %s/%s deployment", nauir.Namespace, nauir.Name)
			return
		}

		if !(nsqAdminDep.Status.ReadyReplicas == *nsqAdminDep.Spec.Replicas && nsqAdminDep.Status.Replicas == nsqAdminDep.Status.ReadyReplicas) {
			klog.Errorf("Waiting for nsqadmin %s/%s pods ready", nauir.Namespace, nauir.Name)
			return
		}

		cancel()
	}, constant.NsqAdminStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqadmin %s/%s reach image %s", nauir.Namespace, nauir.Name, nauir.Image)
		return ctx.Err()
	}

	return nil
}

func UpdateNsqLookupdImage(kubeClient *kubernetes.Clientset, nsqClient *versioned.Clientset, nluir *types.NsqLookupdUpdateImageRequest) error {
	klog.Infof("Set nsqlookupd %s/%s image to %s", nluir.Namespace, nluir.Name, nluir.Image)
	ctx, cancel := context.WithTimeout(context.Background(), *nluir.WaitTimeout)

	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqLookupd, err := nsqClient.NsqV1alpha1().NsqLookupds(nluir.Namespace).Get(nluir.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqlookupd %s/%s error: %v", nluir.Namespace, nluir.Name, err)
			return
		}

		nsqLookupdCopy := nsqLookupd.DeepCopy()
		nsqLookupdCopy.Spec.Image = nluir.Image
		_, err = nsqClient.NsqV1alpha1().NsqLookupds(nluir.Namespace).Update(nsqLookupdCopy)
		if err != nil {
			klog.Errorf("Update nsqlookupd %s/%s error: %v", nluir.Namespace, nluir.Name, err)
			return
		}

		cancel()
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for setting nsqlookupd %s/%s image", nluir.Namespace, nluir.Name)
		return ctx.Err()
	}

	klog.Infof("Waiting for nsqlookupd %s/%s image to %s", nluir.Namespace, nluir.Name, nluir.Image)
	ctx, cancel = context.WithTimeout(context.Background(), *nluir.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": pkgcommon.NsqLookupdDeploymentName(nluir.Name)}}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			klog.Errorf("Failed to generate label selector for nsqlookupd %s/%s: %v", nluir.Namespace, nluir.Name, err)
			return
		}

		podList, err := kubeClient.CoreV1().Pods(nluir.Namespace).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})

		if err != nil {
			klog.Errorf("Failed to list nsqlookupd %s/%s pods: %v", nluir.Namespace, nluir.Name, err)
			return
		}

		for _, pod := range podList.Items {
			if pod.Spec.Containers[0].Image != nluir.Image {
				klog.Infof("Spec and status image does not match for nsqlookupd %s/%s. Pod: %v. Spec image: %v, new image: %v",
					nluir.Namespace, nluir.Name, pod.Name, pod.Spec.Containers[0].Image, nluir.Image)
				return
			}
		}

		nsqLookupdDep, err := kubeClient.AppsV1().Deployments(nluir.Namespace).Get(pkgcommon.NsqLookupdDeploymentName(nluir.Name), metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get nsqlookupd %s/%s deployment", nluir.Namespace, nluir.Name)
			return
		}

		if !(nsqLookupdDep.Status.ReadyReplicas == *nsqLookupdDep.Spec.Replicas && nsqLookupdDep.Status.Replicas == nsqLookupdDep.Status.ReadyReplicas) {
			klog.Errorf("Waiting for nsqlookupd %s/%s pods ready", nluir.Namespace, nluir.Name)
			return
		}

		cancel()
	}, constant.NsqLookupdStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqlookupd %s/%s reach image %s", nluir.Namespace, nluir.Name, nluir.Image)
		return ctx.Err()
	}

	return nil
}

func UpdateNsqdImage(kubeClient *kubernetes.Clientset, nsqClient *versioned.Clientset, nduir *types.NsqdUpdateImageRequest) error {
	klog.Infof("Set nsqd %s/%s image to %s", nduir.Namespace, nduir.Name, nduir.Image)
	ctx, cancel := context.WithTimeout(context.Background(), *nduir.WaitTimeout)

	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqd, err := nsqClient.NsqV1alpha1().Nsqds(nduir.Namespace).Get(nduir.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqd %s/%s error: %v", nduir.Namespace, nduir.Name, err)
			return
		}

		nsqdCopy := nsqd.DeepCopy()
		nsqdCopy.Spec.Image = nduir.Image
		_, err = nsqClient.NsqV1alpha1().Nsqds(nduir.Namespace).Update(nsqdCopy)
		if err != nil {
			klog.Errorf("Update nsqd %s/%s error: %v", nduir.Namespace, nduir.Name, err)
			return
		}

		cancel()
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for setting nsqd %s/%s image", nduir.Namespace, nduir.Name)
		return ctx.Err()
	}

	klog.Infof("Waiting for nsqd %s/%s image to %s", nduir.Namespace, nduir.Name, nduir.Image)
	ctx, cancel = context.WithTimeout(context.Background(), *nduir.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": pkgcommon.NsqdStatefulSetName(nduir.Name)}}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			klog.Errorf("Failed to generate label selector for nsqd %s/%s: %v", nduir.Namespace, nduir.Name, err)
			return
		}

		podList, err := kubeClient.CoreV1().Pods(nduir.Namespace).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})

		if err != nil {
			klog.Errorf("Failed to list nsqd %s/%s pods: %v", nduir.Namespace, nduir.Name, err)
			return
		}

		for _, pod := range podList.Items {
			if pod.Spec.Containers[0].Image != nduir.Image {
				klog.Infof("Spec and status image does not match for nsqd %s/%s. Pod: %v. Spec image: %v, new image: %v",
					nduir.Namespace, nduir.Name, pod.Name, pod.Spec.Containers[0].Image, nduir.Image)
				return
			}
		}

		nsqdSs, err := kubeClient.AppsV1().StatefulSets(nduir.Namespace).Get(pkgcommon.NsqdStatefulSetName(nduir.Name), metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get nsqd %s/%s statefulset", nduir.Namespace, nduir.Name)
			return
		}

		if nsqdSs.Status.ReadyReplicas != *nsqdSs.Spec.Replicas {
			klog.Errorf("Waiting for nsqd %s/%s pods ready", nduir.Namespace, nduir.Name)
			return
		}

		cancel()
	}, constant.NsqdStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqd %s/%s reach image %s", nduir.Namespace, nduir.Name, nduir.Image)
		return ctx.Err()
	}

	return nil
}

func AdjustNsqdConfig(kubeClient *kubernetes.Clientset, ndcr *types.NsqdConfigRequest) error {
	klog.Infof("Update nsqd %s/%s config to %#v", ndcr.Namespace, ndcr.Name, ndcr.AssembleNsqdConfigMap([]string{}).Data)
	ctx, cancel := context.WithTimeout(context.Background(), *ndcr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		// Create nsqd configmap
		labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": pkgcommon.NsqLookupdDeploymentName(ndcr.Name)}}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			klog.Errorf("Gen label selector for nsqlookupd %s/%s pods error: %v", ndcr.Namespace, ndcr.Name, err)
			return
		}

		podList, err := kubeClient.CoreV1().Pods(ndcr.Namespace).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})

		if err != nil {
			klog.Errorf("List nsqlookupd %s/%s pods error: %v", ndcr.Namespace, ndcr.Name, err)
			return
		}

		var addresses []string
		for _, pod := range podList.Items {
			addresses = append(addresses, fmt.Sprintf("%s:%v", pod.Status.PodIP, constant.NsqLookupdTcpPort))
		}

		nsqdCMWanted := ndcr.AssembleNsqdConfigMap(addresses)

		nsqdCM, err := kubeClient.CoreV1().ConfigMaps(ndcr.Namespace).Get(pkgcommon.NsqdConfigMapName(ndcr.Name), metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqd %s/%s configmap error: %v", ndcr.Namespace, ndcr.Name, err)
			return
		}

		nsqdCMNew := nsqdCM.DeepCopy()
		nsqdCMNew.Data = nsqdCMWanted.Data

		_, err = kubeClient.CoreV1().ConfigMaps(ndcr.Namespace).Update(nsqdCMNew)
		if err != nil {
			klog.Errorf("Update nsqd %s/%s configmap error: %v", ndcr.Namespace, ndcr.Name, err)
		}

		klog.Infof("Update nsqd %s/%s configmap success", ndcr.Namespace, ndcr.Name)
		cancel()
	}, constant.ConfigMapStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqd %s/%s configmap to be ready", ndcr.Namespace, ndcr.Name)
		return ctx.Err()
	}

	configmap, err := kubeClient.CoreV1().ConfigMaps(ndcr.Namespace).Get(pkgcommon.NsqdConfigMapName(ndcr.Name), metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Get configmap for nsqd %s/%s error: %v", ndcr.Namespace, ndcr.Name, err)
		return err
	}
	configmapHash, err := pkgcommon.Hash(configmap.Data)
	if err != nil {
		klog.Errorf("Hash configmap data for nsqd %s/%s error: %v", ndcr.Namespace, ndcr.Name, err)
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), *ndcr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": pkgcommon.NsqdStatefulSetName(ndcr.Name)}}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			klog.Errorf("Failed to generate label selector for nsqd %s/%s: %v", ndcr.Namespace, ndcr.Name, err)
			return
		}

		podList, err := kubeClient.CoreV1().Pods(ndcr.Namespace).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})

		if err != nil {
			klog.Errorf("Failed to list nsqd %s/%s pods: %v", ndcr.Namespace, ndcr.Name, err)
			return
		}

		for _, pod := range podList.Items {
			if val, exists := pod.GetAnnotations()[constant.NsqConfigMapAnnotationKey]; exists && val != configmapHash {
				klog.Infof("Spec and status signature annotation do not match for nsqd %s/%s. Pod: %v. "+
					"Spec signature annotation: %v, new signature annotation: %v",
					ndcr.Namespace, ndcr.Name, pod.Name, pod.GetAnnotations()[constant.NsqConfigMapAnnotationKey], configmapHash)
				return
			}
		}

		nsqdSs, err := kubeClient.AppsV1().StatefulSets(ndcr.Namespace).Get(pkgcommon.NsqdStatefulSetName(ndcr.Name), metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get nsqd %s/%s statefulset", ndcr.Namespace, ndcr.Name)
			return
		}

		if nsqdSs.Status.ReadyReplicas != *nsqdSs.Spec.Replicas {
			klog.Errorf("Waiting for nsqd %s/%s pods ready", ndcr.Namespace, ndcr.Name)
			return
		}

		klog.Infof("Nsqd %s/%s configmap change rolling update success", ndcr.Namespace, ndcr.Name)
		cancel()
	}, constant.NsqdStatusCheckPeriod)

	return nil
}

func AdjustNsqdMemoryResources(nsqClient *versioned.Clientset, ndcr *types.NsqdConfigRequest) error {
	klog.Infof("Update nsqd %s/%s config to %s", ndcr.Namespace, ndcr.Name, ndcr.String())
	ctx, cancel := context.WithTimeout(context.Background(), *ndcr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqd, err := nsqClient.NsqV1alpha1().Nsqds(ndcr.Namespace).Get(ndcr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get nsqd %s/%s: %v", ndcr.Namespace, ndcr.Name, err)
			return
		}

		nsqdCopy := nsqd.DeepCopy()
		nsqdCopy.Spec.MessageAvgSize = ndcr.GetMessageAvgSize()
		nsqdCopy.Spec.MemoryQueueSize = ndcr.GetMemoryQueueSize()
		nsqdCopy.Spec.MemoryOverBookingPercent = ndcr.GetMemoryOverBookingPercent()
		nsqdCopy.Spec.ChannelCount = ndcr.GetChannelCount()

		_, err = nsqClient.NsqV1alpha1().Nsqds(ndcr.Namespace).Update(nsqdCopy)
		if err != nil {
			klog.Errorf("Update nsqd %s/%s error: %v", ndcr.Namespace, ndcr.Name, err)
			return
		}

		klog.Infof("Update nsqd %s/%s success", ndcr.Namespace, ndcr.Name)
		cancel()
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqd %s/%s spec change", ndcr.Namespace, ndcr.Name)
		return ctx.Err()
	}

	ctx, cancel = context.WithTimeout(context.Background(), *ndcr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqd, err := nsqClient.NsqV1alpha1().Nsqds(ndcr.Namespace).Get(ndcr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get nsqd %s/%s: %v", ndcr.Namespace, ndcr.Name, err)
			return
		}

		if !(nsqd.Status.MessageAvgSize == ndcr.MessageAvgSize &&
			nsqd.Status.MemoryQueueSize == ndcr.MemoryQueueSize &&
			nsqd.Status.MemoryOverBookingPercent == ndcr.MemoryOverBookingPercent &&
			nsqd.Status.ChannelCount == ndcr.ChannelCount) {
			klog.Errorf("Waiting for nsqd %s/%s status reaches its spec", ndcr.Namespace, ndcr.Name)
			return
		}

		klog.Infof("Update nsqd %s/%s success", ndcr.Namespace, ndcr.Name)
		cancel()
	}, constant.NsqdStatusCheckPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqd %s/%s reaces its spec", ndcr.Namespace, ndcr.Name)
		return ctx.Err()
	}

	return nil
}

func AdjustNsqdScale(nsqClient *versioned.Clientset, ndsur *types.NsqdScaleUpdateRequest) error {
	klog.Infof("Update nsqdscale %s/%s to %+v", ndsur.Namespace, ndsur.Name, ndsur)
	ctx, cancel := context.WithTimeout(context.Background(), *ndsur.WaitTimeout)

	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqdScale, err := nsqClient.NsqV1alpha1().NsqdScales(ndsur.Namespace).Get(ndsur.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqdscale %s/%s error: %v", ndsur.Namespace, ndsur.Name, err)
			return
		}

		nsqdScaleCopy := nsqdScale.DeepCopy()
		nsqdScaleCopy.Spec.QpsThreshold = ndsur.QpsThreshold
		nsqdScaleCopy.Spec.Minimum = ndsur.Minimum
		nsqdScaleCopy.Spec.Maximum = ndsur.Maximum
		nsqdScaleCopy.Spec.Enabled = ndsur.Enabled
		_, err = nsqClient.NsqV1alpha1().NsqdScales(ndsur.Namespace).Update(nsqdScaleCopy)
		if err != nil {
			klog.Errorf("Update nsqdscale %s/%s error: %v", ndsur.Namespace, ndsur.Name, err)
			return
		}

		klog.Infof("Update nsqdscale %s/%s success", ndsur.Namespace, ndsur.Name)
		cancel()
	}, constant.ResourceUpdateRetryPeriod)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for updating nsqdscale %s/%s", ndsur.Namespace, ndsur.Name)
		return ctx.Err()
	}

	return nil
}
