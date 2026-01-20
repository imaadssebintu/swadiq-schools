-- Migration script to transition events and event_categories to UUIDv4
-- WARNING: This script will delete all existing data in these tables.

-- 1. Drop existing tables (order matters for Foreign Keys)
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS event_categories;

-- 2. Create event_categories table with UUID
CREATE TABLE event_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    color VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create events table with UUID
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    type VARCHAR(50),
    category_id UUID REFERENCES event_categories(id),
    location VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Insert default categories
INSERT INTO event_categories (name, color)
VALUES 
    ('Academic', '#ef4444'),
    ('Holiday', '#3b82f6'),
    ('Sports', '#22c55e'),
    ('Cultural', '#a855f7');
