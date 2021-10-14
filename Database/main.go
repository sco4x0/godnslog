package database

import (
	"database/sql"
	"fmt"
	Config "godnslog/Config"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type HttpRecord struct {
	Method   string
	Url      string
	Head     string
	Body     string
	Ip       string
	Datetime string
}
type DnsRecord struct {
	Record   string
	Type     string
	Ip       string
	Datetime string
}

var conn *sql.DB

func init() {
	conn, _ = sql.Open("sqlite3", "./data.db3")
	dnsRecordSql := `CREATE TABLE IF NOT EXISTS dns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		record text,
		type text,
		ip text,
		datetime text 
	)`
	httpRecordSql := `CREATE TABLE IF NOT EXISTS http (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		method text,
		url text,
		head text,
		body text,
		ip text,
		datetime text
	)`
	_, err := conn.Exec(dnsRecordSql)
	if err != nil {
		log.Fatal("DNS record table create failed !")
	}
	_, err = conn.Exec(httpRecordSql)
	if err != nil {
		log.Println(err)
		log.Fatal("HTTP record table create failed !")
	}
}
func delete(table string) (err error) {
	stmt, err := conn.Prepare(fmt.Sprintf("delete from %s", table))
	if err != nil {
		return err
	}
	res, err := stmt.Exec()
	if err != nil {
		return err
	}
	if _, err = res.RowsAffected(); err != nil {
		return err
	}
	return nil
}
func query(table string, where string, page int) (result []interface{}, total int, err error) {
	sql := fmt.Sprintf("select * from %s %s", table, where)
	offset := (page - 1) * Config.Config.HttpOffset
	sql += fmt.Sprintf(" order by id desc limit %d,%d", offset, Config.Config.HttpOffset)
	if rows, err := conn.Query(sql); err == nil {
		defer rows.Close()
		switch table {
		case "dns":
			for rows.Next() {
				var record struct {
					id int
					DnsRecord
				}
				err = rows.Scan(&record.id, &record.Record, &record.Type, &record.Ip, &record.Datetime)
				if err != nil {
					continue
				}
				result = append(result, record)
			}
		case "http":
			for rows.Next() {
				var record struct {
					id int
					HttpRecord
				}
				err = rows.Scan(&record.id, &record.Method, &record.Url, &record.Head, &record.Body, &record.Ip, &record.Datetime)
				if err != nil {
					continue
				}
				result = append(result, record)
			}
		}
	} else {
		return result, 0, err
	}
	err = conn.QueryRow(fmt.Sprintf("select count(*) from %s %s", table, where)).Scan(&total)
	return result, total, nil
}

func Clean(table string) bool {
	if err := delete(table); err != nil {
		return true
	}
	return false
}

func Get(table string, page int, where string) ([]interface{}, int) {
	if result, total, err := query(table, where, page); err == nil {
		return result, total
	}
	return make([]interface{}, 0), 0
}

func (hr *HttpRecord) Insert() bool {
	sql := `INSERT INTO http (method, url, head, body, ip, datetime) values(?,?,?,?,?,?)`
	if stmt, err := conn.Prepare(sql); err == nil {
		if _, err = stmt.Exec(hr.Method, hr.Url, hr.Head, hr.Body, hr.Ip, hr.Datetime); err != nil {
			return false
		}
		return true
	}
	return false
}

func (dr *DnsRecord) Insert() bool {
	sql := `INSERT INTO dns (record, type, ip, datetime) values(?,?,?,?)`
	if stmt, err := conn.Prepare(sql); err == nil {
		if _, err = stmt.Exec(dr.Record, dr.Type, dr.Ip, dr.Datetime); err != nil {
			return false
		}
		return true
	}
	return false
}
