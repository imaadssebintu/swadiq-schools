-- Update existing paper codes to use first 3 characters format
UPDATE papers SET code = 'ENG-1' WHERE code = 'ENG001-1';
UPDATE papers SET code = 'MAT-1' WHERE code = 'MATH001-1';
UPDATE papers SET code = 'SCI-1' WHERE code = 'SCI001-1';
UPDATE papers SET code = 'SST-1' WHERE code = 'SST001-1';