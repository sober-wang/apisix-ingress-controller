// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package translation

import (
	"context"
	"testing"

	fakeapisix "github.com/apache/apisix-ingress-controller/pkg/kube/apisix/client/clientset/versioned/fake"
	apisixinformers "github.com/apache/apisix-ingress-controller/pkg/kube/apisix/client/informers/externalversions"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"k8s.io/apimachinery/pkg/util/intstr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"

	configv2alpha1 "github.com/apache/apisix-ingress-controller/pkg/kube/apisix/apis/config/v2alpha1"
)

func TestRouteMatchExpr(t *testing.T) {
	tr := &translator{}
	value1 := "text/plain"
	value2 := "gzip"
	value3 := "13"
	value4 := ".*\\.php"
	exprs := []configv2alpha1.ApisixRouteHTTPMatchExpr{
		{
			Subject: configv2alpha1.ApisixRouteHTTPMatchExprSubject{
				Scope: configv2alpha1.ScopeHeader,
				Name:  "Content-Type",
			},
			Op:    configv2alpha1.OpEqual,
			Value: &value1,
		},
		{
			Subject: configv2alpha1.ApisixRouteHTTPMatchExprSubject{
				Scope: configv2alpha1.ScopeHeader,
				Name:  "Content-Encoding",
			},
			Op:    configv2alpha1.OpNotEqual,
			Value: &value2,
		},
		{
			Subject: configv2alpha1.ApisixRouteHTTPMatchExprSubject{
				Scope: configv2alpha1.ScopeQuery,
				Name:  "ID",
			},
			Op:    configv2alpha1.OpGreaterThan,
			Value: &value3,
		},
		{
			Subject: configv2alpha1.ApisixRouteHTTPMatchExprSubject{
				Scope: configv2alpha1.ScopeQuery,
				Name:  "ID",
			},
			Op:    configv2alpha1.OpLessThan,
			Value: &value3,
		},
		{
			Subject: configv2alpha1.ApisixRouteHTTPMatchExprSubject{
				Scope: configv2alpha1.ScopeQuery,
				Name:  "ID",
			},
			Op:    configv2alpha1.OpRegexMatch,
			Value: &value4,
		},
		{
			Subject: configv2alpha1.ApisixRouteHTTPMatchExprSubject{
				Scope: configv2alpha1.ScopeQuery,
				Name:  "ID",
			},
			Op:    configv2alpha1.OpRegexMatchCaseInsensitive,
			Value: &value4,
		},
		{
			Subject: configv2alpha1.ApisixRouteHTTPMatchExprSubject{
				Scope: configv2alpha1.ScopeQuery,
				Name:  "ID",
			},
			Op:    configv2alpha1.OpRegexNotMatch,
			Value: &value4,
		},
		{
			Subject: configv2alpha1.ApisixRouteHTTPMatchExprSubject{
				Scope: configv2alpha1.ScopeQuery,
				Name:  "ID",
			},
			Op:    configv2alpha1.OpRegexNotMatchCaseInsensitive,
			Value: &value4,
		},
		{
			Subject: configv2alpha1.ApisixRouteHTTPMatchExprSubject{
				Scope: configv2alpha1.ScopeCookie,
				Name:  "domain",
			},
			Op: configv2alpha1.OpIn,
			Set: []string{
				"a.com",
				"b.com",
			},
		},
	}
	results, err := tr.translateRouteMatchExprs(exprs)
	assert.Nil(t, err)
	assert.Len(t, results, 9)

	assert.Len(t, results[0], 3)
	assert.Equal(t, results[0][0].StrVal, "http_content_type")
	assert.Equal(t, results[0][1].StrVal, "==")
	assert.Equal(t, results[0][2].StrVal, "text/plain")

	assert.Len(t, results[1], 3)
	assert.Equal(t, results[1][0].StrVal, "http_content_encoding")
	assert.Equal(t, results[1][1].StrVal, "~=")
	assert.Equal(t, results[1][2].StrVal, "gzip")

	assert.Len(t, results[2], 3)
	assert.Equal(t, results[2][0].StrVal, "arg_id")
	assert.Equal(t, results[2][1].StrVal, ">")
	assert.Equal(t, results[2][2].StrVal, "13")

	assert.Len(t, results[3], 3)
	assert.Equal(t, results[3][0].StrVal, "arg_id")
	assert.Equal(t, results[3][1].StrVal, "<")
	assert.Equal(t, results[3][2].StrVal, "13")

	assert.Len(t, results[4], 3)
	assert.Equal(t, results[4][0].StrVal, "arg_id")
	assert.Equal(t, results[4][1].StrVal, "~~")
	assert.Equal(t, results[4][2].StrVal, ".*\\.php")

	assert.Len(t, results[5], 3)
	assert.Equal(t, results[5][0].StrVal, "arg_id")
	assert.Equal(t, results[5][1].StrVal, "~*")
	assert.Equal(t, results[5][2].StrVal, ".*\\.php")

	assert.Len(t, results[6], 4)
	assert.Equal(t, results[6][0].StrVal, "arg_id")
	assert.Equal(t, results[6][1].StrVal, "!")
	assert.Equal(t, results[6][2].StrVal, "~~")
	assert.Equal(t, results[6][3].StrVal, ".*\\.php")

	assert.Len(t, results[7], 4)
	assert.Equal(t, results[7][0].StrVal, "arg_id")
	assert.Equal(t, results[7][1].StrVal, "!")
	assert.Equal(t, results[7][2].StrVal, "~*")
	assert.Equal(t, results[7][3].StrVal, ".*\\.php")

	assert.Len(t, results[8], 3)
	assert.Equal(t, results[8][0].StrVal, "cookie_domain")
	assert.Equal(t, results[8][1].StrVal, "in")
	assert.Equal(t, results[8][2].SliceVal, []string{"a.com", "b.com"})
}

func TestTranslateApisixRouteV2alpha1WithDuplicatedName(t *testing.T) {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc",
			Namespace: "test",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "port1",
					Port: 80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 9080,
					},
				},
				{
					Name: "port2",
					Port: 443,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 9443,
					},
				},
			},
		},
	}
	endpoints := &corev1.Endpoints{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc",
			Namespace: "test",
		},
		Subsets: []corev1.EndpointSubset{
			{
				Ports: []corev1.EndpointPort{
					{
						Name: "port1",
						Port: 9080,
					},
					{
						Name: "port2",
						Port: 9443,
					},
				},
				Addresses: []corev1.EndpointAddress{
					{IP: "192.168.1.1"},
					{IP: "192.168.1.2"},
				},
			},
		},
	}

	client := fake.NewSimpleClientset()
	informersFactory := informers.NewSharedInformerFactory(client, 0)
	svcInformer := informersFactory.Core().V1().Services().Informer()
	svcLister := informersFactory.Core().V1().Services().Lister()
	epInformer := informersFactory.Core().V1().Endpoints().Informer()
	epLister := informersFactory.Core().V1().Endpoints().Lister()
	apisixClient := fakeapisix.NewSimpleClientset()
	apisixInformersFactory := apisixinformers.NewSharedInformerFactory(apisixClient, 0)

	_, err := client.CoreV1().Endpoints("test").Create(context.Background(), endpoints, metav1.CreateOptions{})
	assert.Nil(t, err)
	_, err = client.CoreV1().Services("test").Create(context.Background(), svc, metav1.CreateOptions{})
	assert.Nil(t, err)

	tr := &translator{
		&TranslatorOptions{
			EndpointsLister:      epLister,
			ServiceLister:        svcLister,
			ApisixUpstreamLister: apisixInformersFactory.Apisix().V1().ApisixUpstreams().Lister(),
		},
	}

	processCh := make(chan struct{})
	svcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			processCh <- struct{}{}
		},
	})
	epInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			processCh <- struct{}{}
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)
	go svcInformer.Run(stopCh)
	go epInformer.Run(stopCh)
	cache.WaitForCacheSync(stopCh, svcInformer.HasSynced)

	<-processCh
	<-processCh

	ar := &configv2alpha1.ApisixRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ar",
			Namespace: "test",
		},
		Spec: &configv2alpha1.ApisixRouteSpec{
			HTTP: []*configv2alpha1.ApisixRouteHTTP{
				{
					Name: "rule1",
					Match: &configv2alpha1.ApisixRouteHTTPMatch{
						Paths: []string{
							"/*",
						},
					},
					Backend: &configv2alpha1.ApisixRouteHTTPBackend{
						ServiceName: "svc",
						ServicePort: intstr.IntOrString{
							IntVal: 80,
						},
					},
				},
				{
					Name: "rule1",
					Match: &configv2alpha1.ApisixRouteHTTPMatch{
						Paths: []string{
							"/*",
						},
					},
					Backend: &configv2alpha1.ApisixRouteHTTPBackend{
						ServiceName: "svc",
						ServicePort: intstr.IntOrString{
							IntVal: 80,
						},
					},
				},
			},
		},
	}

	_, err = tr.TranslateRouteV2alpha1(ar)
	assert.Equal(t, err.Error(), "duplicated route rule name")
}
