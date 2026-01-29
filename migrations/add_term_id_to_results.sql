-- Add term_id column to results table
ALTER TABLE results ADD COLUMN IF NOT EXISTS term_id UUID;

-- Add foreign key constraint
ALTER TABLE results ADD CONSTRAINT fk_results_term 
    FOREIGN KEY (term_id) REFERENCES terms(id);

-- Add index for better query performance
CREATE INDEX IF NOT EXISTS idx_results_term_id ON results(term_id);
