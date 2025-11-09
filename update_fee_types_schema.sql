-- Add amount field to fee_types table
ALTER TABLE fee_types ADD COLUMN IF NOT EXISTS amount INTEGER DEFAULT 0;
ALTER TABLE fee_types ADD COLUMN IF NOT EXISTS payment_frequency VARCHAR(20) DEFAULT 'per_term';
ALTER TABLE fee_types ADD COLUMN IF NOT EXISTS scope VARCHAR(50) DEFAULT 'manual';

-- Update existing records to have default payment frequency
UPDATE fee_types SET payment_frequency = 'per_term' WHERE payment_frequency IS NULL OR payment_frequency = '';

-- Create fee type assignments table
CREATE TABLE IF NOT EXISTS fee_type_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fee_type_id UUID NOT NULL REFERENCES fee_types(id) ON DELETE CASCADE,
    class_id UUID REFERENCES classes(id) ON DELETE CASCADE,
    student_id UUID REFERENCES students(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fee_type_assignments_check CHECK (
        (class_id IS NOT NULL AND student_id IS NULL) OR 
        (class_id IS NULL AND student_id IS NOT NULL)
    )
);