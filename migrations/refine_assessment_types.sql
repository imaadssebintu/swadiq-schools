-- Refine Assessment Types Table
ALTER TABLE assessment_types ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES assessment_types(id);
ALTER TABLE assessment_types ADD COLUMN IF NOT EXISTS category VARCHAR(50);
ALTER TABLE assessment_types ADD COLUMN IF NOT EXISTS weight NUMERIC(5,2) DEFAULT 1.0;

-- Update Exams Table to reference Assessment Types
ALTER TABLE exams ADD COLUMN IF NOT EXISTS assessment_type_id UUID REFERENCES assessment_types(id);

-- Seed Standard Assessment Types
-- 1. Main Term Exams
INSERT INTO assessment_types (name, code, category, color, is_active)
VALUES 
('Beginning of Term', 'BOT', 'Exam', 'indigo', true),
('Mid Term', 'MT', 'Exam', 'purple', true),
('End of Term', 'EOT', 'Exam', 'pink', true)
ON CONFLICT (code) DO UPDATE SET category = 'Exam';

-- 2. Categories for sub-types
INSERT INTO assessment_types (name, code, category, color, is_active)
VALUES 
('Class Test', 'CTEST', 'Test', 'amber', true),
('Course Project', 'CPROJ', 'Project', 'emerald', true)
ON CONFLICT (code) DO UPDATE SET category = EXCLUDED.category;

-- 3. Sub-types (Sample)
-- We need IDs to link them, using a sub-query for clarity in this script
DO $$ 
DECLARE 
    ctest_id UUID;
    cproj_id UUID;
BEGIN
    SELECT id INTO ctest_id FROM assessment_types WHERE code = 'CTEST';
    SELECT id INTO cproj_id FROM assessment_types WHERE code = 'CPROJ';

    -- Sub-types for Class Test
    INSERT INTO assessment_types (name, code, parent_id, category, color, is_active)
    VALUES 
    ('Test 1', 'T1', ctest_id, 'Test', 'amber', true),
    ('Test 2', 'T2', ctest_id, 'Test', 'amber', true)
    ON CONFLICT (code) DO NOTHING;

    -- Sub-types for Course Project
    INSERT INTO assessment_types (name, code, parent_id, category, color, is_active)
    VALUES 
    ('Project 1', 'P1', cproj_id, 'Project', 'emerald', true),
    ('Lab Report', 'LR1', cproj_id, 'Project', 'emerald', true)
    ON CONFLICT (code) DO NOTHING;
END $$;
