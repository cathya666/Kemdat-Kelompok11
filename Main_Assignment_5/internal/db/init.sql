DROP TABLE IF EXISTS consents CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS grades CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS users CASCADE;

CREATE TABLE users (
    id SERIAL PRIMARY KEY, 
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL, 
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'judge', 'student')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP 
);

CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT, 
    student_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE grades (
    id SERIAL PRIMARY KEY,
    project_id INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE, 
    judge_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, 
    score INT CHECK(score >= 0 AND score <= 100), 
    feedback TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, judge_id) 
);

CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id INT REFERENCES users(id) ON DELETE SET NULL, 
    username VARCHAR(50),
    method VARCHAR(100),
    path TEXT,
    status_code INT,
    ip_address VARCHAR(50),
    user_agent TEXT
);
