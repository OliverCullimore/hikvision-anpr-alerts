package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/faabiosr/cachego"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/url"
	"strings"
	"time"
)

type DB struct {
	Conn              *sqlx.DB
	cache             cachego.Cache
	logger            *log.Logger
	ErrRecordNotFound error
}

type WhereFields struct {
	Field              string
	ComparisonOperator string
	Value              any
}

type OrderFields struct {
	Field     string
	Value     any
	Extra     string
	SubFields []OrderFields
}

// Init sets up the connection pool global variable.
func (d *DB) Init(config Config, cache cachego.Cache, logger *log.Logger) error {
	// Form database connection string
	dsn := fmt.Sprintf("file:%s?&_loc=%s", config.DBFile, url.QueryEscape("Europe/London"))
	// Open database connection
	db, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	// Verify connection to the database was successful
	err = db.Ping()
	if err != nil {
		return err
	}
	d.Conn = db
	d.cache = cache
	d.logger = logger
	d.ErrRecordNotFound = errors.New("record not found")
	return nil
}

func (d *DB) WhereSQL(operator string, fields []WhereFields) (string, []interface{}) {
	whereSQL := ""
	operator = fmt.Sprintf(" %s ", strings.ToUpper(strings.TrimSpace(operator)))
	operatorBefore := strings.TrimSpace(operator)
	var values []interface{}
	var where []string
	for _, field := range fields {
		// fmt.Printf("Field: %s Operator: %s Comparison Operator: %s Value: %s\n", field.Field, operatorBefore, field.ComparisonOperator, strings.TrimSpace(fmt.Sprint(field.Value)))
		values = append(values, field.Value)
		where = append(where, fmt.Sprintf("%s %s ?", field.Field, field.ComparisonOperator))
	}
	if len(where) > 0 {
		if strings.Contains(operator, " AND ") || strings.Contains(operator, " OR ") {
			operatorBefore = ""
		}
		whereSQL = " WHERE " + operatorBefore + strings.Join(where, operator)
	}
	return whereSQL, values
}
func (d *DB) OrderSQL(fields []OrderFields, values []interface{}) (string, []interface{}) {
	orderSQL := ""
	var order []string
	for _, field := range fields {
		fieldSQL := fmt.Sprintf("%s", field.Field)
		if field.Value != "" {
			values = append(values, field.Value)
			fieldSQL += " ?"
		}
		var subOrder []string
		for _, subField := range field.SubFields {
			subFieldSQL := fmt.Sprintf("%s", subField.Field)
			if subField.Value != "" {
				values = append(values, subField.Value)
				subFieldSQL += "?"
			}
			if subField.Extra != "" {
				subFieldSQL += fmt.Sprintf(" %s", subField.Extra)
			}
			if subFieldSQL != "" {
				subOrder = append(subOrder, subFieldSQL)
			}
		}
		if len(subOrder) > 0 {
			fieldSQL += " " + strings.Join(subOrder, " ")
		}
		if field.Extra != "" {
			fieldSQL += fmt.Sprintf(" %s", field.Extra)
		}
		if fieldSQL != "" {
			order = append(order, fieldSQL)
		}
	}
	if len(order) > 0 {
		orderSQL = " ORDER BY " + strings.Join(order, ", ")
	}
	return orderSQL, values
}

func (d *DB) LimitSQL(perPage int, pageNumber int) string {
	limitSQL := ""
	if perPage > 0 {
		if pageNumber < 1 {
			pageNumber = 1
		}
		limitSQL = fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, (pageNumber-1)*perPage)
	}
	return limitSQL
}

func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	d.logger.Println(query)
	return d.Conn.Exec(query, args...)
}

func (d *DB) Query(resStruct interface{}, query string, args ...interface{}) error {
	// TODO: Implement caching
	// Check cache
	/*if err := d.cacheFetch(query, resStruct); err != nil {
	    d.logger.Println(err)
	}*/
	d.logger.Println(query)
	// Expand slice values in args
	query, args, err := sqlx.In(query, args...)
	// Rebind the query
	query = d.Conn.Rebind(query)
	// Execute the query
	err = d.Conn.Select(resStruct, query, args...)
	if err != nil {
		return err
	}
	// Save to cache
	/*if err := d.cacheSave(query, resStruct, 10*time.Second); err != nil {
	    d.logger.Println(err)
	}*/
	return nil
}

func (d *DB) cacheFetch(query string, resStruct interface{}) error {
	// Fetch from cache
	if cacheRes, err := d.cache.Fetch("db_" + query); err != nil {
		return err
	} else if err := json.Unmarshal([]byte(cacheRes), &resStruct); err != nil {
		return err
	}
	return nil
}

func (d *DB) cacheSave(query string, resStruct interface{}, ttl time.Duration) error {
	// Convert to JSON and save to cache
	if str, err := json.Marshal(resStruct); err != nil {
		return err
	} else if err := d.cache.Save("db_"+query, string(str), ttl); err != nil {
		return err
	}
	return nil
}

// Migrate performs database migrations.
func (d *DB) Migrate(env *Env) error {
	// Migrate
	numberPlate := NumberPlate{}
	if _, err := numberPlate.Migrate(env); err != nil {
		return err
	}
	camera := Camera{}
	if _, err := camera.Migrate(env); err != nil {
		return err
	}
	user := User{}
	if _, err := user.Migrate(env); err != nil {
		return err
	}
	adminLog := AdminLog{}
	if _, err := adminLog.Migrate(env); err != nil {
		return err
	}

	// Add default user if none exist
	_, resCount, err := user.Find(env, "AND", []WhereFields{}, 0, 1)
	if err != nil {
		return err
	}
	if resCount == 0 {
		env.Logger.Println("Creating default user")
		user2 := User{Email: "changeme@example.com", Password: "changeme"}
		_, err = user2.Add(env)
		if err != nil {
			return err
		}
	}

	return nil
}
