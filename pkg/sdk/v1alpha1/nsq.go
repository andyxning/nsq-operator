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
	"time"

	"github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
	pkgcommon "github.com/andyxning/nsq-operator/pkg/common"
	"github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned"
	"github.com/andyxning/nsq-operator/pkg/sdk/v1alpha1/common"
	"github.com/andyxning/nsq-operator/pkg/sdk/v1alpha1/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

func CreateCluster(kubeClient *kubernetes.Clientset, nsqClient *versioned.Clientset, nr *types.NsqCreateRequest) error {
	// Create nsqlookupd configmap
	nsqLookupdCM := common.AssembleNsqLookupdConfigMap(nr)
	klog.Infof("Create nsqlookupd configmap %s/%s", nsqLookupdCM.Namespace, nsqLookupdCM.Name)
	_, err := kubeClient.CoreV1().ConfigMaps(nr.Namespace).Create(nsqLookupdCM)
	if errors.IsAlreadyExists(err) {
		klog.Infof("Nsqlookupd configmap %s/%s exists. Update it", nsqLookupdCM.Namespace, nsqLookupdCM.Name)
		var old *corev1.ConfigMap
		if klog.V(4) {
			old, err = kubeClient.CoreV1().ConfigMaps(nr.Namespace).Get(nsqLookupdCM.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get old nsqlookupd configmap %s/%s error: %v", nsqLookupdCM.Namespace, nsqLookupdCM.Name, err)
				return err
			}

			klog.V(4).Infof("Old nsqlookupd configmap %s/%s: %#v", nsqLookupdCM.Namespace, nsqLookupdCM.Name, *old)
		}

		oldCopy := old.DeepCopy()
		oldCopy.Data = nsqLookupdCM.Data
		klog.V(4).Infof("New nsqlookupd configmap %s/%s: %#v", nsqLookupdCM.Namespace, nsqLookupdCM.Name, *oldCopy)
		_, err = kubeClient.CoreV1().ConfigMaps(nr.Namespace).Update(oldCopy)
	}

	if err != nil {
		klog.Errorf("Create nsqlookupd configmap error: %v", err)
		return err
	}

	// Create nsqlookupd
	nsqLookupd := common.AssembleNsqLookupd(nr)
	klog.Infof("Create nsqlookupd %s/%s", nsqLookupd.Namespace, nsqLookupd.Name)
	_, err = nsqClient.NsqV1alpha1().NsqLookupds(nr.Namespace).Create(nsqLookupd)
	if errors.IsAlreadyExists(err) {
		klog.Infof("Nsqlookupd %s/%s exists. Update it", nsqLookupd.Namespace, nsqLookupd.Name)
		var old *v1alpha1.NsqLookupd
		if klog.V(4) {
			old, err = nsqClient.NsqV1alpha1().NsqLookupds(nr.Namespace).Get(nsqLookupd.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get old nsqlookupd %s/%s error: %v", nsqLookupd.Namespace, nsqLookupd.Name, err)
				return err
			}

			klog.V(4).Infof("Old nsqlookupd %s/%s: %#v", nsqLookupd.Namespace, nsqLookupd.Name, *old)
		}

		oldCopy := old.DeepCopy()
		oldCopy.Spec = nsqLookupd.Spec
		klog.V(4).Infof("New nsqlookupd %s/%s: %#v", nsqLookupd.Namespace, nsqLookupd.Name, *oldCopy)
		_, err = nsqClient.NsqV1alpha1().NsqLookupds(nr.Namespace).Update(oldCopy)
	}

	if err != nil {
		klog.Errorf("Create nsqlookupd error: %v", err)
		return err
	}

	// Wait for nsqlookupd to reach its spec
	klog.V(2).Infof("Waiting for nsqlookupd %s/%s reach its spec", nsqLookupd.Namespace, nsqLookupd.Name)
	ctx, cancel := context.WithTimeout(context.Background(), *nr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqLookupd, err := nsqClient.NsqV1alpha1().NsqLookupds(nr.Namespace).Get(nr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqlookupd error: %v", err)
			return
		}

		if *nsqLookupd.Spec.Replicas != nsqLookupd.Status.AvailableReplicas {
			klog.V(2).Infof("Spec and status does not match for nsqlookupd %s/%s. Spec: %v, Status: %v",
				nr.Namespace, nr.Name, nsqLookupd.Spec, nsqLookupd.Status)
			return
		}

		// Spec and status matches
		cancel()
	}, 8*time.Second)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqlookupd resource to reach its spec")
		return ctx.Err()
	}

	// Check for existence of nsqadmin configmap
	klog.V(2).Infof("Waiting for nsqadmin configmap %s/%s reach its spec", nr.Namespace, pkgcommon.NsqAdminConfigMapName(nr.Name))
	ctx, cancel = context.WithTimeout(context.Background(), *nr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqadminCM, err := kubeClient.CoreV1().ConfigMaps(nr.Namespace).Get(pkgcommon.NsqAdminConfigMapName(nr.Name), metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqadmin configmap error: %v", err)
			return
		}

		if nsqadminCM.Data != nil && len(nsqadminCM.Data) != 0 {
			// nsqadmin configmap exists
			cancel()
		}
	}, time.Second)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqadmin configmap to be ready")
		return ctx.Err()
	}

	// Create nsqadmin
	nsqadmin := common.AssembleNsqAdmin(nr)
	klog.Infof("Create nsqadmin %s/%s", nsqadmin.Namespace, nsqadmin.Name)
	_, err = nsqClient.NsqV1alpha1().NsqAdmins(nr.Namespace).Create(nsqadmin)
	if errors.IsAlreadyExists(err) {
		klog.Infof("Nsqadmin %s/%s exists. Update it", nsqadmin.Namespace, nsqadmin.Name)
		var old *v1alpha1.NsqAdmin
		if klog.V(4) {
			old, err = nsqClient.NsqV1alpha1().NsqAdmins(nr.Namespace).Get(nsqadmin.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get old nsqadmin %s/%s error: %v", nsqadmin.Namespace, nsqadmin.Name, err)
				return err
			}

			klog.V(4).Infof("Old nsqadmin %s/%s: %#v", nsqadmin.Namespace, nsqadmin.Name, *old)
		}

		oldCopy := old.DeepCopy()
		oldCopy.Spec = nsqadmin.Spec
		klog.V(4).Infof("New nsqadmin %s/%s: %#v", nsqadmin.Namespace, nsqadmin.Name, *oldCopy)
		_, err = nsqClient.NsqV1alpha1().NsqAdmins(nr.Namespace).Update(oldCopy)
	}

	if err != nil {
		klog.Errorf("Create nsqadmin error: %v", err)
		return err
	}

	// Wait for nsqadmin to reach its spec
	klog.V(2).Infof("Waiting for nsqadmin %s/%s reach its spec", nsqadmin.Namespace, nsqadmin.Name)
	ctx, cancel = context.WithTimeout(context.Background(), *nr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqAdmin, err := nsqClient.NsqV1alpha1().NsqAdmins(nr.Namespace).Get(nr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqadmin error: %v", err)
			return
		}

		if *nsqAdmin.Spec.Replicas != nsqAdmin.Status.AvailableReplicas {
			klog.V(2).Infof("Spec and status does not match for nsqadmin %s/%s. Spec: %v, Status: %v",
				nr.Namespace, nr.Name, nsqAdmin.Spec, nsqAdmin.Status)
			return
		}

		// Spec and status matches
		cancel()
	}, 8*time.Second)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqadmin resource to reach its spec")
		return ctx.Err()
	}

	// Create nsqd configmap
	labelSelector := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster": pkgcommon.NsqLookupdDeploymentName(nr.Name)}}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		klog.Errorf("Gen label selector for nsqlookupd %s/%s pods error: %v", nr.Namespace, nr.Name, err)
		return err
	}

	podList, err := kubeClient.CoreV1().Pods(nr.Namespace).List(metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		klog.Errorf("List nsqlookupd %s/%s pods error: %v", nr.Namespace, nr.Name, err)
		return err
	}

	var addresses []string
	for _, pod := range podList.Items {
		addresses = append(addresses, fmt.Sprintf("%s:%v", pod.Status.PodIP, 4161))
	}

	nsqdCM := common.AssembleNsqdConfigMap(nr, addresses)
	klog.Infof("Create nsqd configmap %s/%s", nsqdCM.Namespace, nsqdCM.Name)
	_, err = kubeClient.CoreV1().ConfigMaps(nr.Namespace).Create(nsqdCM)
	if errors.IsAlreadyExists(err) {
		klog.Infof("Nsqd configmap %s/%s exists. Update it", nsqdCM.Namespace, nsqdCM.Name)
		var old *corev1.ConfigMap
		if klog.V(4) {
			old, err = kubeClient.CoreV1().ConfigMaps(nr.Namespace).Get(nsqdCM.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get old nsqd configmap %s/%s error: %v", nsqdCM.Namespace, nsqdCM.Name, err)
				return err
			}

			klog.V(4).Infof("Old nsqd configmap %s/%s: %#v", nsqdCM.Namespace, nsqdCM.Name, *old)
		}

		oldCopy := old.DeepCopy()
		oldCopy.Data = nsqdCM.Data
		klog.V(4).Infof("New nsqd configmap %s/%s: %#v", nsqdCM.Namespace, nsqdCM.Name, *oldCopy)
		_, err = kubeClient.CoreV1().ConfigMaps(nr.Namespace).Update(oldCopy)
	}

	if err != nil {
		klog.Errorf("Create nsqd configmap error: %v", err)
		return err
	}

	// Create nsqd
	nsqd := common.AssembleNsqd(nr)
	klog.Infof("Create nsqd %s/%s", nsqd.Namespace, nsqd.Name)
	_, err = nsqClient.NsqV1alpha1().Nsqds(nr.Namespace).Create(nsqd)
	if errors.IsAlreadyExists(err) {
		klog.Infof("Nsqd %s/%s exists. Update it", nsqd.Namespace, nsqd.Name)
		var old *v1alpha1.Nsqd
		if klog.V(4) {
			old, err = nsqClient.NsqV1alpha1().Nsqds(nr.Namespace).Get(nsqd.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Get old nsqd %s/%s error: %v", nsqd.Namespace, nsqd.Name, err)
				return err
			}

			klog.V(4).Infof("Old nsqd %s/%s: %#v", nsqd.Namespace, nsqd.Name, *old)
		}

		oldCopy := old.DeepCopy()
		oldCopy.Spec = nsqd.Spec
		klog.V(4).Infof("New nsqd %s/%s: %#v", nsqd.Namespace, nsqd.Name, *oldCopy)
		_, err = nsqClient.NsqV1alpha1().Nsqds(nr.Namespace).Update(oldCopy)
	}

	if err != nil {
		klog.Errorf("Create nsqd error: %v", err)
		return err
	}

	// Wait for nsqd to reach its spec
	klog.V(2).Infof("Waiting for nsqd %s/%s reach its spec", nsqd.Namespace, nsqd.Name)
	ctx, cancel = context.WithTimeout(context.Background(), *nr.WaitTimeout)
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqd, err := nsqClient.NsqV1alpha1().Nsqds(nr.Namespace).Get(nr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqd error: %v", err)
			return
		}

		if *nsqd.Spec.Replicas != nsqd.Status.AvailableReplicas {
			klog.V(2).Infof("Spec and status does not match for nsqd %s/%s. Spec: %v, Status: %v",
				nr.Namespace, nr.Name, nsqd.Spec, nsqd.Status)
			return
		}

		// Spec and status matches
		cancel()
	}, 8*time.Second)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqd resource to reach its spec")
		return ctx.Err()
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
	}, 8*time.Second)

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
	}, 8*time.Second)

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
	}, 8*time.Second)

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
	}, time.Second)

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
	}, time.Second)

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
	}, time.Second)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Delete nsqd configmap %s/%s error: %v", ndr.Namespace, pkgcommon.NsqdConfigMapName(ndr.Name), ctx.Err())
		return ctx.Err()
	}

	return nil
}

func ScaleNsqAdmin(nsqClient *versioned.Clientset, nasr *types.NsqAdminScaleRequest) error {
	klog.Infof("Set nsqadmin %s/%s replicas to %d", nasr.Namespace, nasr.Name, nasr.Replicas)
	ctx, cancel := context.WithTimeout(context.Background(), *nasr.WaitTimeout)

	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqAdmin, err := nsqClient.NsqV1alpha1().NsqAdmins(nasr.Namespace).Get(nasr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqadmin %s/%s error: %v", nasr.Namespace, nasr.Name, err)
			return
		}

		nsqAdminCopy := nsqAdmin.DeepCopy()
		nsqAdminCopy.Spec.Replicas = &nasr.Replicas
		_, err = nsqClient.NsqV1alpha1().NsqAdmins(nasr.Namespace).Update(nsqAdminCopy)
		if err != nil {
			klog.Errorf("Update nsqadmin %s/%s error: %v", nasr.Namespace, nasr.Name, err)
			return
		}

		cancel()
	}, 1*time.Second)

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
			klog.V(2).Infof("Spec and status replicas do not match for nsqadmin %s/%s. Old replicas: %v, New replicas: %v",
				nasr.Namespace, nasr.Name, nsqAdmin.Status.AvailableReplicas, nasr.Replicas)
			return
		}

		cancel()
	}, 8*time.Second)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqadmin %s/%s reach its replicas %d", nasr.Namespace, nasr.Name, nasr.Replicas)
		return ctx.Err()
	}

	return nil
}

func ScaleNsqLookupd(nsqClient *versioned.Clientset, nlsr *types.NsqLookupdScaleRequest) error {
	klog.Infof("Set nsqlookupd %s/%s replicas to %d", nlsr.Namespace, nlsr.Name, nlsr.Replicas)
	ctx, cancel := context.WithTimeout(context.Background(), *nlsr.WaitTimeout)

	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqLookupd, err := nsqClient.NsqV1alpha1().NsqLookupds(nlsr.Namespace).Get(nlsr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqLookupd %s/%s error: %v", nlsr.Namespace, nlsr.Name, err)
			return
		}

		nsqLookupdCopy := nsqLookupd.DeepCopy()
		nsqLookupdCopy.Spec.Replicas = &nlsr.Replicas
		_, err = nsqClient.NsqV1alpha1().NsqLookupds(nlsr.Namespace).Update(nsqLookupdCopy)
		if err != nil {
			klog.Errorf("Update nsqlookupd %s/%s error: %v", nlsr.Namespace, nlsr.Name, err)
			return
		}

		cancel()
	}, 1*time.Second)

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
			klog.V(2).Infof("Spec and status replicas do not match for nsqlookupd %s/%s. Old replicas: %v, New replicas: %v",
				nlsr.Namespace, nlsr.Name, nsqLookupd.Status.AvailableReplicas, nlsr.Replicas)
			return
		}

		cancel()
	}, 8*time.Second)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqlookupd %s/%s reach its replicas %d", nlsr.Namespace, nlsr.Name, nlsr.Replicas)
		return ctx.Err()
	}

	return nil
}

func ScaleNsqd(nsqClient *versioned.Clientset, ndsr *types.NsqdScaleRequest) error {
	klog.Infof("Set nsqd %s/%s replicas to %d", ndsr.Namespace, ndsr.Name, ndsr.Replicas)
	ctx, cancel := context.WithTimeout(context.Background(), *ndsr.WaitTimeout)

	wait.UntilWithContext(ctx, func(ctx context.Context) {
		nsqd, err := nsqClient.NsqV1alpha1().Nsqds(ndsr.Namespace).Get(ndsr.Name, metav1.GetOptions{})
		if err != nil {
			klog.Warningf("Get nsqd %s/%s error: %v", ndsr.Namespace, ndsr.Name, err)
			return
		}

		nsqdCopy := nsqd.DeepCopy()
		nsqdCopy.Spec.Replicas = &ndsr.Replicas
		_, err = nsqClient.NsqV1alpha1().Nsqds(ndsr.Namespace).Update(nsqdCopy)
		if err != nil {
			klog.Errorf("Update nsqd %s/%s error: %v", ndsr.Namespace, ndsr.Name, err)
			return
		}

		cancel()
	}, 1*time.Second)

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
			klog.V(2).Infof("Spec and status replicas do not match for nsqd %s/%s. Old replicas: %v, New replicas: %v",
				ndsr.Namespace, ndsr.Name, nsqd.Status.AvailableReplicas, ndsr.Replicas)
			return
		}

		cancel()
	}, 8*time.Second)

	if ctx.Err() != context.Canceled {
		klog.Errorf("Timeout for waiting nsqd %s/%s reach its replicas %d", ndsr.Namespace, ndsr.Name, ndsr.Replicas)
		return ctx.Err()
	}

	return nil
}
