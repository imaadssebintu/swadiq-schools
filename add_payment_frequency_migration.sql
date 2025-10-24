-- Migration to add payment_frequency column to fee_types table

-- Add payment_frequency column
ALTER TABLE fee_types ADD COLUMN IF NOT EXISTS payment_frequency VARCHAR(20);

-- Update existing records with default value
UPDATE fee_types SET payment_frequency = 'once' WHERE payment_frequency IS NULL;

-- Make the column NOT NULL
ALTER TABLE fee_types ALTER COLUMN payment_frequency SET NOT NULL;

-- Add check constraint for valid payment frequencies
ALTER TABLE fee_types ADD CONSTRAINT fee_types_payment_frequency_check 
CHECK (payment_frequency IN ('once', 'per_term', 'per_year', 'on_demand'));
