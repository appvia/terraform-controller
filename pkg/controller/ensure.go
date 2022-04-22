/*
 * Copyright (C) 2022 Rohith Jayawardene <gambol99@gmail.com>
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either version 2
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package controller

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1alphav1 "github.com/appvia/terraform-controller/pkg/apis/core/v1alpha1"
)

var (
	// ErrIgnore is used to used to stop the ensure chain
	ErrIgnore = errors.New("resource is being ignored")
	// DefaultEnsureHandler is the default sequential runner for a list of ensure functions
	DefaultEnsureHandler = EnsureRunner{}
)

// EnsureFunc defines a method to ensure a state
type EnsureFunc func(ctx context.Context) (reconcile.Result, error)

// EnsureRunner provides a wrapper for running ensure funcs
type EnsureRunner struct{}

// RequeueImmediate should be used anywhere we wish an immediate / ASAP requeue to be performed
var RequeueImmediate = reconcile.Result{RequeueAfter: 5 * time.Millisecond}

// Run is a generic handler for running the ensure methods
func (e *EnsureRunner) Run(ctx context.Context, cc client.Client, resource Object, ensures []EnsureFunc) (result reconcile.Result, rerr error) {
	original := resource.DeepCopyObject()
	status := resource.GetCommonStatus()

	if status.LastReconcile == nil {
		status.LastReconcile = &corev1alphav1.LastReconcileStatus{}
	}
	status.LastReconcile.Time = metav1.NewTime(time.Now())
	status.LastReconcile.Generation = resource.GetGeneration()

	// @here we are responsible for updating the transition times of the conditions where we
	// see a drift. And updating the status of the resource overall
	defer func() {
		// @step: we need to update the status of the resource
		if err := cc.Status().Patch(ctx, resource, client.MergeFrom(original.(client.Object))); err != nil {
			if err := client.IgnoreNotFound(err); err != nil {
				log.WithError(err).Error("failed to update the status of resource")

				rerr = err
				result = reconcile.Result{}
			}
		}
	}()

	for _, x := range ensures {
		result, rerr = x(ctx)
		if rerr != nil {
			switch {
			case kerrors.IsConflict(rerr):
				rerr = nil
				result = RequeueImmediate
				return

			case rerr == ErrIgnore:
				rerr = nil
				result = reconcile.Result{}
				return
			}

			return reconcile.Result{}, rerr
		}
		if result.Requeue || result.RequeueAfter > 0 {
			return result, nil
		}
	}

	cond := ConditionMgr(resource, corev1alphav1.ConditionReady)
	cond.Success("Resource ready")

	if status.LastSuccess == nil {
		status.LastSuccess = &corev1alphav1.LastReconcileStatus{}
	}
	status.LastSuccess.Time = metav1.NewTime(time.Now())
	status.LastSuccess.Generation = resource.GetGeneration()

	return reconcile.Result{}, nil
}
