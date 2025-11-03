-- Quick fix: Update API to include class_code in response
-- The issue is that class_code is not being populated in the API response
-- This is a temporary workaround until the Go code is fixed

-- Check current class data
SELECT id, name, code FROM classes WHERE is_active = true ORDER BY name;