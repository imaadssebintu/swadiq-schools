package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"swadiq-schools/app/models"
)

// InitEventDatabase ensures the event_categories table exists and has default values,
// and that the events table has a category_id column.
func InitEventDatabase(db *sql.DB) error {
	// Create event_categories table
	categoriesTableQuery := `
		CREATE TABLE IF NOT EXISTS event_categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) UNIQUE NOT NULL,
			color VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.Exec(categoriesTableQuery); err != nil {
		return fmt.Errorf("failed to create event_categories table: %v", err)
	}

	// Insert default categories if none exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM event_categories").Scan(&count)
	if err == nil && count == 0 {
		insertQuery := `
			INSERT INTO event_categories (name, color)
			VALUES 
				('Academic', '#ef4444'),
				('Holiday', '#3b82f6'),
				('Sports', '#22c55e'),
				('Cultural', '#a855f7')
		`
		if _, err := db.Exec(insertQuery); err != nil {
			return fmt.Errorf("failed to insert default categories: %v", err)
		}
	}

	// Add category_id to events table if it doesn't exist
	addColumnQuery := `
		DO $$ 
		BEGIN 
			-- Ensure category_id use UUID
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='events' AND column_name='category_id') THEN
				ALTER TABLE events ADD COLUMN category_id UUID REFERENCES event_categories(id);
			ELSIF (SELECT data_type FROM information_schema.columns WHERE table_name='events' AND column_name='category_id') != 'uuid' THEN
				ALTER TABLE events ALTER COLUMN category_id TYPE UUID USING category_id::TEXT::UUID;
			END IF;

			-- Also ensure events.id is UUID
			IF (SELECT data_type FROM information_schema.columns WHERE table_name='events' AND column_name='id') != 'uuid' THEN
				-- This is a destructive migration if there's existing data, but user provided schema shows it's still integer.
				-- We'll try to convert it.
				ALTER TABLE events ALTER COLUMN id TYPE UUID USING gen_random_uuid(), ALTER COLUMN id SET DEFAULT gen_random_uuid();
			END IF;

			-- Remove color column from events table as it's now category-based
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='events' AND column_name='color') THEN
				ALTER TABLE events DROP COLUMN color;
			END IF;

			-- Add term_id to events table if it doesn't exist
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='events' AND column_name='term_id') THEN
				ALTER TABLE events ADD COLUMN term_id UUID REFERENCES terms(id);
			END IF;

			-- Add suspends_classes to events table if it doesn't exist
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='events' AND column_name='suspends_classes') THEN
				ALTER TABLE events ADD COLUMN suspends_classes BOOLEAN DEFAULT FALSE;
			END IF;

			-- Add suspension_type to events table if it doesn't exist
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='events' AND column_name='suspension_type') THEN
				ALTER TABLE events ADD COLUMN suspension_type VARCHAR(20) DEFAULT 'NONE';
			END IF;

			-- Add suspended_class_ids to events table if it doesn't exist
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='events' AND column_name='suspended_class_ids') THEN
				ALTER TABLE events ADD COLUMN suspended_class_ids JSONB DEFAULT '[]'::jsonb;
			END IF;
		END $$;
	`
	if _, err := db.Exec(addColumnQuery); err != nil {
		return fmt.Errorf("failed to migrate events table: %v", err)
	}

	return nil
}

// GetEventCategories retrieves all event categories
func GetEventCategories(db *sql.DB) ([]models.EventCategory, error) {
	query := `SELECT id, name, color, created_at, updated_at FROM event_categories ORDER BY name ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.EventCategory
	for rows.Next() {
		var c models.EventCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Color, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// CreateEventCategory adds a new event category to the database
func CreateEventCategory(db *sql.DB, category *models.EventCategory) error {
	query := `
		INSERT INTO event_categories (name, color, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return db.QueryRow(query, category.Name, category.Color).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

// UpdateEventCategory updates an existing event category
func UpdateEventCategory(db *sql.DB, category *models.EventCategory) error {
	query := `
		UPDATE event_categories 
		SET name = $1, color = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := db.Exec(query, category.Name, category.Color, category.ID)
	return err
}

// DeleteEventCategory deletes an event category
func DeleteEventCategory(db *sql.DB, id string) error {
	// First, set all events with this category to null or a default?
	// For now, let's just delete the category and let the foreign key handle it (if any)
	// Or explicitly nullify them
	nullifyQuery := `UPDATE events SET category_id = NULL WHERE category_id = $1`
	db.Exec(nullifyQuery, id)

	query := `DELETE FROM event_categories WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// CreateEvent adds a new event to the database
func CreateEvent(db *sql.DB, event *models.Event) error {
	suspendedClassIDsJSON, err := json.Marshal(event.SuspendedClassIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal suspended class ids: %w", err)
	}

	query := `
		INSERT INTO events (title, description, start_date, end_date, type, category_id, term_id, location, suspension_type, suspended_class_ids, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::UUID, NULLIF($7, '')::UUID, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return db.QueryRow(
		query,
		event.Title,
		event.Description,
		event.StartDate,
		event.EndDate,
		event.Type,
		event.CategoryID,
		event.TermID,
		event.Location,
		event.SuspensionType,
		suspendedClassIDsJSON,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
}

// GetEvents retrieves events from the database ordered by start_date.
// If upcomingOnly is true, it only returns events where end_date is now or in the future.
func GetEvents(db *sql.DB, upcomingOnly bool) ([]models.Event, error) {
	whereClause := ""
	if upcomingOnly {
		whereClause = "WHERE e.end_date >= NOW()"
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.title, e.description, e.start_date, e.end_date, e.type, e.category_id, 
		       COALESCE(c.name, e.type) as category_name, e.term_id, COALESCE(t.name, '') as term_name, e.location, 
		       COALESCE(c.color, '#0f172a') as color, e.suspension_type, e.suspended_class_ids, e.created_at, e.updated_at
		FROM events e
		LEFT JOIN event_categories c ON e.category_id = c.id
		LEFT JOIN terms t ON e.term_id = t.id
		%s
		ORDER BY e.start_date ASC
	`, whereClause)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		var suspendedClassIDsJSON []byte
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.StartDate, &e.EndDate,
			&e.Type, &e.CategoryID, &e.CategoryName, &e.TermID, &e.TermName, &e.Location, &e.Color, &e.SuspensionType, &suspendedClassIDsJSON, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if len(suspendedClassIDsJSON) > 0 {
			if err := json.Unmarshal(suspendedClassIDsJSON, &e.SuspendedClassIDs); err != nil {
				e.SuspendedClassIDs = []string{}
			}
		} else {
			e.SuspendedClassIDs = []string{}
		}
		events = append(events, e)
	}
	return events, nil
}

// DeleteEvent deletes an event by ID
func DeleteEvent(db *sql.DB, id string) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}

// GetEventCategoryCounts returns the count of events for each category.
// If upcomingOnly is true, it only counts upcoming events.
func GetEventCategoryCounts(db *sql.DB, upcomingOnly bool) (map[string]int, error) {
	whereClause := ""
	if upcomingOnly {
		whereClause = "WHERE e.end_date >= NOW()"
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(c.name, e.type), COUNT(*) 
		FROM events e
		LEFT JOIN event_categories c ON e.category_id = c.id
		%s
		GROUP BY COALESCE(c.name, e.type)
	`, whereClause)
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
	suspendedClassIDsJSON, err := json.Marshal(event.SuspendedClassIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal suspended class ids: %w", err)
	}

	query := `
		UPDATE events
		SET title = $1, description = $2, start_date = $3, end_date = $4,
			type = $5, category_id = NULLIF($6, '')::UUID, term_id = NULLIF($7, '')::UUID, location = $8, 
			suspension_type = $9, suspended_class_ids = $10, updated_at = NOW()
		WHERE id = $11
	`
	_, err = db.Exec(query,
		event.Title, event.Description, event.StartDate, event.EndDate,
		event.Type, event.CategoryID, event.TermID, event.Location, event.SuspensionType, suspendedClassIDsJSON, event.ID,
	)
	return err
}
