/*
 * Copyright (C) 2022  Appvia Ltd <info@appvia.io>
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

package providers

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/appvia/terraform-controller/pkg/schema"
	"github.com/appvia/terraform-controller/test/fixtures"
)

func TestReconcile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Running Test Suite")
}

var _ = Describe("Provider Validation", func() {
	ctx := context.Background()
	var cc client.Client
	var v *validator

	JustAfterEach(func() {
		cc = fake.NewClientBuilder().WithScheme(schema.GetScheme()).WithRuntimeObjects(fixtures.NewNamespace("default")).Build()
		v = &validator{cc: cc}
	})

	When("creating a provider", func() {
		It("should not error when creating a valid provider", func() {
			err := v.ValidateCreate(ctx, fixtures.NewValidAWSProvider("default", "test"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not error when updating a valid provider", func() {
			err := v.ValidateUpdate(ctx, nil, fixtures.NewValidAWSProvider("default", "test"))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("creating a provider with an incorrect provider", func() {
		It("should throw error", func() {
			policy := fixtures.NewValidAWSProvider("default", "test")
			policy.Spec.Provider = "invalid"
			msg := "spec.provider: invalid is not supported (must be aws,google,azurerm)"

			err := v.ValidateCreate(ctx, policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(msg))

			err = v.ValidateUpdate(ctx, nil, policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(msg))
		})
	})

	When("creating a provider with a secret", func() {
		It("should throw error when no secret reference", func() {
			policy := fixtures.NewValidAWSProvider("default", "test")
			policy.Spec.SecretRef = nil
			msg := "spec.secretRef: secret is required when source is secret"

			err := v.ValidateCreate(ctx, policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(msg))

			err = v.ValidateUpdate(ctx, nil, policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(msg))
		})

		It("should throw error when no secret name in reference", func() {
			policy := fixtures.NewValidAWSProvider("default", "test")
			policy.Spec.SecretRef.Name = ""
			msg := "spec.secretRef.name: name is required when source is secret"

			err := v.ValidateCreate(ctx, policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(msg))

			err = v.ValidateUpdate(ctx, nil, policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(msg))
		})

		It("should throw error when no secret namespace in reference", func() {
			policy := fixtures.NewValidAWSProvider("default", "test")
			policy.Spec.SecretRef.Namespace = ""
			msg := "spec.secretRef.namespace: namespace is required when source is secret"

			err := v.ValidateCreate(ctx, policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(msg))

			err = v.ValidateUpdate(ctx, nil, policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(msg))
		})
	})

	When("creating a provider with a injected identity", func() {
		It("should throw error when no service account", func() {
			policy := fixtures.NewValidAWSProvider("default", "test")
			policy.Spec.Source = "injected"
			policy.Spec.ServiceAccount = nil
			msg := "spec.serviceAccount: serviceAccount is required when source is injected"

			err := v.ValidateCreate(ctx, policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(msg))

			err = v.ValidateUpdate(ctx, nil, policy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(msg))
		})

		It("should not throw an error when service account defined", func() {
			policy := fixtures.NewValidAWSProvider("default", "test")
			policy.Spec.SecretRef.Name = ""
			policy.Spec.Source = "injected"
			policy.Spec.ServiceAccount = pointer.String("test")

			err := v.ValidateCreate(ctx, policy)
			Expect(err).ToNot(HaveOccurred())
			err = v.ValidateUpdate(ctx, nil, policy)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
