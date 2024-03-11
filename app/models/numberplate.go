package models

import (
	"database/sql"
)

// NumberPlate struct
type NumberPlate struct {
	ID        int    `json:"id"`
	Plate     string `json:"plate" validate:"required"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt" db:"created_at"`
	UpdatedAt string `json:"updatedAt" db:"updated_at"`
}

// Add number plate
func (e *NumberPlate) Add(env *Env) (int64, error) {
	// Add to database
	res, err := env.DB.Exec(
		"INSERT INTO number_plates (plate, name, created_at, updated_at) VALUES (?, ?, DATE(), DATE())",
		&e.Plate, &e.Name,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Get number plate by ID provided
func (e *NumberPlate) Get(env *Env) (*NumberPlate, error) {
	// Get from database
	var numberPlates []NumberPlate
	err := env.DB.Query(&numberPlates, "SELECT * FROM number_plates WHERE id = ? LIMIT 1", &e.ID)
	if err != nil {
		return nil, err
	}
	return &numberPlates[0], nil
}

// Find number plates by fields provided
func (e *NumberPlate) Find(env *Env, operator string, fields []WhereFields, perPage int, pageNumber int) (*[]NumberPlate, int, error) {
	resCount := 0
	// Where
	whereSQL, values := env.DB.WhereSQL(operator, fields)
	// Limit
	limitSQL := env.DB.LimitSQL(perPage, pageNumber)
	if limitSQL != "" {
		// Get count from database
		var numberPlates []NumberPlate
		err := env.DB.Query(&numberPlates, "SELECT id FROM number_plates"+whereSQL, values...)
		if err != nil {
			return nil, 0, err
		}
		resCount = len(numberPlates)
	}
	// Get from database
	var numberPlates []NumberPlate
	err := env.DB.Query(&numberPlates, "SELECT * FROM number_plates"+whereSQL+" ORDER BY name, plate"+limitSQL, values...)
	if err != nil {
		return nil, 0, err
	}
	if resCount == 0 {
		resCount = len(numberPlates)
	}
	return &numberPlates, resCount, nil
}

// Update number plate
func (e *NumberPlate) Update(env *Env) (int64, error) {
	// Update database
	res, err := env.DB.Exec(
		"UPDATE number_plates SET plate = ?, name = ?, updated_at = DATE() WHERE id = ?",
		&e.Plate, &e.Name, &e.ID,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Delete number plate
func (e *NumberPlate) Delete(env *Env) (int64, error) {
	// Delete from database
	res, err := env.DB.Exec("DELETE FROM number_plates WHERE id = ?", &e.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Migrate number plates
func (e *NumberPlate) Migrate(env *Env) (sql.Result, error) {
	// Create table if not exists
	res, err := env.DB.Exec(`
	CREATE TABLE IF NOT EXISTS number_plates (
		id INTEGER NOT NULL PRIMARY KEY,
		plate TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT 0
	);
	`)
	if err != nil {
		return nil, err
	}
	return res, nil
}
