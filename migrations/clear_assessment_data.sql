-- Clear Assessment Data and Results
-- Order matters to satisfy foreign key constraints

-- 1. Clear Results (references Exams)
TRUNCATE TABLE results CASCADE;

-- 2. Clear Exams (references Assessment Types)
TRUNCATE TABLE exams CASCADE;

-- 3. Clear Assessment Type Classes (references Assessment Types)
TRUNCATE TABLE assessment_type_classes CASCADE;

-- 4. Clear Assessment Types
TRUNCATE TABLE assessment_types CASCADE;

-- Note: We keep assessment_categories as the user only asked to clear types and results.
