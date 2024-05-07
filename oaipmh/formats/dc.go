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

package formats

import (
	"encoding/xml"
	"strings"

	"github.com/czcorpus/cnc-vlo/oaipmh"
)

const DublinCoreMetadataPrefix = "oai_dc"

// note - omitempties are optional

type DublinCore struct {
	XMLName           xml.Name `xml:"oai_dc:dc"`
	XMLNSOAIDC        string   `xml:"xmlns:oai_dc,attr"`
	XMLNSDC           string   `xml:"xmlns:dc,attr"`
	XMLNSXSI          string   `xml:"xmlns:xsi,attr"`
	XSISchemaLocation string   `xml:"xsi:schemaLocation,attr"`

	Title       MultilangArray `xml:"dc:title"`
	Creator     MultilangArray `xml:"dc:creator"`
	Subject     MultilangArray `xml:"dc:subject"`
	Description MultilangArray `xml:"dc:description"`
	Publisher   MultilangArray `xml:"dc:publisher"`
	Contributor MultilangArray `xml:"dc:contributor"`
	Date        MultilangArray `xml:"dc:date"` // ISO 8601
	Type        MultilangArray `xml:"dc:type"`
	Format      MultilangArray `xml:"dc:format"`
	Identifier  MultilangArray `xml:"dc:identifier"`
	Source      MultilangArray `xml:"dc:source"`
	Language    MultilangArray `xml:"dc:language"` // ISO 639 + optionally ISO 3166
	Relation    MultilangArray `xml:"dc:relation"`
	Coverage    MultilangArray `xml:"dc:coverage"`
	Rights      MultilangArray `xml:"dc:rights"`
}

func NewDublinCore() DublinCore {
	return DublinCore{
		XMLNSOAIDC: "http://www.openarchives.org/OAI/2.0/oai_dc/",
		XMLNSDC:    "http://purl.org/dc/elements/1.1/",
		XMLNSXSI:   "http://www.w3.org/2001/XMLSchema-instance",
		XSISchemaLocation: strings.Join([]string{
			"http://www.openarchives.org/OAI/2.0/oai_dc/",
			"http://www.openarchives.org/OAI/2.0/oai_dc.xsd",
		}, " "),
	}
}

func GetDublinCoreFormat() oaipmh.OAIPMHMetadataFormat {
	return oaipmh.OAIPMHMetadataFormat{
		MetadataPrefix:    DublinCoreMetadataPrefix,
		Schema:            "http://www.openarchives.org/OAI/2.0/oai_dc.xsd",
		MetadataNamespace: "http://www.openarchives.org/OAI/2.0/oai_dc/",
	}
}
