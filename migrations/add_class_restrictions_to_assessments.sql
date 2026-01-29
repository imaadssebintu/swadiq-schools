-- Add all_classes column to assessment_types
ALTER TABLE assessment_types ADD COLUMN IF NOT EXISTS all_classes BOOLEAN DEFAULT TRUE;

-- Create junction table for assessment types and classes
CREATE TABLE IF NOT EXISTS assessment_type_classes (
    assessment_type_id UUID REFERENCES assessment_types(id) ON DELETE CASCADE,
    class_id UUID REFERENCES classes(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (assessment_type_id, class_id)
);

-- Index for performance
CREATE INDEX IF NOT EXISTS idx_assessment_type_classes_assessment_type_id ON assessment_type_classes(assessment_type_id);
CREATE INDEX IF NOT EXISTS idx_assessment_type_classes_class_id ON assessment_type_classes(class_id);
