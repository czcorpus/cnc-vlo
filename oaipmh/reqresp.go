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

import (
	"encoding/xml"
	"time"

	"github.com/rs/zerolog/log"
)

// note - omitempties are optional

type OAIPMHRequest struct {
	URL string `xml:",chardata"`

	Verb            Verb   `xml:"verb,attr,omitempty"`
	Identifier      string `xml:"identifier,attr,omitempty"`
	MetadataPrefix  string `xml:"metadataPrefix,attr,omitempty"`
	From            string `xml:"from,attr,omitempty"`
	Until           string `xml:"until,attr,omitempty"`
	Set             string `xml:"set,attr,omitempty"`
	ResumptionToken string `xml:"resumptionToken,attr,omitempty"`
}

type OAIPMHResponse struct {
	XMLName           xml.Name `xml:"OAI-PMH"`
	XMLNS             string   `xml:"xmlns,attr"`
	XMLNSXSI          string   `xml:"xmlns:xsi,attr"`
	XSISchemaLocation string   `xml:"xsi:schemaLocation,attr"`

	ResponseDate time.Time      `xml:"responseDate"`
	Request      *OAIPMHRequest `xml:"request"`
	Errors       OAIPMHErrors   `xml:"error,omitempty"`

	Identify            *OAIPMHIdentify         `xml:"Identify,omitempty"`
	GetRecord           *OAIPMHRecord           `xml:"GetRecord>record,omitempty"`
	ListMetadataFormats *[]OAIPMHMetadataFormat `xml:"ListMetadataFormats>metadataFormat,omitempty"`
	ListIdentifiers     *[]OAIPMHRecordHeader   `xml:"ListIdentifiers>header,omitempty"`
	ListRecords         *[]OAIPMHRecord         `xml:"ListRecords>record,omitempty"`
	ListSets            *[]OAIPMHSet            `xml:"ListSets>set,omitempty"`

	ProtocolVersion string `xml:"-"`
}

type OAIPMHErrors []OAIPMHError

func (r *OAIPMHErrors) Add(code OAIPMHErrorCode, message string) {
	*r = append(*r, OAIPMHError{Code: code.String(), Message: message})
}

func (r *OAIPMHErrors) HasErrors() bool {
	if len(*r) > 0 {
		log.Debug().Any("errors", r).Send()
		return true
	}
	return false
}

func NewOAIPMHResponse(request *OAIPMHRequest) *OAIPMHResponse {
	return &OAIPMHResponse{
		XMLNS:             "http://www.openarchives.org/OAI/2.0/",
		XMLNSXSI:          "http://www.w3.org/2001/XMLSchema-instance",
		XSISchemaLocation: "http://www.openarchives.org/OAI/2.0/ http://www.openarchives.org/OAI/2.0/OAI-PMH.xsd",
		ResponseDate:      time.Now(),
		Request:           request,
		ProtocolVersion:   "2.0",
	}
}
