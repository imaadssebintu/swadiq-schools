-- Add term_id column to assessment_types table
ALTER TABLE assessment_types ADD COLUMN term_id UUID;

-- Create index for faster lookups by term_id
CREATE INDEX idx_assessment_types_term_id ON assessment_types(term_id);

-- Optional: Add Foreign Key constraint (if terms table exists and you want strict referential integrity)
-- ALTER TABLE assessment_types ADD CONSTRAINT fk_assessment_types_term_id FOREIGN KEY (term_id) REFERENCES terms(id);
