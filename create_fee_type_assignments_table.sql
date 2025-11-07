-- Create fee_type_assignments table
CREATE TABLE fee_type_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fee_type_id UUID NOT NULL REFERENCES fee_types(id) ON DELETE CASCADE,
    student_id UUID REFERENCES students(id) ON DELETE CASCADE,
    class_id UUID REFERENCES classes(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    -- Ensure either student_id or class_id is provided, but not both
    CONSTRAINT chk_assignment_target CHECK (
        (student_id IS NOT NULL AND class_id IS NULL) OR 
        (student_id IS NULL AND class_id IS NOT NULL)
    )
);

-- Create indexes
CREATE INDEX idx_fee_type_assignments_fee_type_id ON fee_type_assignments(fee_type_id);
CREATE INDEX idx_fee_type_assignments_student_id ON fee_type_assignments(student_id);
CREATE INDEX idx_fee_type_assignments_class_id ON fee_type_assignments(class_id);
CREATE INDEX idx_fee_type_assignments_deleted_at ON fee_type_assignments(deleted_at);

-- Create unique constraint to prevent duplicate assignments
CREATE UNIQUE INDEX idx_fee_type_assignments_unique_student 
ON fee_type_assignments(fee_type_id, student_id) 
WHERE student_id IS NOT NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX idx_fee_type_assignments_unique_class 
ON fee_type_assignments(fee_type_id, class_id) 
WHERE class_id IS NOT NULL AND deleted_at IS NULL;