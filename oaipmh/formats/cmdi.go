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

const CMDIMetadataPrefix = "cmdi"

// note - omitempties are optional

type CMDIFormat struct {
	XMLName           xml.Name `xml:"cmd:CMD"`
	XMLNSXSI          string   `xml:"xmlns:xsi,attr"`
	XMLNSCMD          string   `xml:"xmlns:cmd,attr"`
	XMLNSCMDP         string   `xml:"xmlns:cmdp,attr"`
	XSISchemaLocation string   `xml:"xsi:schemaLocation,attr"`
	Version           string   `xml:"CMDVersion,attr"`

	Header     CMDIHeader    `xml:"cmd:Header"`
	Resources  CMDIResources `xml:"cmd:Resources"`
	IsPartOf   *[]string     `xml:"cmd:IsPartOfList>IsPartOf,omitempty"`
	Components any           `xml:"cmd:Components"`
}

// --------------------- Header ---------------------
type CMDIHeader struct {
	MdCreator               []string `xml:"cmd:MdCreator,omitempty"`
	MdCreationDate          string   `xml:"cmd:MdCreationDate,omitempty"`
	MdSelfLink              string   `xml:"cmd:MdSelfLink,omitempty"`
	MdProfile               string   `xml:"cmd:MdProfile"`
	MdCollectionDisplayName string   `xml:"cmd:MdCollectionDisplayName,omitempty"`
}

// --------------------- Resources ------------------

type CMDIResources struct {
	ResourceProxyList    []CMDIResourceProxy    `xml:"cmd:ResourceProxyList>cmd:ResourceProxy,omitempty"`
	JournalFileProxyList []string               `xml:"cmd:JournalFileProxyList>cmd:JournaFileProxy>cmd:ResourceRef,omitempty"`
	ResourceRelationList []CMDIResourceRelation `xml:"cmd:ResourceRelationList>cmd:ResourceRelation,omitempty"`
}

type CMDIResourceProxy struct {
	ID           string           `xml:"id,attr"`
	ResourceType CMDIResourceType `xml:"cmd:ResourceType"`
	ResourceRef  string           `xml:"cmd:ResourceRef"`
}

type CMDIResourceType struct {
	MimeType string `xml:"mimetype,attr"`
	Value    string `xml:",chardata"`
}

type CMDIResourceRelation struct {
	RelationType CMDIRelationType `xml:"cmd:RelationType"`
	Resources    [2]CMDIResource  `xml:"cmd:Resource"`
}

type CMDIRelationType struct {
	ConceptLink string `xml:"cmd:ConceptLink,attr"`
	Value       string `xml:",chardata"`
}

type CMDIResource struct {
	Ref  string            `xml:"ref,attr"`
	Role *CMDIRelationType `xml:"cmd:Role,omitempty"`
}

// -------------------------------------------------------

type CMDIProfile interface {
	GetSchemaURL() string
}

func NewCMDI(profile CMDIProfile) CMDIFormat {
	return CMDIFormat{
		XMLNSXSI:  "http://www.w3.org/2001/XMLSchema-instance",
		XMLNSCMD:  "http://www.clarin.eu/cmd/1",
		XMLNSCMDP: profile.GetSchemaURL(),
		XSISchemaLocation: strings.Join([]string{
			"http://www.clarin.eu/cmd/1",
			"http://www.clarin.eu/cmd/1/xsd/cmd-envelop.xsd",
			profile.GetSchemaURL(),
		}, " "),
		Version:    "1.2",
		Header:     CMDIHeader{MdProfile: profile.GetSchemaURL()},
		Components: profile,
	}
}

func GetCMDIFormat(profile CMDIProfile) oaipmh.OAIPMHMetadataFormat {
	return oaipmh.OAIPMHMetadataFormat{
		MetadataPrefix:    CMDIMetadataPrefix,
		Schema:            profile.GetSchemaURL(),
		MetadataNamespace: "http://www.clarin.eu/cmd/",
	}
}
