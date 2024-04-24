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

func getAuthorList(data *cncdb.DBData) []components.AuthorComponent {
	authors := []components.AuthorComponent{}
	for _, author := range strings.Split(strings.ReplaceAll(data.Authors, "\r\n", "\n"), "\n") {
		sAuthor := strings.Split(strings.Trim(author, " "), " ")
		if len(sAuthor) == 1 {
			authors = append(authors, components.AuthorComponent{LastName: sAuthor[0]})
		} else if len(sAuthor) > 1 {
			authors = append(authors, components.AuthorComponent{FirstName: sAuthor[0], LastName: sAuthor[1]})
		}
	}
	return authors
}

func (c *CNCHook) dcRecordFromData(data *cncdb.DBData) oaipmh.OAIPMHRecord {
	recordID := fmt.Sprint(data.ID)
	metadata := formats.NewDublinCore()
	metadata.Title.Add(data.Title, "en")
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
	metadata.Description.Add(data.Description.String, "en")
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
	profile := &profiles.LindatClarinProfile{
		BibliographicInfo: components.BibliographicInfoComponent{
			Titles: []formats.MultilangElement{
				{Lang: "en", Value: data.Title},
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
		DataInfoInfo: components.DataInfoComponent{
			Type:        data.Type,
			Description: data.Description.String,
		},
		LicenseInfo: []profiles.LicenseElement{
			{URI: data.License},
		},
	}

	if data.Link.String != "" {
		profile.DataInfoInfo.Links = &[]formats.TypedElement{
			{Value: data.Link.String},
		}
	}
	switch MetadataType(data.Type) {
	case CorpusMetadataType:
		profile.DataInfoInfo.SizeInfo = &[]components.SizeComponent{
			{Size: fmt.Sprint(data.CorpusData.Size.Int32), Unit: "words"},
		}
		if data.CorpusData.Locale != nil {
			base, _ := data.CorpusData.Locale.Base()
			profile.DataInfoInfo.Languages = &[]components.LanguageComponent{
				{Name: display.English.Languages().Name(base), Code: base.String()},
			}
		}
	case ServiceMetadataType:
	default:
	}

	metadata := formats.NewCMDI(profile)
	metadata.Header.MdSelfLink = fmt.Sprintf("%s/record/%s?format=cmdi", c.conf.RepositoryInfo.BaseURL, recordID)
	// TODO Clarin requires at least one resource proxy for record to be harvested
	metadata.Resources.ResourceProxyList = []formats.CMDIResourceProxy{
		{ID: "TODO", ResourceType: formats.CMDIResourceType{MimeType: "TODO", Value: "TODO"}, ResourceRef: "TODO"},
	}
	record := oaipmh.NewOAIPMHRecord(metadata)
	record.Header.Datestamp = data.Date.In(time.UTC)
	record.Header.Identifier = recordID
	return record
}
