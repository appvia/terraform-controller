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

package assets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsset(t *testing.T) {
	a, err := Asset("github/main.yaml.tpl")
	assert.NoError(t, err)
	assert.NotEmpty(t, a)
}

func TestMustAsset(t *testing.T) {
	a := MustAsset("github/main.yaml.tpl")
	assert.NotEmpty(t, a)
}

func TestAssetNames(t *testing.T) {
	assert.NotEmpty(t, AssetNames())
}

func TestRecursiveAssetNames(t *testing.T) {
	assert.NotEmpty(t, RecursiveAssetNames("github"))
}
