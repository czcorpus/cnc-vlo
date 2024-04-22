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
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
)

func getTypedArg[T ~string](args url.Values, name string) T {
	return T(args.Get(name))
}

func writeXMLResponse(w http.ResponseWriter, code int, value any) {
	xmlAns, err := xml.Marshal(value)
	if err != nil {
		log.Err(err).Msg("failed to encode a result to XML")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	_, err = w.Write([]byte(xml.Header + "\n" + string(xmlAns)))
	if err != nil {
		log.Err(err).Msg("failed to write XML to response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/xml")
}
