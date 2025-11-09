-- Sample data for fees system

-- Insert sample fee types
INSERT INTO fee_types (name, code, description, is_active) VALUES
('Tuition Fee', 'TUITION', 'Regular tuition fees for academic year', true),
('Lunch Fee', 'LUNCH', 'Daily lunch fees', true),
('Transport Fee', 'TRANSPORT', 'School transport fees', true),
('Activity Fee', 'ACTIVITY', 'Extra-curricular activity fees', true),
('Exam Fee', 'EXAM', 'Examination fees', true)
ON CONFLICT (code) DO NOTHING;

-- Get some sample student IDs and fee type IDs for creating fees
-- Note: This assumes you have students in your database
-- You may need to adjust the student IDs based on your actual data

-- Insert sample fees (you'll need to replace the UUIDs with actual student IDs from your database)
-- This is just a template - you'll need to run a query to get actual student IDs first