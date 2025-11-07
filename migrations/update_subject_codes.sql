-- Update subject codes to use 3-character format
UPDATE subjects SET code = 'ENG' WHERE code = 'ENG001';
UPDATE subjects SET code = 'MAT' WHERE code = 'MATH001';
UPDATE subjects SET code = 'SCI' WHERE code = 'SCI001';
UPDATE subjects SET code = 'SST' WHERE code = 'SST001';