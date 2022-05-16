#!/usr/bin/env bats
#
# Copyright 2021 Appvia Ltd <info@appvia.io>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

load test/e2e/lib/helper.bash

setup() {
  [[ ! -f ${BATS_PARENT_TMPNAME}.skip ]] || skip "skip remaining tests"
}

teardown() {
  [[ -n "$BATS_TEST_COMPLETED" ]] || touch ${BATS_PARENT_TMPNAME}.skip
}

@test "We should be able to create a namespace for testing" {
  cat <<EOF > ${BATS_TMPDIR}/resource.yaml
apiVersion: v1
kind: Namespace
metadata:
  labels:
    kubernetes.io/metadata.name: apps
  name: apps
EOF
  runit "kubectl apply -f ${BATS_TMPDIR}/resource.yaml"
  [[ "$status" -eq 0 ]]
  runit "kubectl -n ${APP_NAMESPACE} delete job --all"
  [[ "$status" -eq 0 ]]
  runit "kubectl -n ${APP_NAMESPACE} delete po --all"
  [[ "$status" -eq 0 ]]
}

@test "We should have a clean terraform namespace for testing" {
  labels="terraform.appvia.io/configuration=bucket,terraform.appvia.io/stage=plan"

  runit "kubectl -n ${NAMESPACE} delete job -l ${labels}"
  [[ "$status" -eq 0 ]]
}

@test "We should be able to create a configuration" {
cat <<EOF > ${BATS_TMPDIR}/resource.yaml
---
apiVersion: terraform.appvia.io/v1alpha1
kind: Configuration
metadata:
  name: bucket
spec:
  module: https://github.com/terraform-aws-modules/terraform-aws-s3-bucket.git?ref=v3.1.0
  providerRef:
    namespace: terraform-system
    name: default
  writeConnectionSecretToRef:
    name: test
  variables:
    unused: $(date +"%s")
    bucket: terraform-controller-ci-bucket
    acl: private
    versioning:
      enabled: true
    block_public_acls: true
    block_public_policy: true
    ignore_public_acls: true
    restrict_public_buckets: true
    server_side_encryption_configuration:
      rule:
        apply_server_side_encryption_by_default:
          sse_algorithm: "aws:kms"
        bucket_key_enabled: true
EOF
  runit "kubectl -n ${APP_NAMESPACE} apply -f ${BATS_TMPDIR}/resource.yaml"
  [[ "$status" -eq 0 ]]
  runit "kubectl -n ${APP_NAMESPACE} get configuration bucket"
  [[ "$status" -eq 0 ]]
}

@test "We should have a condition indicating the provider is ready" {
  runit "kubectl -n ${APP_NAMESPACE} get configuration bucket -o json" "jq -r '.status.conditions[0].name' | grep -q 'Provider ready'"
  [[ "$status" -eq 0 ]]
  runit "kubectl -n ${APP_NAMESPACE} get configuration bucket -o json" "jq -r '.status.conditions[0].status' | grep -q 'True'"
  [[ "$status" -eq 0 ]]
}

@test "We should have a job created in the terraform-system running the plan" {
  labels="terraform.appvia.io/configuration=bucket,terraform.appvia.io/stage=plan"

  runit "kubectl -n ${NAMESPACE} get job -l ${labels}"
  [[ "$status" -eq 0 ]]
}

@test "We should have a watcher job created in the configuration namespace" {
  labels="terraform.appvia.io/configuration=bucket,terraform.appvia.io/stage=plan"

  runit "kubectl -n ${APP_NAMESPACE} get job -l ${labels}"
  [[ "$status" -eq 0 ]]
}

@test "We should see the terraform plan complete sucessfully" {
  labels="terraform.appvia.io/configuration=bucket,terraform.appvia.io/stage=plan"

  retry 10 "kubectl -n ${NAMESPACE} get job -l ${labels} -o json" "jq -r '.items[0].status.conditions[0].type' | grep -q Complete"
  [[ "$status" -eq 0 ]]
  runit "kubectl -n ${NAMESPACE} get job -l ${labels} -o json" "jq -r '.items[0].status.conditions[0].status' | grep -q True"
  [[ "$status" -eq 0 ]]
}

@test "We should have a completed watcher job in the application namespace" {
  labels="terraform.appvia.io/configuration=bucket,terraform.appvia.io/stage=plan"

  runit "kubectl -n ${APP_NAMESPACE} get job -l ${labels} -o json" "jq -r '.items[0].status.conditions[0].type' | grep -q Complete"
  [[ "$status" -eq 0 ]]
  runit "kubectl -n ${APP_NAMESPACE} get job -l ${labels} -o json" "jq -r '.items[0].status.conditions[0].status' | grep -q True"
  [[ "$status" -eq 0 ]]
}

@test "We should have a configuration in pending approval" {
  expected="Waiting for terraform apply annotation to be set to true"

  runit "kubectl -n ${APP_NAMESPACE} get configuration bucket -o json" "jq -r '.status.conditions[2].name' | grep -q 'Terraform Apply'"
  [[ "$status" -eq 0 ]]
  runit "kubectl -n ${APP_NAMESPACE} get configuration bucket -o json" "jq -r '.status.conditions[2].reason' | grep -q 'ActionRequired'"
  [[ "$status" -eq 0 ]]
  runit "kubectl -n ${APP_NAMESPACE} get configuration bucket -o json" "jq -r '.status.conditions[2].status' | grep -q 'False'"
  [[ "$status" -eq 0 ]]
  runit "kubectl -n ${APP_NAMESPACE} get configuration bucket -o json" "jq -r '.status.conditions[2].type' | grep -q 'TerraformApply'"
  [[ "$status" -eq 0 ]]
  runit "kubectl -n ${APP_NAMESPACE} get configuration bucket -o json" "jq -r '.status.conditions[2].message' | grep -q '${expected}'"
  [[ "$status" -eq 0 ]]
}

