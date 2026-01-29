-- Update the exam to add term_id and academic_year_id
UPDATE exams 
SET 
    term_id = '33130100-6151-457f-8c25-5a8aeec79df2',  -- Term 1 (2026-2027)
    academic_year_id = '2e79f9b2-c603-4184-8611-d50a614f6689',  -- 2026-2027 (current)
    updated_at = NOW()
WHERE id = '1448ea1e-283b-4a60-aa87-03927444815c';  -- English Paper 1

-- Verify the update
SELECT id, name, term_id, academic_year_id 
FROM exams 
WHERE id = '1448ea1e-283b-4a60-aa87-03927444815c';
