-- Migration to add code column to classes table
-- Run this script to update existing database

-- Add code column to classes table
ALTER TABLE classes ADD COLUMN IF NOT EXISTS code VARCHAR(20);

-- Update existing classes with temporary codes using a different approach
DO $$
DECLARE
    class_record RECORD;
    counter INTEGER := 1;
BEGIN
    FOR class_record IN SELECT id FROM classes WHERE code IS NULL ORDER BY created_at LOOP
        UPDATE classes SET code = 'C-' || counter WHERE id = class_record.id;
        counter := counter + 1;
    END LOOP;
END $$;

-- Add unique constraint to code column
ALTER TABLE classes ADD CONSTRAINT classes_code_unique UNIQUE (code);

-- Make code column NOT NULL
ALTER TABLE classes ALTER COLUMN code SET NOT NULL;
