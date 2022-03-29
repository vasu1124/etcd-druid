// Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lease

import (
	"context"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"

	gardenercomponent "github.com/gardener/gardener/pkg/operation/botanist/component"
	coordinationv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type component struct {
	client    client.Client
	namespace string

	values Values
}

func (c *component) Deploy(ctx context.Context) error {
	var (
		deltaSnapshotLease = c.emptyLease(c.values.DeltaSnapshotLeaseName)
		fullSnapshotLease  = c.emptyLease(c.values.FullSnapshotLeaseName)
	)

	if err := c.syncSnapshotLease(ctx, deltaSnapshotLease); err != nil {
		return err
	}

	if err := c.syncSnapshotLease(ctx, fullSnapshotLease); err != nil {
		return err
	}

	if err := c.syncMemberLeases(ctx); err != nil {
		return err
	}

	return nil
}

func (c *component) Destroy(ctx context.Context) error {
	var (
		deltaSnapshotLease = c.emptyLease(c.values.DeltaSnapshotLeaseName)
		fullSnapshotLease  = c.emptyLease(c.values.FullSnapshotLeaseName)
	)

	if err := c.deleteSnapshotLease(ctx, deltaSnapshotLease); err != nil {
		return err
	}

	if err := c.deleteSnapshotLease(ctx, fullSnapshotLease); err != nil {
		return err
	}

	if err := c.deleteAllMemberLeases(ctx); err != nil {
		return err
	}

	return nil
}

// New creates a new lease deployer instance.
func New(c client.Client, namespace string, values Values) gardenercomponent.Deployer {
	return &component{
		client:    c,
		namespace: namespace,
		values:    values,
	}
}

func (c *component) emptyLease(name string) *coordinationv1.Lease {
	return &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.namespace,
		},
	}
}

func getOwnerReferences(val Values) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         druidv1alpha1.GroupVersion.String(),
			Kind:               "Etcd",
			Name:               val.EtcdName,
			UID:                val.EtcdUID,
			Controller:         pointer.BoolPtr(true),
			BlockOwnerDeletion: pointer.BoolPtr(true),
		},
	}
}