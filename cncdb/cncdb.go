// Copyright 2024 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
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

package cncdb

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
)

// DBOverrides handles differences between KonText default
// database schema and the CNC-one which is slightly different
type DBOverrides struct {
	CorporaTableName      string `json:"corporaTableName"`
	UserTableName         string `json:"userTableName"`
	UserTableFirstNameCol string `json:"userTableFirstNameCol"`
	UserTableLastNameCol  string `json:"userTableLastNameCol"`
}

type CNCMySQLHandler struct {
	conn             *sql.DB
	overrides        DBOverrides
	publicCorplistID int
}

type DBData struct {
	ID            int
	Date          time.Time
	Hosted        bool
	Type          string
	Name          string
	DescEN        sql.NullString
	DescCS        sql.NullString
	DateIssued    string
	TitleEN       string
	TitleCS       string
	Link          sql.NullString
	License       string
	Authors       string
	ContactPerson ContactPersonData
	CorpusData    CorpusData
}

type ContactPersonData struct {
	Firstname   string
	Lastname    string
	Email       string
	Affiliation sql.NullString
}

type CorpusData struct {
	Size     sql.NullInt64
	Locale   *language.Tag
	Keywords sql.NullString
}

func (c *CNCMySQLHandler) GetFirstDate() (time.Time, error) {
	var date time.Time
	row := c.conn.QueryRow("SELECT MIN(created) FROM vlo_metadata_common")
	err := row.Scan(&date)
	return date, err
}

func (c *CNCMySQLHandler) IdentifierExists(identifier string) (bool, error) {
	var id int
	row := c.conn.QueryRow(
		fmt.Sprintf(
			"SELECT m.id FROM vlo_metadata_common AS m "+
				"LEFT JOIN vlo_metadata_corpus AS mc ON m.corpus_metadata_id = mc.id "+
				"LEFT JOIN %s AS c ON m.corpus_name = c.name "+
				"LEFT JOIN corplist_corpus AS cc ON c.id = cc.corpus_id "+
				"WHERE m.id = ? AND m.deleted = FALSE "+
				"AND ((m.type = 'corpus' AND cc.corplist_id = ?) OR m.type != 'corpus')",
			c.overrides.CorporaTableName,
		),
		identifier, c.publicCorplistID,
	)
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check identifier existence record info: %w", err)
	}
	return true, nil
}

func (c *CNCMySQLHandler) parseLocale(loc string) (ans language.Tag, err error) {
	tmp := strings.Split(loc, ".")
	base := tmp[0]
	ans, err = language.Parse(base)
	if err != nil {
		log.Error().
			Err(err).
			Str("value", loc).
			Msg("Failed to parse database language record. Trying partial parsing.")
		tmp := strings.Split(loc, "_")
		if len(tmp) == 0 {
			tmp = strings.Split(loc, "-")
		}
		if len(tmp) != 2 {
			err = fmt.Errorf("unable to parse locale %s", loc)
			return
		}
		ans, err = language.Parse(tmp[0])
		return
	}
	return
}

func (c *CNCMySQLHandler) GetRecordInfo(identifier string) (*DBData, error) {
	var data DBData
	var locale sql.NullString

	row := c.conn.QueryRow(
		fmt.Sprintf(
			"SELECT "+
				"m.id, "+
				"GREATEST(m.created, m.updated), "+
				"m.hosted, "+
				"m.type, "+
				"m.desc_en, "+
				"m.desc_cs, "+
				"m.date_issued, "+
				"m.license_info, "+
				"m.authors, "+
				"u.%s, "+
				"u.%s, "+
				"u.email, "+
				"u.affiliation, "+
				"COALESCE(c.name, ms.name), "+
				"COALESCE(rc.name, c.name, ms.name), "+
				"COALESCE(rc.name, c.name, ms.name), "+
				"COALESCE(c.web, ms.link), "+
				"c.size, c.locale, GROUP_CONCAT(k.label_en ORDER BY k.display_order SEPARATOR ',') "+
				"FROM vlo_metadata_common AS m "+
				"LEFT JOIN vlo_metadata_corpus AS mc ON m.corpus_metadata_id = mc.id "+
				"LEFT JOIN vlo_metadata_service AS ms ON m.service_metadata_id = ms.id "+
				"LEFT JOIN %s AS c ON mc.corpus_name = c.name "+
				"LEFT JOIN kontext_keyword_corpus AS kc ON kc.corpus_name = c.name "+
				"LEFT JOIN kontext_keyword AS k ON kc.keyword_id = k.id "+
				"LEFT JOIN corplist_corpus AS cc ON c.id = cc.corpus_id "+
				"LEFT JOIN corplist_parallel_corpus AS cpc ON cpc.parallel_corpus_id = c.parallel_corpus_id "+
				"LEFT JOIN registry_conf AS rc ON mc.corpus_name = rc.corpus_name "+
				"JOIN %s AS u ON m.contact_user_id = u.id "+
				"WHERE m.id = ? AND m.deleted = FALSE "+
				"AND ((m.type = 'corpus' AND cc.corplist_id = ?) OR (cpc.corplist_id = ?) OR m.type != 'corpus') "+
				"GROUP BY kc.corpus_name ",
			c.overrides.UserTableFirstNameCol, c.overrides.UserTableLastNameCol,
			c.overrides.CorporaTableName, c.overrides.UserTableName,
		), identifier, c.publicCorplistID, c.publicCorplistID,
	)
	err := row.Scan(
		&data.ID, &data.Date, &data.Hosted, &data.Type, &data.DescEN, &data.DescCS, &data.DateIssued, &data.License, &data.Authors,
		&data.ContactPerson.Firstname, &data.ContactPerson.Lastname, &data.ContactPerson.Email,
		&data.ContactPerson.Affiliation, &data.Name, &data.TitleEN, &data.TitleCS, &data.Link,
		&data.CorpusData.Size, &locale, &data.CorpusData.Keywords,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get record info: %w", err)
	}
	if locale.Valid {
		tag, err := c.parseLocale(locale.String)
		if err != nil {
			return nil, fmt.Errorf("failed to get record info: %w", err)
		}
		data.CorpusData.Locale = &tag
	}
	return &data, nil
}

func (c *CNCMySQLHandler) ListRecordInfo(from *time.Time, until *time.Time) ([]DBData, error) {
	whereClause := []string{
		"m.deleted = ?",
		"((m.type = 'corpus' AND cc.corplist_id = ?) OR cpc.corplist_id = ? OR m.type != 'corpus')",
	}
	whereValues := []any{
		"FALSE",
		c.publicCorplistID,
		c.publicCorplistID,
	}
	if from != nil {
		whereClause = append(whereClause, "GREATEST(m.created, m.updated) >= ?")
		whereValues = append(whereValues, from)
	}
	if until != nil {
		whereClause = append(whereClause, "GREATEST(m.created, m.updated) <= ?")
		whereValues = append(whereValues, until)
	}
	query := fmt.Sprintf(
		"SELECT "+
			"m.id, "+
			" GREATEST(m.created, m.updated), "+
			"m.hosted, "+
			"m.type, "+
			"m.desc_en, "+
			"m.desc_cs, "+
			"m.date_issued, "+
			"m.license_info, "+
			"m.authors, "+
			"u.%s, "+
			"u.%s, "+
			"u.email, "+
			"u.affiliation, "+
			"COALESCE(c.name, ms.name), "+
			"COALESCE(rc.name, c.name, ms.name), "+
			"COALESCE(rc.name, c.name, ms.name), "+
			"COALESCE(c.web, ms.link), "+
			"c.size, "+
			"c.locale, "+
			"GROUP_CONCAT(k.label_en ORDER BY k.display_order SEPARATOR ',') "+
			"FROM vlo_metadata_common AS m "+
			"LEFT JOIN vlo_metadata_corpus AS mc ON m.corpus_metadata_id = mc.id "+
			"LEFT JOIN vlo_metadata_service AS ms ON m.service_metadata_id = ms.id "+
			"LEFT JOIN %s AS c ON mc.corpus_name = c.name "+
			"LEFT JOIN kontext_keyword_corpus AS kc ON kc.corpus_name = c.name "+
			"LEFT JOIN kontext_keyword AS k ON kc.keyword_id = k.id "+
			"LEFT JOIN corplist_corpus AS cc ON c.id = cc.corpus_id "+
			"LEFT JOIN corplist_parallel_corpus AS cpc ON cpc.parallel_corpus_id = c.parallel_corpus_id "+
			"LEFT JOIN registry_conf AS rc ON mc.corpus_name = rc.corpus_name "+
			"JOIN %s AS u ON m.contact_user_id = u.id ",
		c.overrides.UserTableFirstNameCol, c.overrides.UserTableLastNameCol,
		c.overrides.CorporaTableName, c.overrides.UserTableName,
	)
	if len(whereClause) > 0 {
		query += " WHERE " + strings.Join(whereClause, " AND ")
	}
	query += " GROUP BY c.name "
	rows, err := c.conn.Query(query, whereValues...)
	if err != nil {
		return nil, fmt.Errorf("failed to list record info: %w", err)
	}
	results := make([]DBData, 0, 10)
	for rows.Next() {
		var row DBData
		var locale sql.NullString
		err := rows.Scan(
			&row.ID, &row.Date, &row.Hosted, &row.Type, &row.DescEN, &row.DescCS, &row.DateIssued, &row.License, &row.Authors,
			&row.ContactPerson.Firstname, &row.ContactPerson.Lastname, &row.ContactPerson.Email,
			&row.ContactPerson.Affiliation, &row.Name, &row.TitleEN, &row.TitleCS, &row.Link,
			&row.CorpusData.Size, &locale, &row.CorpusData.Keywords,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to list record info: %w", err)
		}
		if locale.String != "" {
			tag, err := c.parseLocale(locale.String)
			if err != nil {
				return nil, fmt.Errorf("failed to list record info: %w", err)
			}
			row.CorpusData.Locale = &tag
		}
		results = append(results, row)
	}
	return results, nil
}

func NewCNCMySQLHandler(cnf DatabaseSetup) (*CNCMySQLHandler, error) {
	conf := mysql.NewConfig()
	conf.Net = "tcp"
	conf.Addr = cnf.Host
	conf.User = cnf.User
	conf.Passwd = cnf.Passwd
	conf.DBName = cnf.Name
	conf.ParseTime = true
	conf.Loc = time.Local
	db, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open CNC DB: %w", err)
	}
	return &CNCMySQLHandler{
		conn:             db,
		overrides:        cnf.Overrides,
		publicCorplistID: cnf.PublicCorplistID,
	}, nil
}
