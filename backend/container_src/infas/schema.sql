CREATE TABLE IF NOT EXISTS doctors (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    email          TEXT NOT NULL UNIQUE,
    password_hash  TEXT NOT NULL,
    full_name      TEXT NOT NULL,
    specialty      TEXT,
    license_number TEXT UNIQUE,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS patients (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name     TEXT NOT NULL,
    test_id       INTEGER,
    phone_number  TEXT NOT NULL UNIQUE,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    password_hash TEXT NOT NULL DEFAULT '',
    email         TEXT,
    status        TEXT NOT NULL DEFAULT 'Waiting for graded',
    doctor_id     INTEGER
);

CREATE TABLE IF NOT EXISTS tests (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    patient_id INTEGER NOT NULL,
    score      INTEGER NOT NULL CHECK (score BETWEEN 1 AND 30),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_patients_doctor_id ON patients(doctor_id);
CREATE INDEX IF NOT EXISTS idx_tests_patient_id ON tests(patient_id);

CREATE TABLE IF NOT EXISTS password_reset_token (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    doctor_id   INTEGER NOT NULL,
    email       TEXT NOT NULL,
    token_hash  BLOB NOT NULL UNIQUE,
    expires_at  INTEGER NOT NULL,
    used_at     INTEGER,
    role        TEXT NOT NULL DEFAULT 'doctor',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_password_reset_token_email ON password_reset_token(email);
