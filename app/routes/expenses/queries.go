package expenses

import (
	"database/sql"
	"fmt"
	"log"
	"swadiq-schools/app/models"
)

// InitExpensesDB ensures necessary tables and columns exist.
func InitExpensesDB(db *sql.DB) error {
	// 1. Create tables if they don't exist
	queries := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) UNIQUE NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,
		`CREATE TABLE IF NOT EXISTS expenses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			amount BIGINT NOT NULL,
			currency VARCHAR(3) DEFAULT 'UGX' NOT NULL,
			date DATE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Printf("Error creating expenses tables: %v", err)
			return err
		}
	}

	// 2. Ensure columns exist (Migrations for existing tables)
	migrations := []string{
		`ALTER TABLE expenses ADD COLUMN IF NOT EXISTS currency VARCHAR(3) DEFAULT 'UGX' NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_category_id ON expenses(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(date)`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_deleted_at ON expenses(deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON categories(deleted_at)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			log.Printf("Error running expenses migration: %v", err)
			// Continue as some might be duplicate index errors depending on PG version
		}
	}

	// 3. Seed default data
	seeds := []string{
		`INSERT INTO categories (name, is_active) VALUES ('Salary', true) ON CONFLICT (name) DO NOTHING`,
	}

	for _, s := range seeds {
		if _, err := db.Exec(s); err != nil {
			log.Printf("Error seeding expenses data: %v", err)
		}
	}

	return nil
}

// Expense Queries
func GetAllExpenses(db *sql.DB) ([]*models.Expense, error) {
	query := `SELECT e.id, e.category_id, e.title, e.amount, e.currency, e.date, 
			  e.created_at, e.updated_at, c.id, c.name
			  FROM expenses e
			  LEFT JOIN categories c ON e.category_id = c.id
			  WHERE e.deleted_at IS NULL
			  ORDER BY e.date DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := []*models.Expense{} // Initialize to empty slice for non-null JSON
	for rows.Next() {
		e := &models.Expense{}
		var catID, catName sql.NullString
		err := rows.Scan(
			&e.ID, &e.CategoryID, &e.Title, &e.Amount, &e.Currency, &e.Date,
			&e.CreatedAt, &e.UpdatedAt, &catID, &catName,
		)
		if err != nil {
			return nil, err
		}

		if catID.Valid {
			e.Category = &models.Category{
				ID:   catID.String,
				Name: catName.String,
			}
		}
		expenses = append(expenses, e)
	}
	return expenses, nil
}

func GetExpenseByID(db *sql.DB, id string) (*models.Expense, error) {
	query := `SELECT e.id, e.category_id, e.title, e.amount, e.currency, e.date, 
			  e.created_at, e.updated_at, c.id, c.name
			  FROM expenses e
			  LEFT JOIN categories c ON e.category_id = c.id
			  WHERE e.id = $1 AND e.deleted_at IS NULL`

	e := &models.Expense{}
	var catID, catName sql.NullString
	err := db.QueryRow(query, id).Scan(
		&e.ID, &e.CategoryID, &e.Title, &e.Amount, &e.Currency, &e.Date,
		&e.CreatedAt, &e.UpdatedAt, &catID, &catName,
	)
	if err != nil {
		return nil, err
	}

	if catID.Valid {
		e.Category = &models.Category{
			ID:   catID.String,
			Name: catName.String,
		}
	}
	return e, nil
}

func CreateExpense(db *sql.DB, e *models.Expense) error {
	query := `INSERT INTO expenses (category_id, title, amount, currency, date, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	return db.QueryRow(query, e.CategoryID, e.Title, e.Amount, e.Currency, e.Date).
		Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

func UpdateExpense(db *sql.DB, e *models.Expense) error {
	query := `UPDATE expenses 
			  SET category_id = $1, title = $2, amount = $3, currency = $4, date = $5, updated_at = NOW()
			  WHERE id = $6 AND deleted_at IS NULL`

	result, err := db.Exec(query, e.CategoryID, e.Title, e.Amount, e.Currency, e.Date, e.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("expense not found")
	}
	return nil
}

func DeleteExpense(db *sql.DB, id string) error {
	query := `UPDATE expenses SET deleted_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// Category Queries
func GetAllCategories(db *sql.DB) ([]*models.Category, error) {
	query := `SELECT id, name, is_active, created_at, updated_at
			  FROM categories
			  WHERE deleted_at IS NULL
			  ORDER BY name ASC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []*models.Category{} // Initialize to empty slice for non-null JSON
	for rows.Next() {
		c := &models.Category{}
		err := rows.Scan(&c.ID, &c.Name, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func CreateCategory(db *sql.DB, c *models.Category) error {
	query := `INSERT INTO categories (name, is_active, created_at, updated_at)
			  VALUES ($1, $2, NOW(), NOW())
			  RETURNING id, created_at, updated_at`

	return db.QueryRow(query, c.Name, c.IsActive).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func UpdateCategory(db *sql.DB, c *models.Category) error {
	query := `UPDATE categories 
			  SET name = $1, is_active = $2, updated_at = NOW()
			  WHERE id = $3 AND deleted_at IS NULL`

	result, err := db.Exec(query, c.Name, c.IsActive, c.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

func DeleteCategory(db *sql.DB, id string) error {
	query := `UPDATE categories SET deleted_at = NOW() WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}
