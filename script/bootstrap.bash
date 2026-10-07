#!/usr/bin/env bash

set -e

if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

DB_NAME=${DB_NAME:-autocall}
DB_USER=${DB_USER:-postgres}

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
        'Jane Doe',
        'Acme Inc',
        '7017495876',
        'pending',
        NULL,
        false
    ),
    (
        'Arjun Sharma',
        'Flipkart',
        '7017495876',
        'pending',
        NULL,
        false
    ),
    (
        'Olivia Brown',
        'DesignLoop Agency',
        '7017495876',
        'interested',
        NULL,
        false
    ),
    (
        'Benjamin Lee',
        'TechInnovate',
        '7017495876',
        'completed',
        NULL,
        false
    ),
    (
        'Rohan Verma',
        'Razorpay',
        '7017495876',
        'failed',
        'no answer',
        false
    ),
    (
        'Sophia Johnson',
        'Predictable Hiring',
        '7017495876',
        'interested',
        NULL,
        true
    )
ON CONFLICT (phone_number) DO NOTHING;

SQL

echo "Bootstrap complete."
