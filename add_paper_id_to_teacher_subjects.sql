-- Add paper_id column to teacher_subjects table
ALTER TABLE teacher_subjects ADD COLUMN paper_id uuid REFERENCES papers(id);

-- Create index for better performance
CREATE INDEX idx_teacher_subjects_paper_id ON teacher_subjects(paper_id);
CREATE INDEX idx_teacher_subjects_teacher_subject_paper ON teacher_subjects(teacher_id, subject_id, paper_id);