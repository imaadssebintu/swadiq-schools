-- Create teacher_payments table
CREATE TABLE IF NOT EXISTS teacher_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES users(id),
    amount BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'base_salary', 'allowance', 'combined'
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reference VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for payment queries
CREATE INDEX IF NOT EXISTS idx_teacher_payments_teacher_date ON teacher_payments(teacher_id, paid_at);
