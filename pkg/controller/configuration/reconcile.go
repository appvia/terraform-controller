/*
 * Copyright (C) 2022 Appvia Ltd <info@appvia.io>
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

package configuration

import (
	"context"

	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	terraformv1alphav1 "github.com/appvia/terraform-controller/pkg/apis/terraform/v1alpha1"
	"github.com/appvia/terraform-controller/pkg/controller"
)

type state struct {
	// auth is an optional secret which is used for authentication
	auth *v1.Secret
	// policies is a list of policies in the cluster
	policies *terraformv1alphav1.PolicyList
	// provider is the credentials provider to use
	provider *terraformv1alphav1.Provider
	// checkovConstraint is the polict constraint for this configuration
	checkovConstraint *terraformv1alphav1.PolicyConstraint
	// jobs is list of all jobs for this configuration and generation
	jobs *batchv1.JobList
	// jobTemplate is the template to use when rendering the job
	jobTemplate []byte
}

// Reconcile is called to handle the reconciliation of the provider resource
func (c *Controller) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	configuration := &terraformv1alphav1.Configuration{}

	if err := c.cc.Get(ctx, request.NamespacedName, configuration); err != nil {
		if kerrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		log.WithError(err).Error("failed to retrieve the configuration resource")

		return reconcile.Result{}, err
	}

	state := &state{}

	finalizer := controller.NewFinalizer(c.cc, controllerName)
	if finalizer.IsDeletionCandidate(configuration) {
		result, err := controller.DefaultEnsureHandler.Run(ctx, c.cc, configuration,
			[]controller.EnsureFunc{
				c.ensurePoliciesList(configuration, state),
				c.ensureJobsList(configuration, state),
				c.ensureProviderIsReady(configuration, state),
				c.ensureAuthenticationSecret(configuration),
				c.ensureJobTemplate(configuration, state),
				c.ensureTerraformDestroy(configuration, state),
				c.ensureTerraformConfigDeleted(configuration),
				c.ensureConfigurationSecretsDeleted(configuration),
				c.ensureConfigurationJobsDeleted(configuration),
				finalizer.EnsureRemoved(configuration),
			})
		if err != nil {
			log.WithError(err).Error("failed to delete the configuration resource")

			return reconcile.Result{}, err
		}

		return result, err
	}

	// @step: ensure the conditions are registered
	controller.EnsureConditionsRegistered(terraformv1alphav1.DefaultConfigurationConditions, configuration)

	result, err := controller.DefaultEnsureHandler.Run(ctx, c.cc, configuration,
		[]controller.EnsureFunc{
			finalizer.EnsurePresent(configuration),
			c.ensurePoliciesList(configuration, state),
			c.ensureJobsList(configuration, state),
			c.ensureNoPreviousGeneration(configuration, state),
			c.ensureInfracostsSecret(configuration),
			c.ensureJobTemplate(configuration, state),
			c.ensureAuthenticationSecret(configuration),
			c.ensureProviderIsReady(configuration, state),
			c.ensureGeneratedConfig(configuration, state),
			c.ensureTerraformPlan(configuration, state),
			c.ensureCostStatus(configuration),
			c.ensureCheckovPolicy(configuration, state),
			c.ensureTerraformApply(configuration, state),
			c.ensureTerraformStatus(configuration),
			c.ensureTerraformSecret(configuration),
		})
	if err != nil {
		log.WithError(err).Error("failed to reconcile the configuration resource")

		return reconcile.Result{}, err
	}

	return result, err
}
