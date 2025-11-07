-- Add is_compulsory column to class_subjects table
ALTER TABLE class_subjects ADD COLUMN is_compulsory BOOLEAN DEFAULT true;

-- Update existing class_subjects to be compulsory by default
UPDATE class_subjects SET is_compulsory = true WHERE is_compulsory IS NULL;