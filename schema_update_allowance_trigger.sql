-- Add allowance_trigger column to teacher_salaries
DO $$ 
BEGIN 
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'teacher_salaries' 
        AND column_name = 'allowance_trigger'
    ) THEN 
        ALTER TABLE teacher_salaries ADD COLUMN allowance_trigger VARCHAR(20) NOT NULL DEFAULT 'start_of_duty';
    END IF; 
END $$;
