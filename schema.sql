-- Database Schema for Swadiq Schools Management System

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- User-Roles join table
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (user_id, role_id)
);

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Role-Permissions join table
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (role_id, permission_id)
);

-- Parents table
CREATE TABLE IF NOT EXISTS parents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(255),
    address TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Classes table
CREATE TABLE IF NOT EXISTS classes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    teacher_id INTEGER REFERENCES users(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Students table
CREATE TABLE IF NOT EXISTS students (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id VARCHAR(20) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'other')),
    address TEXT,
    class_id UUID REFERENCES classes(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Student-Parents join table
CREATE TABLE IF NOT EXISTS student_parents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    parent_id UUID NOT NULL REFERENCES parents(id) ON DELETE CASCADE,
    relationship VARCHAR(50) NOT NULL,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (student_id, parent_id)
);

-- Departments table
CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) UNIQUE NOT NULL,
    description TEXT,
    head_of_department_id INTEGER REFERENCES users(id),
    assistant_head_id INTEGER REFERENCES users(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Subjects table
CREATE TABLE IF NOT EXISTS subjects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) UNIQUE NOT NULL,
    department_id UUID REFERENCES departments(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Class-Subjects join table
CREATE TABLE IF NOT EXISTS class_subjects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (class_id, subject_id)
);

-- Class Promotions table
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

-- Attendance table
CREATE TABLE IF NOT EXISTS attendance (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    status VARCHAR(10) NOT NULL CHECK (status IN ('present', 'absent', 'late')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Exams table
CREATE TABLE IF NOT EXISTS exams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Papers table
CREATE TABLE IF NOT EXISTS papers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    teacher_id INTEGER REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Grades table
CREATE TABLE IF NOT EXISTS grades (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(10) NOT NULL,
    min_marks NUMERIC(5, 2) NOT NULL,
    max_marks NUMERIC(5, 2) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Results table
CREATE TABLE IF NOT EXISTS results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    exam_id UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    paper_id UUID NOT NULL REFERENCES papers(id) ON DELETE CASCADE,
    marks NUMERIC(5, 2) NOT NULL,
    grade_id UUID REFERENCES grades(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Fees table
CREATE TABLE IF NOT EXISTS fees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    amount NUMERIC(10, 2) NOT NULL,
    paid BOOLEAN DEFAULT false,
    due_date DATE NOT NULL,
    paid_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Payments table
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fee_id UUID NOT NULL REFERENCES fees(id) ON DELETE CASCADE,
    amount NUMERIC(10, 2) NOT NULL,
    paid_by INTEGER NOT NULL REFERENCES users(id),
    paid_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Schedules table
CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    teacher_id INTEGER NOT NULL REFERENCES users(id),
    day VARCHAR(10) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Expenses table
CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    amount NUMERIC(10, 2) NOT NULL,
    date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    recipient_id UUID NOT NULL,
    recipient_type VARCHAR(20) NOT NULL,
    email VARCHAR(255) NOT NULL,
    is_sent BOOLEAN DEFAULT false,
    sent_at TIMESTAMP WITH TIME ZONE,
    template VARCHAR(100),
    retry_count INTEGER DEFAULT 0,
    attachment_urls TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for all tables
DO $$
DECLARE
    t TEXT;
BEGIN
    FOR t IN 
        SELECT table_name FROM information_schema.columns 
        WHERE table_schema = 'public' AND column_name = 'updated_at'
    LOOP
        EXECUTE format('CREATE TRIGGER update_%I_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END;
$$;


-- Seed Data

-- Insert Roles
INSERT INTO roles (name) VALUES ('admin'), ('head_teacher'), ('class_teacher'), ('subject_teacher'), ('parent'), ('student') ON CONFLICT (name) DO NOTHING;

-- Insert Departments with heads and assistants
INSERT INTO departments (name, code, description, head_of_department_id, assistant_head_id) VALUES
    ('Mathematics', 'MATH', 'Mathematics and related subjects',
     (SELECT id FROM users WHERE email = 'john.mathematics@swadiqschools.com'),
     (SELECT id FROM users WHERE email = 'robert.physics@swadiqschools.com')),
    ('Science', 'SCI', 'Natural sciences including Physics, Chemistry, Biology',
     (SELECT id FROM users WHERE email = 'mary.science@swadiqschools.com'),
     (SELECT id FROM users WHERE email = 'jennifer.chemistry@swadiqschools.com')),
    ('Languages', 'LANG', 'English, Literature, and other languages',
     (SELECT id FROM users WHERE email = 'david.english@swadiqschools.com'),
     (SELECT id FROM users WHERE email = 'amanda.literature@swadiqschools.com')),
    ('Social Studies', 'SOC', 'History, Geography, Civics',
     (SELECT id FROM users WHERE email = 'sarah.history@swadiqschools.com'),
     (SELECT id FROM users WHERE email = 'christopher.geography@swadiqschools.com')),
    ('Arts', 'ARTS', 'Fine Arts, Music, Drama',
     (SELECT id FROM users WHERE email = 'michael.arts@swadiqschools.com'),
     (SELECT id FROM users WHERE email = 'jessica.music@swadiqschools.com')),
    ('Physical Education', 'PE', 'Physical Education and Sports',
     (SELECT id FROM users WHERE email = 'lisa.pe@swadiqschools.com'),
     NULL),
    ('Computer Science', 'CS', 'Computer Science and ICT',
     (SELECT id FROM users WHERE email = 'james.cs@swadiqschools.com'),
     NULL),
    ('Business Studies', 'BUS', 'Business, Economics, Entrepreneurship',
     (SELECT id FROM users WHERE email = 'emma.business@swadiqschools.com'),
     (SELECT id FROM users WHERE email = 'daniel.economics@swadiqschools.com'))
ON CONFLICT (code) DO NOTHING;

-- Insert Subjects
INSERT INTO subjects (name, code, department_id) VALUES
    ('Mathematics', 'MATH001', (SELECT id FROM departments WHERE code = 'MATH')),
    ('Physics', 'SCI001', (SELECT id FROM departments WHERE code = 'SCI')),
    ('Chemistry', 'SCI002', (SELECT id FROM departments WHERE code = 'SCI')),
    ('Biology', 'SCI003', (SELECT id FROM departments WHERE code = 'SCI')),
    ('English Language', 'LANG001', (SELECT id FROM departments WHERE code = 'LANG')),
    ('Literature', 'LANG002', (SELECT id FROM departments WHERE code = 'LANG')),
    ('History', 'SOC001', (SELECT id FROM departments WHERE code = 'SOC')),
    ('Geography', 'SOC002', (SELECT id FROM departments WHERE code = 'SOC')),
    ('Civics', 'SOC003', (SELECT id FROM departments WHERE code = 'SOC')),
    ('Fine Arts', 'ARTS001', (SELECT id FROM departments WHERE code = 'ARTS')),
    ('Music', 'ARTS002', (SELECT id FROM departments WHERE code = 'ARTS')),
    ('Physical Education', 'PE001', (SELECT id FROM departments WHERE code = 'PE')),
    ('Computer Science', 'CS001', (SELECT id FROM departments WHERE code = 'CS')),
    ('Business Studies', 'BUS001', (SELECT id FROM departments WHERE code = 'BUS')),
    ('Economics', 'BUS002', (SELECT id FROM departments WHERE code = 'BUS'))
ON CONFLICT (code) DO NOTHING;

-- Insert Permissions
INSERT INTO permissions (name) VALUES 
    ('users:create'), ('users:read'), ('users:update'), ('users:delete'),
    ('roles:assign'), ('students:manage'), ('fees:manage'), ('exams:manage')
ON CONFLICT (name) DO NOTHING;

-- Assign all permissions to admin role
WITH admin_role AS (SELECT id FROM roles WHERE name = 'admin')
INSERT INTO role_permissions (role_id, permission_id)
SELECT admin_role.id, p.id FROM admin_role, permissions p
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Insert default admin user
-- IMPORTANT: Replace the password hash with a real one for 'Ertdfgx @0'
INSERT INTO users (email, password, first_name, last_name)
VALUES ('imaad.ssebintu@gmail.com', '$2b$14$oeNl1VLiMNAy4mpwbJ4dTOiDzuEQnrjM3snmnTWKtNPFva873y296', 'imaad', 'ssebintu')
ON CONFLICT (email) DO UPDATE SET
    password = '$2b$14$oeNl1VLiMNAy4mpwbJ4dTOiDzuEQnrjM3snmnTWKtNPFva873y296',
    first_name = 'imaad',
    last_name = 'ssebintu';

-- Assign admin role to the new user
WITH admin_role AS (SELECT id FROM roles WHERE name = 'admin'),
     new_user AS (SELECT id FROM users WHERE email = 'imaad.ssebintu@gmail.com')
INSERT INTO user_roles (user_id, role_id)
SELECT new_user.id, admin_role.id FROM new_user, admin_role
ON CONFLICT (user_id, role_id) DO NOTHING;

-- Insert demo teachers
INSERT INTO users (email, password, first_name, last_name, is_active) VALUES
    ('john.mathematics@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'John', 'Smith', true),
    ('mary.science@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Mary', 'Johnson', true),
    ('david.english@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'David', 'Williams', true),
    ('sarah.history@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Sarah', 'Brown', true),
    ('michael.arts@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Michael', 'Davis', true),
    ('lisa.pe@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Lisa', 'Wilson', true),
    ('james.cs@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'James', 'Miller', true),
    ('emma.business@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Emma', 'Taylor', true),
    ('robert.physics@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Robert', 'Anderson', true),
    ('jennifer.chemistry@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Jennifer', 'Thomas', true),
    ('william.biology@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'William', 'Jackson', true),
    ('amanda.literature@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Amanda', 'White', true),
    ('christopher.geography@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Christopher', 'Harris', true),
    ('jessica.music@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Jessica', 'Martin', true),
    ('daniel.economics@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Daniel', 'Thompson', true),
    ('ashley.civics@swadiqschools.com', '$2a$14$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9S2', 'Ashley', 'Garcia', true)
ON CONFLICT (email) DO NOTHING;

-- Assign class_teacher role to demo teachers
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE u.email IN (
    'john.mathematics@swadiqschools.com', 'mary.science@swadiqschools.com', 'david.english@swadiqschools.com',
    'sarah.history@swadiqschools.com', 'michael.arts@swadiqschools.com', 'lisa.pe@swadiqschools.com',
    'james.cs@swadiqschools.com', 'emma.business@swadiqschools.com', 'robert.physics@swadiqschools.com',
    'jennifer.chemistry@swadiqschools.com', 'william.biology@swadiqschools.com', 'amanda.literature@swadiqschools.com',
    'christopher.geography@swadiqschools.com', 'jessica.music@swadiqschools.com', 'daniel.economics@swadiqschools.com',
    'ashley.civics@swadiqschools.com'
) AND r.name = 'class_teacher'
ON CONFLICT (user_id, role_id) DO NOTHING;