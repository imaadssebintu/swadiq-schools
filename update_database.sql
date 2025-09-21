-- Update script for class promotions feature
-- Run this in your PostgreSQL database

-- Step 1: Create academic_years table first (required for foreign key)
CREATE TABLE IF NOT EXISTS academic_years (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Step 2: Insert sample academic year if none exists
INSERT INTO academic_years (name, start_date, end_date, is_active)
SELECT '2024-2025', '2024-09-01', '2025-06-30', true
WHERE NOT EXISTS (SELECT 1 FROM academic_years WHERE name = '2024-2025');

-- Step 3: Now create class_promotions table (with foreign key to academic_years)
CREATE TABLE IF NOT EXISTS class_promotions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    to_class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    academic_year_id UUID REFERENCES academic_years(id),
    promotion_criteria TEXT, -- JSON string with promotion criteria
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (from_class_id, academic_year_id)
);

-- Verify tables were created
SELECT 'academic_years table created' as status
WHERE EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'academic_years');

SELECT 'class_promotions table created' as status
WHERE EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'class_promotions');
