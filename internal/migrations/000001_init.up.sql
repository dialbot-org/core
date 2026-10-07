CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE leads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    full_name TEXT NOT NULL,
    company TEXT NOT NULL,
    phone_number TEXT NOT NULL,

    status TEXT NOT NULL DEFAULT 'pending',
    error TEXT,

    google_sheet_synced BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_leads_phone_number ON leads(phone_number);
CREATE INDEX idx_leads_status ON leads(status);
