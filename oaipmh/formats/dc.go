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

type DataArray []MultilangElement

func (d *DataArray) Add(value string, lang string) {
	*d = append(*d, MultilangElement{Value: value, Lang: lang})
}

// note - omitempties are optional

type DublinCore struct {
	XMLName           xml.Name `xml:"oai_dc:dc"`
	XMLNSOAIDC        string   `xml:"xmlns:oai_dc,attr"`
	XMLNSDC           string   `xml:"xmlns:dc,attr"`
	XMLNSXSI          string   `xml:"xmlns:xsi,attr"`
	XSISchemaLocation string   `xml:"xsi:schemaLocation,attr"`

	Title       DataArray `xml:"dc:title"`
	Creator     DataArray `xml:"dc:creator"`
	Subject     DataArray `xml:"dc:subject"`
	Description DataArray `xml:"dc:description"`
	Publisher   DataArray `xml:"dc:publisher"`
	Contributor DataArray `xml:"dc:contributor"`
	Date        DataArray `xml:"dc:date"` // ISO 8601
	Type        DataArray `xml:"dc:type"`
	Format      DataArray `xml:"dc:format"`
	Identifier  DataArray `xml:"dc:identifier"`
	Source      DataArray `xml:"dc:source"`
	Language    DataArray `xml:"dc:language"` // ISO 639 + optionally ISO 3166
	Relation    DataArray `xml:"dc:relation"`
	Coverage    DataArray `xml:"dc:coverage"`
	Rights      DataArray `xml:"dc:rights"`
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
