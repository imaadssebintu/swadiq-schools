-- Add is_required column to fee_types table
ALTER TABLE fee_types ADD COLUMN is_required BOOLEAN DEFAULT true NOT NULL;

-- Create index on is_required column
CREATE INDEX idx_fee_types_is_required ON fee_types(is_required);

-- Set specific fee types as required (must pay)
UPDATE fee_types SET is_required = true WHERE LOWER(name) LIKE '%school%' OR LOWER(name) LIKE '%tuition%' OR LOWER(name) LIKE '%registration%';

-- Set specific fee types as optional
UPDATE fee_types SET is_required = false WHERE LOWER(name) LIKE '%trip%' OR LOWER(name) LIKE '%excursion%' OR LOWER(name) LIKE '%optional%';