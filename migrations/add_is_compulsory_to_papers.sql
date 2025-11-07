-- Add is_compulsory column to papers table
ALTER TABLE papers ADD COLUMN is_compulsory BOOLEAN DEFAULT true;

-- Update existing papers to be compulsory by default
UPDATE papers SET is_compulsory = true WHERE is_compulsory IS NULL;