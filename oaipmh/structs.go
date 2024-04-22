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

package oaipmh

// wrapper to be able to embed custom element with name defined by XMLName
type ElementWrapper struct {
	Value any
}

// note - omitempties are optional

type OAIPMHRecordHeader struct {
	Status     string   `xml:"status,attr,omitempty"` // only `deleted` status
	Identifier string   `xml:"identifier"`            // URL
	Datestamp  string   `xml:"datestamp"`             // creation, modification or deletion of the record for the purpose of selective harvesting
	SetSpec    []string `xml:"setSpec,omitempty"`
}

// ----------------------- Identify ---------------------------

type OAIPMHIdentify struct {
	RepositoryName    string           `xml:"repositoryName"`
	BaseURL           string           `xml:"baseURL"`
	AdminEmail        []string         `xml:"adminEmail"`
	EarliestDatestamp string           `xml:"earliestDatestamp"`
	DeletedRecord     string           `xml:"deletedRecord"` // are we tracking deleted records no/transient/persistent?
	Granularity       string           `xml:"granularity"`   // all repositories must support YYYY-MM-DD, extra YYYY-MM-DDThh:mm:ssZ
	Compression       string           `xml:"compression,omitempty"`
	Description       []ElementWrapper `xml:"description,omitempty"`

	ProtocolVersion string `xml:"protocolVersion"` // filled automatically by handler
}

// --------------------- ListMetadataFormats ------------------

type OAIPMHMetadataFormat struct {
	MetadataPrefix    string `xml:"metadataPrefix"`
	Schema            string `xml:"schema"`
	MetadataNamespace string `xml:"metadataNamespace"`
}

// ----------------------- GetRecord/ListRecords --------------

type OAIPMHRecord struct {
	Header   *OAIPMHRecordHeader `xml:"header"`
	Metadata *ElementWrapper     `xml:"metadata,omitempty"`
}

func NewOAIPMHRecord(metadata any) OAIPMHRecord {
	return OAIPMHRecord{
		Header:   &OAIPMHRecordHeader{},
		Metadata: &ElementWrapper{Value: metadata},
	}
}

// ----------------------- ListSets ---------------------

type OAIPMHSet struct {
	SetSpec        string          `xml:"setSpec"`
	SetName        string          `xml:"setName"`
	SetDescription *ElementWrapper `xml:"setDescription,omitempty"`
}
