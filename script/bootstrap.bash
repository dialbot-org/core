#!/usr/bin/env bash

set -e

if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

DB_NAME=${DB_NAME:-autocall}
DB_USER=${DB_USER:-visshon}

echo "Seeding database '${DB_NAME}' as user '${DB_USER}'..."

psql -U "$DB_USER" -d "$DB_NAME" <<'SQL'

INSERT INTO leads (
    full_name,
    company,
    phone_number,
    status,
    error,
    google_sheet_synced
)
VALUES
    (
        'Nakul Thareja',
        'Arista',
        '9540202263',
        'pending',
        NULL,
        false
    ),
    (
        'Vishnu Shon',
        'Beenverified',
        '+917017495876',
        'pending',
        NULL,
        false
    ),
    (
        'Akshat Saini',
        'Arista',
        '+91 72908 64480',
        'interested',
        NULL,
        false
    ),
    (
        'Divyam Gupta',
        'Mathworks',
        '+91-9872439611',
        'pending',
        NULL,
        false
    ),
    (
        'Yashwardhan Chaudhuri',
        'Orbitron AI',
        '9315525988',
        'pending',
        NULL,
        false
    ),
    (
        'Shreyanshu Sharma',
        'National Australian Bank',
        '74045-69304',
        'pending',
        NULL,
        false
    )
ON CONFLICT (phone_number) DO NOTHING;

SQL

echo "Bootstrap complete."
