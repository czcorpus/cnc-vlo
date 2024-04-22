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

package components

import "github.com/czcorpus/cnc-vlo/oaipmh/formats"

// note - omitempties are optional

type BibliographicInfoComponent struct {
	ProjectUrl    string                     `xml:"cmdp:projectUrl,omitempty"`
	Version       string                     `xml:"cmdp:version,omitempty"`
	Titles        []formats.MultilangElement `xml:"cmdp:titles>cmdp:title"`
	Authors       []AuthorComponent          `xml:"cmdp:authors>cmdp:author"`
	Dates         *DatesComponent            `xml:"cmdp:dates,omitempty"`
	Identifiers   []formats.TypedElement     `xml:"cmdp:identifiers>cmdp:identifier"`
	Funds         *[]FundingComponent        `xml:"cmdp:funding>cmdp:funds,omitempty"`
	ContactPerson ContactPersonComponent     `xml:"cmdp:contactPerson"`
	Publishers    []string                   `xml:"cmdp:publishers>cmdp:publisher"`
}

type AuthorComponent struct {
	LastName  string `xml:"cmdp:lastName"`
	FirstName string `xml:"cmdp:firstName,omitempty"`
}

type DatesComponent struct {
	Dates      []formats.TypedElement `xml:"cmdp:date,omitempty"` // type is value scheme
	DateIssued string                 `xml:"cmdp:dateIssued,omitempty"`
}

type FundingComponent struct {
	Organization string `xml:"cmdp:organization"`
	Code         string `xml:"cmdp:code"` // grant or project id
	ProjectName  string `xml:"cmdp:projectName"`
	FundsType    string `xml:"cmdp:fundsType"`
}

type ContactPersonComponent struct {
	LastName    string `xml:"cmdp:lastName"`
	FirstName   string `xml:"cmdp:firstName"`
	Email       string `xml:"cmdp:email"`
	Affiliation string `xml:"cmdp:affiliation"`
}
