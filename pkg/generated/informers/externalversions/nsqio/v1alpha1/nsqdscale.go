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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	nsqiov1alpha1 "github.com/andyxning/nsq-operator/pkg/apis/nsqio/v1alpha1"
	versioned "github.com/andyxning/nsq-operator/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/andyxning/nsq-operator/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/andyxning/nsq-operator/pkg/generated/listers/nsqio/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// NsqdScaleInformer provides access to a shared informer and lister for
// NsqdScales.
type NsqdScaleInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.NsqdScaleLister
}

type nsqdScaleInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewNsqdScaleInformer constructs a new informer for NsqdScale type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNsqdScaleInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNsqdScaleInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredNsqdScaleInformer constructs a new informer for NsqdScale type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNsqdScaleInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NsqV1alpha1().NsqdScales(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NsqV1alpha1().NsqdScales(namespace).Watch(options)
			},
		},
		&nsqiov1alpha1.NsqdScale{},
		resyncPeriod,
		indexers,
	)
}

func (f *nsqdScaleInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNsqdScaleInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *nsqdScaleInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&nsqiov1alpha1.NsqdScale{}, f.defaultInformer)
}

func (f *nsqdScaleInformer) Lister() v1alpha1.NsqdScaleLister {
	return v1alpha1.NewNsqdScaleLister(f.Informer().GetIndexer())
}
