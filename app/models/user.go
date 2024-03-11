package models

import (
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

// User struct
type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email" validate:"omitempty,email"`
	Password     string `json:"-" validate:"omitempty,passwd"`
	PasswordHash string `json:"-" db:"password_hash"`
	CreatedAt    string `json:"createdAt" db:"created_at"`
	UpdatedAt    string `json:"updatedAt" db:"updated_at"`
}

// Add user
func (s *User) Add(env *Env) (int64, error) {
	// Check if email already exists
	var user User
	_, resCount, err := user.Find(env, "AND", []WhereFields{{"email", "=", s.Email}}, 0, 1)
	if err != nil && !errors.Is(err, env.DB.ErrRecordNotFound) {
		return 0, err
	}
	if resCount > 0 {
		return 0, errors.New("email address already exists")
	}
	// Hash password
	if s.Password != "" {
		err := s.hashPassword()
		if err != nil {
			return 0, err
		}
	}
	// Add to database
	res, err := env.DB.Exec(
		"INSERT INTO users (email, password_hash, created_at, updated_at) VALUES (?, ?, DATE(), DATE())",
		&s.Email, &s.PasswordHash,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Get user by ID provided
func (s *User) Get(env *Env) (*User, error) {
	// Get from database
	var user []User
	err := env.DB.Query(&user, "SELECT * FROM users WHERE id = ? LIMIT 1", &s.ID)
	if err != nil {
		return nil, err
	}
	return &user[0], nil
}

// Find user by fields provided
func (s *User) Find(env *Env, operator string, fields []WhereFields, perPage int, pageNumber int) (*[]User, int, error) {
	resCount := 0
	// Where
	whereSQL, values := env.DB.WhereSQL(operator, fields)
	// Limit
	limitSQL := env.DB.LimitSQL(perPage, pageNumber)
	if limitSQL != "" {
		// Get count from database
		var users []User
		err := env.DB.Query(&users, "SELECT id FROM users"+whereSQL, values...)
		if err != nil {
			return nil, 0, err
		}
		resCount = len(users)
	}
	// Get from database
	var users []User
	err := env.DB.Query(&users, "SELECT * FROM users"+whereSQL+" ORDER BY email"+limitSQL, values...)
	if err != nil {
		return nil, 0, err
	}
	if resCount == 0 {
		resCount = len(users)
	}
	return &users, resCount, nil
}

// Update user
func (s *User) Update(env *Env) (int64, error) {
	// Check that ID exists
	user := User{ID: s.ID}
	res, err := user.Get(env)
	if err != nil && !errors.Is(err, env.DB.ErrRecordNotFound) {
		return 0, err
	}
	if res == nil {
		return 0, errors.New("user not found")
	}
	// Check if another user already exists with the email address
	if s.Email != "" {
		res2, resCount, err := user.Find(env, "AND", []WhereFields{{"email", "=", s.Email}}, 0, 1)
		if err != nil && !errors.Is(err, env.DB.ErrRecordNotFound) {
			return 0, err
		}
		if resCount > 0 {
			for _, r2 := range *res2 {
				if r2.ID != s.ID {
					return 0, errors.New("email address already exists")
				}
			}
		}
	}
	// Hash password
	if s.Password != "" {
		err = s.hashPassword()
		if err != nil {
			return 0, err
		}
	}
	// Update database
	res3, err := env.DB.Exec(
		"UPDATE users SET email = ?, password_hash = ?, updated_at = DATE() WHERE id = ?",
		&s.Email, &s.PasswordHash, &s.ID,
	)
	if err != nil {
		return 0, err
	}
	return res3.RowsAffected()
}

// Delete user
func (s *User) Delete(env *Env) (int64, error) {
	// Delete from database
	res, err := env.DB.Exec("DELETE FROM users WHERE id = ?", &s.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// hashPassword generates a hash of a user's password
func (s *User) hashPassword() error {
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(s.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// Update user's hash
	s.PasswordHash = string(hash)

	return nil
}

// CheckLogin checks if an email and password combination are valid
func (s *User) CheckLogin(env *Env) (*User, bool, error) {
	// Check user exists
	var user User
	res, resCount, err := user.Find(env, "AND", []WhereFields{{"email", "=", s.Email}}, 0, 1)
	if err != nil && !errors.Is(err, env.DB.ErrRecordNotFound) {
		return nil, false, err
	}
	if resCount > 0 {
		for _, r := range *res {
			// Compare the password with the password hash if not blank
			if r.PasswordHash != "" {
				if err := bcrypt.CompareHashAndPassword([]byte(r.PasswordHash), []byte(s.Password)); err != nil {
					if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
						return nil, false, err
					} else {
						return nil, false, nil
					}
				}
				return &r, true, nil
			}
		}
	}
	return nil, false, errors.New("user not found")
}

// Migrate user
func (s *User) Migrate(env *Env) (sql.Result, error) {
	// Create table if not exists
	res, err := env.DB.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER NOT NULL PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password_hash text NOT NULL,
		created_at TEXT NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT 0
	);
	`)
	if err != nil {
		return nil, err
	}
	return res, nil
}
