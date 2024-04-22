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

package cncdb

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type CNCMySQLHandler struct {
	conn             *sql.DB
	corporaTableName string
	userTableName    string
}

type DBData struct {
	ID            int
	Date          string
	Type          string
	Name          string
	Title         string
	Description   sql.NullString
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
	Size   sql.NullInt32
	Locale sql.NullString
}

func (c *CNCMySQLHandler) GetFirstDate() (string, error) {
	var date string
	row := c.conn.QueryRow("SELECT MIN(created) FROM metadata_common")
	err := row.Scan(&date)
	return date, err
}

func (c *CNCMySQLHandler) IdentifierExists(identifier string) (bool, error) {
	var id int
	row := c.conn.QueryRow("SELECT id FROM metadata_common WHERE id = ? AND deleted = FALSE", identifier)
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *CNCMySQLHandler) GetRecordInfo(identifier string) (*DBData, error) {
	var data DBData
	row := c.conn.QueryRow(
		fmt.Sprintf(
			"SELECT m.id, GREATEST(m.created, m.updated), m.type, m.title, m.license_info, m.authors, "+
				"u.firstname, u.lastname, u.email, u.affiliation, "+
				"COALESCE(c.name, ms.name), "+
				"COALESCE(c.description_en, ms.description), "+
				"COALESCE(c.web, ms.link), "+
				"c.size, c.locale "+
				"FROM metadata_common AS m "+
				"LEFT JOIN metadata_corpus AS mc ON m.corpus_metadata_id = mc.id "+
				"LEFT JOIN metadata_service AS ms ON m.service_metadata_id = ms.id "+
				"LEFT JOIN %s AS c ON mc.corpus_name = c.name "+
				"JOIN %s AS u ON m.contact_user_id = u.id "+
				"WHERE m.id = ? AND m.deleted = FALSE",
			c.corporaTableName, c.userTableName,
		), identifier,
	)
	err := row.Scan(
		&data.ID, &data.Date, &data.Type, &data.Title, &data.License, &data.Authors,
		&data.ContactPerson.Firstname, &data.ContactPerson.Lastname, &data.ContactPerson.Email, &data.ContactPerson.Affiliation,
		&data.Name, &data.Description, &data.Link,
		&data.CorpusData.Size, &data.CorpusData.Locale,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &data, nil
}

func (c *CNCMySQLHandler) ListRecordInfo(from string, until string) ([]DBData, error) {
	whereClause := []string{"m.deleted = ?"}
	whereValues := []any{"FALSE"}
	if from != "" {
		whereClause = append(whereClause, "GREATEST(m.created, m.updated) >= ?")
		whereValues = append(whereValues, from)
	}
	if until != "" {
		if strings.Contains(until, "T") {
			whereClause = append(whereClause, "GREATEST(m.created, m.updated) <= ?")
		} else {
			whereClause = append(whereClause, "GREATEST(m.created, m.updated) < ? + INTERVAL 1 DAY")
		}
		whereValues = append(whereValues, until)
	}
	query := fmt.Sprintf(
		"SELECT m.id, GREATEST(m.created, m.updated), m.type, m.title, m.license_info, m.authors, "+
			"u.firstname, u.lastname, u.email, u.affiliation, "+
			"COALESCE(c.name, ms.name), "+
			"COALESCE(c.description_en, ms.description), "+
			"COALESCE(c.web, ms.link), "+
			"c.size, c.locale "+
			"FROM metadata_common AS m "+
			"LEFT JOIN metadata_corpus AS mc ON m.corpus_metadata_id = mc.id "+
			"LEFT JOIN metadata_service AS ms ON m.service_metadata_id = ms.id "+
			"LEFT JOIN %s AS c ON mc.corpus_name = c.name "+
			"JOIN %s AS u ON m.contact_user_id = u.id",
		c.corporaTableName, c.userTableName,
	)
	if len(whereClause) > 0 {
		query += " WHERE " + strings.Join(whereClause, " AND ")
	}
	rows, err := c.conn.Query(query, whereValues...)
	if err != nil {
		return nil, err
	}
	results := make([]DBData, 0, 10)
	for rows.Next() {
		var row DBData
		err := rows.Scan(
			&row.ID, &row.Date, &row.Type, &row.Title, &row.License, &row.Authors,
			&row.ContactPerson.Firstname, &row.ContactPerson.Lastname, &row.ContactPerson.Email, &row.ContactPerson.Affiliation,
			&row.Name, &row.Description, &row.Link,
			&row.CorpusData.Size, &row.CorpusData.Locale,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, nil
}

func NewCNCMySQLHandler(host, user, pass, dbName, corporaTableName, userTableName string) (*CNCMySQLHandler, error) {
	conf := mysql.NewConfig()
	conf.Net = "tcp"
	conf.Addr = host
	conf.User = user
	conf.Passwd = pass
	conf.DBName = dbName
	conf.ParseTime = true
	conf.Loc = time.Local
	db, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		return nil, err
	}
	return &CNCMySQLHandler{
		conn:             db,
		corporaTableName: corporaTableName,
		userTableName:    userTableName,
	}, nil
}
