package models

import (
	"database/sql"
)

// AdminLog struct
type AdminLog struct {
	ID      int    `json:"id"`
	Type    string `json:"type" validate:"required"`
	Details string `json:"details" validate:"required"`
	UserID  int    `json:"userID" validate:"required" db:"user_id"`
	Time    string `json:"time" validate:"required"`
}

// Add admin log entry
func (l *AdminLog) Add(env *Env) (int64, error) {
	// Add to database
	res, err := env.DB.Exec(
		"INSERT INTO admin_logs (type, details, user_id, time) VALUES (?, ?, ?, DATE())",
		&l.Type, &l.Details, &l.UserID,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Get admin log entries by ID provided
func (l *AdminLog) Get(env *Env) (*AdminLog, error) {
	// Get from database
	var adminLogs []AdminLog
	err := env.DB.Query(&adminLogs, "SELECT * FROM admin_logs WHERE id = ? LIMIT 1", &l.ID)
	if err != nil {
		return nil, err
	}
	return &adminLogs[0], nil
}

// Find admin log entries by fields provided
func (l *AdminLog) Find(env *Env, operator string, fields []WhereFields, perPage int, pageNumber int) (*[]AdminLog, int, error) {
	resCount := 0
	// Where
	whereSQL, values := env.DB.WhereSQL(operator, fields)
	// Limit
	limitSQL := env.DB.LimitSQL(perPage, pageNumber)
	if limitSQL != "" {
		// Get count from database
		var adminLogs []AdminLog
		err := env.DB.Query(&adminLogs, "SELECT id FROM admin_logs"+whereSQL, values...)
		if err != nil {
			return nil, 0, err
		}
		resCount = len(adminLogs)
	}
	// Get from database
	var adminLogs []AdminLog
	err := env.DB.Query(&adminLogs, "SELECT * FROM admin_logs"+whereSQL+" ORDER BY time DESC, type ASC, user_id ASC"+limitSQL, values...)
	if err != nil {
		return nil, 0, err
	}
	if resCount == 0 {
		resCount = len(adminLogs)
	}
	return &adminLogs, resCount, nil
}

// Migrate admin log
func (l *AdminLog) Migrate(env *Env) (sql.Result, error) {
	// Create table if not exists
	res, err := env.DB.Exec(`
	CREATE TABLE IF NOT EXISTS admin_logs (
		id INTEGER NOT NULL PRIMARY KEY,
		type TEXT NOT NULL UNIQUE,
		details text NOT NULL,
		user_id INTEGER NOT NULL DEFAULT 0,
		time INTEGER NOT NULL DEFAULT 0
	);
	`)
	if err != nil {
		return nil, err
	}
	return res, nil
}
