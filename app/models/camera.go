package models

import (
	"database/sql"
)

// Camera struct
type Camera struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IPAddress string `json:"ipAddress" validate:"required" db:"ip_address"`
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	CreatedAt string `json:"createdAt" db:"created_at"`
	UpdatedAt string `json:"updatedAt" db:"updated_at"`
}

// Add number plate
func (e *Camera) Add(env *Env) (int64, error) {
	// Add to database
	res, err := env.DB.Exec(
		"INSERT INTO cameras (name, ip_address, username, password, created_at, updated_at) VALUES (?, ?, ?, ?, DATE(), DATE())",
		&e.Name, &e.IPAddress, &e.Username, &e.Password,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Get number plate by ID provided
func (e *Camera) Get(env *Env) (*Camera, error) {
	// Get from database
	var cameras []Camera
	err := env.DB.Query(&cameras, "SELECT * FROM cameras WHERE id = ? LIMIT 1", &e.ID)
	if err != nil {
		return nil, err
	}
	return &cameras[0], nil
}

// Find number plates by fields provided
func (e *Camera) Find(env *Env, operator string, fields []WhereFields, perPage int, pageNumber int) (*[]Camera, int, error) {
	resCount := 0
	// Where
	whereSQL, values := env.DB.WhereSQL(operator, fields)
	// Limit
	limitSQL := env.DB.LimitSQL(perPage, pageNumber)
	if limitSQL != "" {
		// Get count from database
		var cameras []Camera
		err := env.DB.Query(&cameras, "SELECT id FROM cameras"+whereSQL, values...)
		if err != nil {
			return nil, 0, err
		}
		resCount = len(cameras)
	}
	// Get from database
	var cameras []Camera
	err := env.DB.Query(&cameras, "SELECT * FROM cameras"+whereSQL+" ORDER BY name, ip_address"+limitSQL, values...)
	if err != nil {
		return nil, 0, err
	}
	if resCount == 0 {
		resCount = len(cameras)
	}
	return &cameras, resCount, nil
}

// Update number plate
func (e *Camera) Update(env *Env) (int64, error) {
	// Update database
	res, err := env.DB.Exec(
		"UPDATE cameras SET name = ?, ip_address = ?, username = ?, password = ?, updated_at = DATE() WHERE id = ?",
		&e.Name, &e.IPAddress, &e.Username, &e.Password, &e.ID,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Delete number plate
func (e *Camera) Delete(env *Env) (int64, error) {
	// Delete from database
	res, err := env.DB.Exec("DELETE FROM cameras WHERE id = ?", &e.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Migrate number plates
func (e *Camera) Migrate(env *Env) (sql.Result, error) {
	// Create table if not exists
	res, err := env.DB.Exec(`
	CREATE TABLE IF NOT EXISTS cameras (
		id INTEGER NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		ip_address TEXT NOT NULL UNIQUE,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT 0
	);
	`)
	if err != nil {
		return nil, err
	}
	return res, nil
}
