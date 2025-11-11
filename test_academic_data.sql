-- Insert sample academic year
INSERT INTO academic_years (name, start_date, end_date, is_current, is_active, created_at, updated_at)
VALUES ('2024-2025', '2024-01-01', '2024-12-31', true, true, NOW(), NOW());

-- Insert sample terms
INSERT INTO terms (academic_year_id, name, start_date, end_date, is_current, is_active, created_at, updated_at)
SELECT id, 'Term 1', '2024-01-01', '2024-04-30', true, true, NOW(), NOW()
FROM academic_years WHERE name = '2024-2025';

INSERT INTO terms (academic_year_id, name, start_date, end_date, is_current, is_active, created_at, updated_at)
SELECT id, 'Term 2', '2024-05-01', '2024-08-31', false, true, NOW(), NOW()
FROM academic_years WHERE name = '2024-2025';

INSERT INTO terms (academic_year_id, name, start_date, end_date, is_current, is_active, created_at, updated_at)
SELECT id, 'Term 3', '2024-09-01', '2024-12-31', false, true, NOW(), NOW()
FROM academic_years WHERE name = '2024-2025';