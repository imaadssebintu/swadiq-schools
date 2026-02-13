package database

import (
	"database/sql"
	"swadiq-schools/app/models"
)

// GetActiveClassesSimple retrieves a simple list of active classes (ID, Name, Code)
func GetActiveClassesSimple(db *sql.DB) ([]models.Class, error) {
	query := `SELECT id, name, code FROM classes WHERE is_active = true ORDER BY name ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []models.Class
	for rows.Next() {
		var c models.Class
		if err := rows.Scan(&c.ID, &c.Name, &c.Code); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}
