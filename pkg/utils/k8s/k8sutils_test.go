// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package k8s

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestToObjectMeta(t *testing.T) {
	assert.Equal(
		t,
		metav1.ObjectMeta{Namespace: "namespace", Name: "name"},
		ToObjectMeta(types.NamespacedName{Namespace: "namespace", Name: "name"}),
	)
}

func TestExtractNamespacedName(t *testing.T) {
	assert.Equal(
		t,
		types.NamespacedName{Namespace: "namespace", Name: "name"},
		ExtractNamespacedName(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "namespace", Name: "name"}}),
	)
}

func TestGetServiceDNSName(t *testing.T) {
	type args struct {
		svc corev1.Service
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "sample service",
			args: args{
				svc: corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns", Name: "test-name"}},
			},
			want: []string{"test-name.test-ns.svc", "test-name.test-ns"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := deep.Equal(GetServiceDNSName(tt.args.svc), tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func TestOverrideControllerReference(t *testing.T) {

	ownerRefFixture := func(name string, controller bool) metav1.OwnerReference {
		return metav1.OwnerReference{
			APIVersion: "v1",
			Kind:       "some",
			Name:       name,
			UID:        "uid",
			Controller: &controller,
		}
	}
	type args struct {
		obj      metav1.Object
		newOwner metav1.OwnerReference
	}
	tests := []struct {
		name      string
		args      args
		assertion func(object metav1.Object)
	}{
		{
			name: "no existing controller",
			args: args{
				obj:      &corev1.Secret{},
				newOwner: ownerRefFixture("obj1", true),
			},
			assertion: func(object metav1.Object) {
				require.Equal(t, object.GetOwnerReferences(), []metav1.OwnerReference{ownerRefFixture("obj1", true)})
			},
		},
		{
			name: "replace existing controller",
			args: args{
				obj: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						OwnerReferences: []metav1.OwnerReference{
							ownerRefFixture("obj1", true),
						},
					},
				},
				newOwner: ownerRefFixture("obj2", true),
			},
			assertion: func(object metav1.Object) {
				require.Equal(t, object.GetOwnerReferences(), []metav1.OwnerReference{
					ownerRefFixture("obj2", true)})
			},
		},
		{
			name: "replace existing controller preserving existing references",
			args: args{
				obj: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						OwnerReferences: []metav1.OwnerReference{
							ownerRefFixture("other", false),
							ownerRefFixture("obj1", true),
						},
					},
				},
				newOwner: ownerRefFixture("obj2", true),
			},
			assertion: func(object metav1.Object) {
				require.Equal(t, object.GetOwnerReferences(), []metav1.OwnerReference{
					ownerRefFixture("other", false),
					ownerRefFixture("obj2", true)})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			OverrideControllerReference(tt.args.obj, tt.args.newOwner)
			tt.assertion(tt.args.obj)
		})
	}
}
