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

type DataInfoComponent struct {
	Type           string                   `xml:"cmdp:type"`                   // e.g. corpus, tool
	DetailedType   string                   `xml:"cmdp:detailedType,omitempty"` // Further specification of the type
	Description    formats.MultilangArray   `xml:"cmdp:description"`
	Languages      *[]LanguageComponent     `xml:"cmdp:languages>cmdp:language,omitempty"`
	Keywords       *[]string                `xml:"cmdp:keywords>cmdp:keyword,omitempty"`
	Links          *[]formats.TypedElement  `xml:"cmdp:links>cmdp:link,omitempty"` // demo url, documentation url
	SizeInfo       *[]SizeComponent         `xml:"cmdp:sizeInfo>cmdp:size,omitempty"`
	Formats        *[]FormatComponent       `xml:"cmdp:formats>cmdp:format,omitempty"`
	Requirements   *[]string                `xml:"cmdp:requirements>cmdp:requirement,omitempty"` // e.g. OS, prerequisities
	CollectionInfo *CollectionInfoComponent `xml:"cmdp:collectionInfo,omitempty"`
	AnnotationInfo *[]string                `xml:"cmdp:annotationInfo>cmdp:annotationType,omitempty"` // tags, lemmas, phrase alignment, coreference, ...
}

type LanguageComponent struct {
	Name string `xml:"cmdp:name"`
	Code string `xml:"cmdp:code"`
}

type SizeComponent struct {
	Size string `xml:"cmdp:size"`
	Unit string `xml:"cmdp:unit"`
}

type FormatComponent struct {
	Type          string `xml:"cmdp:type,attr,omitempty"`
	Name          string `xml:"cmdp:name,omitempty"`
	Medium        string `xml:"cmdp:medium,omitempty"` // text, audio, ...
	Documentation string `xml:"cmdp:documentation,omitempty"`
	Description   string `xml:"cmdp:description,omitempty"` // e.g. vertical format, where each line is "form/lemma/tag"
}

type CollectionInfoComponent struct {
	TimePeriods []string `xml:"cmdp:timePeriod,omitempty"`        // When the data were gathered, which era do they come from
	Places      []string `xml:"cmdp:place,omitempty"`             // The origin of the data. e.g. The data were gathered in Bohemia
	Forms       []string `xml:"cmdp:forms>cmdp:form,omitempty"`   // spoken, written,...
	Genres      []string `xml:"cmdp:genres>cmdp:genre,omitempty"` // fiction, news, blog
}
