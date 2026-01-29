-- Rename team_id to term_id
ALTER TABLE assessment_types RENAME COLUMN team_id TO term_id;

-- Rename the index
ALTER INDEX idx_assessment_types_team_id RENAME TO idx_assessment_types_term_id;
