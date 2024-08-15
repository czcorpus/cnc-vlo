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
	"strings"
	"time"

	"github.com/czcorpus/cnc-vlo/cncdb"
	"github.com/czcorpus/cnc-vlo/cnchook/profiles"
	"github.com/czcorpus/cnc-vlo/cnchook/profiles/components"
	"github.com/czcorpus/cnc-vlo/oaipmh"
	"github.com/czcorpus/cnc-vlo/oaipmh/formats"
	"golang.org/x/text/language/display"
)

func (c *CNCHook) dcRecordFromData(data *cncdb.DBData) oaipmh.OAIPMHRecord {
	recordID := fmt.Sprint(data.ID)
	metadata := formats.NewDublinCore()
	metadata.Title.Add(data.TitleEN, "en")
	metadata.Title.Add(data.TitleCS, "cs")
	if data.DescCS.Valid {
		metadata.Description.Add(data.DescCS.String, "cs")
	}
	if data.DescEN.Valid {
		metadata.Description.Add(data.DescEN.String, "en")
	}
	metadata.Date.Add(data.Date.In(time.UTC).Format(time.RFC3339), "")
	for _, author := range getAuthorList(data) {
		if author.FirstName == "" {
			metadata.Creator.Add(author.LastName, "")
		} else {
			metadata.Creator.Add(author.FirstName+" "+author.LastName, "")
		}
	}
	metadata.Identifier.Add(data.Name, "")
	metadata.Type.Add(data.Type, "")
	metadata.Rights.Add(data.License, "")

	switch MetadataType(data.Type) {
	case CorpusMetadataType:
		if data.CorpusData.Locale != nil {
			base, _ := data.CorpusData.Locale.Base()
			metadata.Language.Add(base.String(), "")
		}
	case ServiceMetadataType:
	default:
	}

	record := oaipmh.NewOAIPMHRecord(metadata)
	record.Header.Datestamp = data.Date.In(time.UTC)
	record.Header.Identifier = recordID
	return record
}

func (c *CNCHook) cmdiLindatClarinRecordFromData(data *cncdb.DBData) oaipmh.OAIPMHRecord {
	recordID := fmt.Sprint(data.ID)
	profile := &profiles.CNCResourceProfile{
		BibliographicInfo: components.BibliographicInfoComponent{
			Titles: formats.MultilangArray{
				{Lang: "en", Value: data.TitleEN},
				{Lang: "cs", Value: data.TitleCS},
			},
			Identifiers: []formats.TypedElement{
				{Value: data.Name},
			},
			Authors: getAuthorList(data),
			ContactPerson: components.ContactPersonComponent{
				LastName:    data.ContactPerson.Lastname,
				FirstName:   data.ContactPerson.Firstname,
				Email:       data.ContactPerson.Email,
				Affiliation: data.ContactPerson.Affiliation.String,
			},
			Publishers: []string{
				c.conf.MetadataValues.Publisher,
			},
		},
		DataInfo: components.DataInfoComponent{
			Type: data.Type,
			Description: formats.MultilangArray{
				{Lang: "en", Value: data.DescEN.String},
				{Lang: "cs", Value: data.DescCS.String},
			},
		},
		LicenseInfo: []profiles.LicenseElement{
			{URI: data.License},
		},
	}
	if data.DateIssued == "" {
		profile.BibliographicInfo.Dates = &components.DatesComponent{DateIssued: data.DateIssued}
	}
	metadata := formats.NewCMDI(profile)
	metadata.Header.MdSelfLink = fmt.Sprintf("%s/record/%s?format=cmdi", c.conf.RepositoryInfo.BaseURL, recordID)

	switch MetadataType(data.Type) {
	case CorpusMetadataType:
		profile.DataInfo.SizeInfo = &[]components.SizeComponent{
			{Size: fmt.Sprint(data.CorpusData.Size.Int64), Unit: "words"},
		}
		if data.CorpusData.Locale != nil {
			base, _ := data.CorpusData.Locale.Base()
			profile.DataInfo.Languages = &[]components.LanguageComponent{
				{Name: display.English.Languages().Name(base), Code: base.String()},
			}
		}
		if data.CorpusData.Keywords.String != "" {
			keywords := strings.Split(data.CorpusData.Keywords.String, ",")
			profile.DataInfo.Keywords = &keywords
		}
		metadata.Resources.ResourceProxyList = append(
			metadata.Resources.ResourceProxyList,
			formats.CMDIResourceProxy{
				ID:           fmt.Sprintf("sp_%s", recordID),
				ResourceType: formats.CMDIResourceType{MimeType: "text/html", Value: formats.RTSearchPage},
				ResourceRef:  getKontextPath(data.Name),
			},
		)

	case ServiceMetadataType:
	default:
	}

	// insert link if available
	if data.Link.String != "" {
		link := data.Link.String
		// generate path to english version wiki
		if strings.Contains(link, "wiki.korpus.cz") {
			link = strings.ReplaceAll(link, "/cnk:", "/en:cnk:")
		}
		metadata.Resources.ResourceProxyList = append(
			metadata.Resources.ResourceProxyList,
			formats.CMDIResourceProxy{
				ID:           fmt.Sprintf("uri_%s", recordID),
				ResourceType: formats.CMDIResourceType{MimeType: "text/html", Value: formats.RTResource},
				ResourceRef:  link,
			},
		)
	}

	record := oaipmh.NewOAIPMHRecord(metadata)
	record.Header.Datestamp = data.Date.In(time.UTC)
	record.Header.Identifier = recordID
	return record
}
