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

type OAIPMHErrorCode string

const (
	// http://www.openarchives.org/OAI/openarchivesprotocol.html#ErrorConditions
	ErrorCodeBadArgument             OAIPMHErrorCode = "badArgument"
	ErrorCodeBadResumptionToken      OAIPMHErrorCode = "badResumptionToken"
	ErrorCodeBadVerb                 OAIPMHErrorCode = "badVerb"
	ErrorCodeCannotDisseminateFormat OAIPMHErrorCode = "cannotDisseminateFormat"
	ErrorCodeIDDoesNotExist          OAIPMHErrorCode = "idDoesNotExist"
	ErrorCodeNoRecordsMatch          OAIPMHErrorCode = "noRecordsMatch"
	ErrorCodeNoMetadataFormats       OAIPMHErrorCode = "noMetadataFormats"
	ErrorCodeNoSetHierarchy          OAIPMHErrorCode = "noSetHierarchy"
)

func (e OAIPMHErrorCode) String() string {
	return string(e)
}

type OAIPMHError struct {
	Code    string `xml:"code,attr"`
	Message string `xml:",chardata"`
}
