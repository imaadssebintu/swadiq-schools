-- Timetable Settings Table
CREATE TABLE IF NOT EXISTS timetable_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    class_id VARCHAR(255), -- Empty string for default settings
    days JSONB NOT NULL DEFAULT '["monday","tuesday","wednesday","thursday","friday"]',
    start_time TIME NOT NULL DEFAULT '08:00',
    end_time TIME NOT NULL DEFAULT '16:00',
    lesson_duration INTEGER NOT NULL DEFAULT 60, -- in minutes
    breaks JSONB NOT NULL DEFAULT '[]',
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Timetable Entries Table
CREATE TABLE IF NOT EXISTS timetable_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    class_id VARCHAR(255) NOT NULL,
    subject_id VARCHAR(255) NOT NULL,
    teacher_id VARCHAR(255),
    day VARCHAR(20) NOT NULL,
    time_slot VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES subjects(id) ON DELETE CASCADE,
    FOREIGN KEY (teacher_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Insert default timetable settings
INSERT INTO timetable_settings (class_id, days, start_time, end_time, lesson_duration, breaks, is_default)
VALUES (
    '',
    '["monday","tuesday","wednesday","thursday","friday"]',
    '08:00',
    '16:00',
    60,
    '[{"name":"Breakfast Break","start_time":"10:00","end_time":"10:30"},{"name":"Lunch Break","start_time":"12:30","end_time":"13:30"}]',
    true
) ON CONFLICT DO NOTHING;

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_timetable_settings_class_id ON timetable_settings(class_id);
CREATE INDEX IF NOT EXISTS idx_timetable_settings_is_default ON timetable_settings(is_default);
CREATE INDEX IF NOT EXISTS idx_timetable_entries_class_id ON timetable_entries(class_id);
CREATE INDEX IF NOT EXISTS idx_timetable_entries_day ON timetable_entries(day);