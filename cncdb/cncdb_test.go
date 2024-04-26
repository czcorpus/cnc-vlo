// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cncdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestParseLocaleOK(t *testing.T) {
	var h CNCMySQLHandler
	tag, err := h.parseLocale("en_US")
	assert.NoError(t, err)
	b, conf := tag.Base()
	assert.Equal(t, language.Exact, conf)
	assert.Equal(t, "en", b.String())
	reg, conf := tag.Region()
	assert.Equal(t, language.Exact, conf)
	assert.Equal(t, "US", reg.String())
}

func TestParseLocaleOKWithEncoding(t *testing.T) {
	var h CNCMySQLHandler
	tag, err := h.parseLocale("en_US.UTF-8")
	assert.NoError(t, err)
	b, conf := tag.Base()
	assert.Equal(t, language.Exact, conf)
	assert.Equal(t, "en", b.String())
	reg, conf := tag.Region()
	assert.Equal(t, language.Exact, conf)
	assert.Equal(t, "US", reg.String())
}

func TestParseLocaleOKBase(t *testing.T) {
	var h CNCMySQLHandler
	tag, err := h.parseLocale("cs")
	assert.NoError(t, err)
	b, conf := tag.Base()
	assert.Equal(t, language.Exact, conf)
	assert.Equal(t, "cs", b.String())
	reg, conf := tag.Region()
	assert.Equal(t, language.Low, conf)
	assert.Equal(t, "CZ", reg.String())
}

func TestParseLocaleBroken(t *testing.T) {
	var h CNCMySQLHandler
	tag, err := h.parseLocale("en_EN")
	assert.NoError(t, err)
	b, conf := tag.Base()
	assert.Equal(t, language.Exact, conf)
	assert.Equal(t, "en", b.String())
	reg, conf := tag.Region()
	assert.Equal(t, language.Low, conf)
	assert.Equal(t, "US", reg.String())
}
