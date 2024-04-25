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

	"github.com/czcorpus/cnc-vlo/cncdb"
	"github.com/czcorpus/cnc-vlo/cnchook/profiles/components"
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

func getKontextPath(corpusID string) string {
	return fmt.Sprintf("https://www.korpus.cz/kontext/query?corpname=%s", corpusID)
}
