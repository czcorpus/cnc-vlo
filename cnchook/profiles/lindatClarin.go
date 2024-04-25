// Copyright 2024 Martin Zimandl <martin.zimandl@gmail.com>
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

package profiles

import (
	"github.com/czcorpus/cnc-vlo/cnchook/profiles/components"
	"github.com/czcorpus/cnc-vlo/oaipmh/formats"
)

// note - omitempties are optional
// profile is derived from LINDAT_CLARIN profile

type CNCResourceProfile struct {
	BibliographicInfo components.BibliographicInfoComponent `xml:"cmdp:CNC_Resource>cmdp:bibliographicInfo"`
	DataInfoInfo      components.DataInfoComponent          `xml:"cmdp:CNC_Resource>cmdp:dataInfo"`
	LicenseInfo       []LicenseElement                      `xml:"cmdp:CNC_Resource>cmdp:licenseInfo>cmdp:license"`
	RelationsInfo     *[]formats.TypedElement               `xml:"cmdp:CNC_Resource>cmdp:relationsInfo>cmdp:relation,omitempty"`
}

func (c *CNCResourceProfile) GetSchemaURL() string {
	return "https://catalog.clarin.eu/ds/ComponentRegistry/rest/registry/1.x/profiles/clarin.eu:cr1:p_1712653174418/xsd"
}

type LicenseElement struct {
	Name string `xml:"cmdp:name,omitempty"`
	URI  string `xml:"cmdp:uri"`
}
