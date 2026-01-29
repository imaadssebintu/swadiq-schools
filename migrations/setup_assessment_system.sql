-- Create assessment_categories table
CREATE TABLE IF NOT EXISTS assessment_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE,
    color VARCHAR(20) DEFAULT 'indigo',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Update assessment_types table
ALTER TABLE assessment_types ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES assessment_categories(id);
ALTER TABLE assessment_types ADD COLUMN IF NOT EXISTS weight NUMERIC(5,2) DEFAULT 1.0;
ALTER TABLE assessment_types ADD COLUMN IF NOT EXISTS color VARCHAR(20) DEFAULT 'indigo';
ALTER TABLE assessment_types ADD COLUMN IF NOT EXISTS all_classes BOOLEAN DEFAULT TRUE;

-- Create junction table for class-specific assessments
CREATE TABLE IF NOT EXISTS assessment_type_classes (
    assessment_type_id UUID REFERENCES assessment_types(id) ON DELETE CASCADE,
    class_id UUID REFERENCES classes(id) ON DELETE CASCADE,
    PRIMARY KEY (assessment_type_id, class_id)
);

-- Seed default categories if empty
INSERT INTO assessment_categories (name, code, color)
SELECT 'Exam', 'EXM', 'indigo' WHERE NOT EXISTS (SELECT 1 FROM assessment_categories WHERE code = 'EXM');
INSERT INTO assessment_categories (name, code, color)
SELECT 'Test', 'TST', 'purple' WHERE NOT EXISTS (SELECT 1 FROM assessment_categories WHERE code = 'TST');
INSERT INTO assessment_categories (name, code, color)
SELECT 'Project', 'PRJ', 'emerald' WHERE NOT EXISTS (SELECT 1 FROM assessment_categories WHERE code = 'PRJ');
