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
	"fmt"
	"net/url"
)

const (
	ArgVerb            string = "verb"            // always required
	ArgIdentifier      string = "identifier"      // req GetRecord, op ListMetadataFormats
	ArgMetadataPrefix  string = "metadataPrefix"  // req GetRecord, req ListIdentifiers, req ListRecords
	ArgFrom            string = "from"            // op ListIdentifiers, op ListRecords
	ArgUntil           string = "until"           // op ListIdentifiers, op ListRecords
	ArgSet             string = "set"             // op ListIdentifiers, op ListRecords
	ArgResumptionToken string = "resumptionToken" // ListIdentifiers, ListRecords, ListSets

	VerbIdentify            Verb = "Identify"
	VerbGetRecord           Verb = "GetRecord"
	VerbListIdentifiers     Verb = "ListIdentifiers"
	VerbListMetadataFormats Verb = "ListMetadataFormats"
	VerbListRecords         Verb = "ListRecords"
	VerbListSets            Verb = "ListSets"
)

// ----

type Verb string

func (v Verb) String() string {
	return string(v)
}

func (v Verb) Validate() error {
	if v == VerbGetRecord || v == VerbIdentify ||
		v == VerbListIdentifiers || v == VerbListMetadataFormats ||
		v == VerbListRecords || v == VerbListSets {
		return nil
	}
	return fmt.Errorf("unknown verb: %s", v)
}

func (v Verb) ValidateArg(arg string) bool {
	switch v {
	case VerbGetRecord:
		return arg == ArgVerb || arg == ArgIdentifier || arg == ArgMetadataPrefix
	case VerbListIdentifiers:
		return arg == ArgVerb || arg == ArgMetadataPrefix || arg == ArgFrom || arg == ArgUntil || arg == ArgSet || arg == ArgResumptionToken
	case VerbListMetadataFormats:
		return arg == ArgVerb || arg == ArgIdentifier
	case VerbListRecords:
		return arg == ArgVerb || arg == ArgMetadataPrefix || arg == ArgFrom || arg == ArgUntil || arg == ArgSet || arg == ArgResumptionToken
	case VerbListSets:
		return arg == ArgVerb || arg == ArgResumptionToken
	default: // VerbIdentify
		return arg == ArgVerb
	}
}

func (v Verb) ValidateRequiredArgs(args url.Values) string {
	reqArgs := []string{ArgVerb}
	switch v {
	case VerbGetRecord:
		reqArgs = append(reqArgs, ArgIdentifier, ArgMetadataPrefix)
	case VerbListIdentifiers:
		reqArgs = append(reqArgs, ArgMetadataPrefix)
	case VerbListRecords:
		reqArgs = append(reqArgs, ArgMetadataPrefix)
	}
	for _, arg := range reqArgs {
		if !args.Has(arg) {
			return arg
		}
	}
	return ""
}
