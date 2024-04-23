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

package cnchook

import (
	"fmt"
	"net/http"
	"time"

	"github.com/czcorpus/cnc-vlo/cncdb"
	"github.com/czcorpus/cnc-vlo/cnchook/profiles"
	"github.com/czcorpus/cnc-vlo/cnf"
	"github.com/czcorpus/cnc-vlo/oaipmh"
	"github.com/czcorpus/cnc-vlo/oaipmh/formats"
	"github.com/rs/zerolog/log"
)

type CNCHook struct {
	conf *cnf.Conf
	db   *cncdb.CNCMySQLHandler
}

func (c *CNCHook) Identify() oaipmh.ResultWrapper[oaipmh.OAIPMHIdentify] {
	earliestDatestamp, err := c.db.GetFirstDate()
	result := oaipmh.NewResultWrapper(
		oaipmh.OAIPMHIdentify{
			RepositoryName:    c.conf.RepositoryInfo.Name,
			BaseURL:           c.conf.RepositoryInfo.BaseURL,
			AdminEmail:        c.conf.RepositoryInfo.AdminEmail,
			EarliestDatestamp: earliestDatestamp.In(time.UTC),
			DeletedRecord:     "no",
			Granularity:       "YYYY-MM-DDThh:mm:ssZ",
		},
	)
	if err != nil {
		log.Error().Err(err).Send()
		result.HTTPCode = http.StatusInternalServerError
	}
	return result
}

func (c *CNCHook) ListMetadataFormats(req oaipmh.OAIPMHRequest) oaipmh.ResultWrapper[[]oaipmh.OAIPMHMetadataFormat] {
	ans := oaipmh.NewResultWrapper(
		[]oaipmh.OAIPMHMetadataFormat{
			formats.GetDublinCoreFormat(),
			formats.GetCMDIFormat(&profiles.LindatClarinProfile{}),
		},
	)
	if req.Identifier != "" {
		exists, err := c.db.IdentifierExists(req.Identifier)
		if err != nil {
			log.Error().Err(err).Send()
			ans.HTTPCode = http.StatusInternalServerError
			return ans

		} else if !exists {
			ans.Errors.Add(oaipmh.ErrorCodeIDDoesNotExist, fmt.Sprintf("Result for ID = %s not found", req.Identifier))
			ans.HTTPCode = http.StatusNotFound
			return ans
		}
	}
	return ans
}

func (c *CNCHook) GetRecord(req oaipmh.OAIPMHRequest) oaipmh.ResultWrapper[oaipmh.OAIPMHRecord] {
	ans := oaipmh.NewResultWrapper(oaipmh.OAIPMHRecord{})
	data, err := c.db.GetRecordInfo(req.Identifier)
	if err != nil {
		log.Error().Err(err).Send()
		ans.HTTPCode = http.StatusInternalServerError
		return ans

	} else if data == nil {
		ans.Errors.Add(oaipmh.ErrorCodeIDDoesNotExist, fmt.Sprintf("Result for ID = %s not found", req.Identifier))
		ans.HTTPCode = http.StatusNotFound
		return ans
	}

	switch req.MetadataPrefix {
	case formats.DublinCoreMetadataPrefix:
		ans.Data = c.dcRecordFromData(data)
	case formats.CMDIMetadataPrefix:
		ans.Data = c.cmdiLindatClarinRecordFromData(data)
	default:
		ans.Errors.Add(oaipmh.ErrorCodeCannotDisseminateFormat, "Unknown metadata format")
		ans.HTTPCode = http.StatusBadRequest
	}
	return ans
}

// same as ListRecords but returns only RecordHeaders
func (c *CNCHook) ListIdentifiers(req oaipmh.OAIPMHRequest) oaipmh.ResultWrapper[[]oaipmh.OAIPMHRecordHeader] {
	ans := oaipmh.NewResultWrapper([]oaipmh.OAIPMHRecordHeader{})
	data, err := c.db.ListRecordInfo(req.From, req.Until)
	if err != nil {
		log.Error().Err(err).Send()
		ans.HTTPCode = http.StatusInternalServerError
		return ans
	}
	if len(data) == 0 {
		ans.Errors.Add(oaipmh.ErrorCodeNoRecordsMatch, "No records")
		return ans
	}
	switch req.MetadataPrefix {
	case formats.DublinCoreMetadataPrefix:
		for _, d := range data {
			ans.Data = append(ans.Data, *c.dcRecordFromData(&d).Header)
		}
	case formats.CMDIMetadataPrefix:
		for _, d := range data {
			ans.Data = append(ans.Data, *c.cmdiLindatClarinRecordFromData(&d).Header)
		}
	default:
		ans.Errors.Add(oaipmh.ErrorCodeCannotDisseminateFormat, "Unknown metadata format")
		ans.HTTPCode = http.StatusBadRequest
	}
	return ans
}

func (c *CNCHook) ListRecords(req oaipmh.OAIPMHRequest) oaipmh.ResultWrapper[[]oaipmh.OAIPMHRecord] {
	ans := oaipmh.NewResultWrapper([]oaipmh.OAIPMHRecord{})
	data, err := c.db.ListRecordInfo(req.From, req.Until)
	if err != nil {
		log.Error().Err(err).Send()
		ans.HTTPCode = http.StatusInternalServerError
		return ans
	}
	if len(data) == 0 {
		ans.Errors.Add(oaipmh.ErrorCodeNoRecordsMatch, "No records")
		return ans
	}
	switch req.MetadataPrefix {
	case formats.DublinCoreMetadataPrefix:
		for _, d := range data {
			ans.Data = append(ans.Data, c.dcRecordFromData(&d))
		}
	case formats.CMDIMetadataPrefix:
		for _, d := range data {
			ans.Data = append(ans.Data, c.cmdiLindatClarinRecordFromData(&d))
		}
	default:
		ans.Errors.Add(oaipmh.ErrorCodeCannotDisseminateFormat, "Unknown metadata format")
		ans.HTTPCode = http.StatusBadRequest
	}
	return ans
}

func (c *CNCHook) ListSets(req oaipmh.OAIPMHRequest) oaipmh.ResultWrapper[[]oaipmh.OAIPMHSet] {
	return oaipmh.NewResultWrapper([]oaipmh.OAIPMHSet{})
}

func (c *CNCHook) SupportsSets() bool {
	return false
}

func (c *CNCHook) SupportedMetadataPrefixes() []string {
	return []string{
		formats.DublinCoreMetadataPrefix,
		formats.CMDIMetadataPrefix,
	}
}

func NewCNCHook(conf *cnf.Conf, db *cncdb.CNCMySQLHandler) *CNCHook {
	return &CNCHook{
		conf: conf,
		db:   db,
	}
}
