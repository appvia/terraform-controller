#!/bin/bash
#
# Copyright (C) 2022  Appvia Ltd <info@appvia.io>
#
# This program is free software; you can redistribute it and/or
# modify it under the terms of the GNU General Public License
# as published by the Free Software Foundation; either version 2
# of the License, or (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#

UNITS="test/e2e/integration"
BATS_OPTIONS=${BATS_OPTIONS:-""}
BUCKET=${BUCKET:-"terraform-controller-ci-bucket"}

run_bats() {
  echo "Running unit: ${@}"

  APP_NAMESPACE="apps" \
  BUCKET=${BUCKET} \
  NAMESPACE="terraform-system" \
  bats ${BATS_OPTIONS} ${@} || exit 1
}

# run-checks runs a collection checks
run_checks() {
  units=(
    "setup"
    "provider"
    "plan"
    "apply"
    "confirm"
    "destroy"
  )

  for x in "${units[@]}"; do
    run_bats ${UNITS}/${x}.bats
  done
}

run_checks
