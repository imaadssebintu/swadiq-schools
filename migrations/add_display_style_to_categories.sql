-- Add display_style column to assessment_categories table
ALTER TABLE assessment_categories ADD COLUMN IF NOT EXISTS display_style VARCHAR(20) DEFAULT 'table';

-- Update existing categories with default value
UPDATE assessment_categories SET display_style = 'table' WHERE display_style IS NULL;
