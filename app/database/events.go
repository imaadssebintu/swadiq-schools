package database

import (
	"database/sql"
	"swadiq-schools/app/models"
)

// CreateEvent adds a new event to the database
func CreateEvent(db *sql.DB, event *models.Event) error {
	query := `
		INSERT INTO events (title, description, start_date, end_date, type, location, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return db.QueryRow(
		query,
		event.Title,
		event.Description,
		event.StartDate,
		event.EndDate,
		event.Type,
		event.Location,
		event.Color,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
}

// GetEvents retrieves all events from the database ordered by start_date
func GetEvents(db *sql.DB) ([]models.Event, error) {
	query := `
		SELECT id, title, description, start_date, end_date, type, location, color, created_at, updated_at
		FROM events
		ORDER BY start_date ASC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.StartDate, &e.EndDate,
			&e.Type, &e.Location, &e.Color, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

// DeleteEvent deletes an event by ID
func DeleteEvent(db *sql.DB, id int) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// GetEventCategoryCounts returns the count of events for each category
func GetEventCategoryCounts(db *sql.DB) (map[string]int, error) {
	query := `SELECT type, COUNT(*) FROM events GROUP BY type`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, err
		}
		counts[category] = count
	}

	// Ensure defaults
	for _, cat := range []string{"academic", "holiday", "sports", "cultural"} {
		if _, ok := counts[cat]; !ok {
			counts[cat] = 0
		}
	}

	return counts, nil
}

// UpdateEvent updates an existing event
func UpdateEvent(db *sql.DB, event *models.Event) error {
	query := `
		UPDATE events
		SET title = $1, description = $2, start_date = $3, end_date = $4,
			type = $5, location = $6, color = $7, updated_at = NOW()
		WHERE id = $8
	`
	_, err := db.Exec(query,
		event.Title, event.Description, event.StartDate, event.EndDate,
		event.Type, event.Location, event.Color, event.ID,
	)
	return err
}
