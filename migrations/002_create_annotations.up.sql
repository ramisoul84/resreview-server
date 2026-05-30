CREATE TABLE IF NOT EXISTS annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version_id UUID NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    data JSONB NOT NULL DEFAULT '{}',
    user_id UUID NOT NULL,
    session_id TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '',
    stroke_w REAL NOT NULL DEFAULT 2,
    stroke_style TEXT NOT NULL DEFAULT 'solid',
    x REAL NOT NULL DEFAULT 0,
    y REAL NOT NULL DEFAULT 0,
    title TEXT NOT NULL DEFAULT '',
    text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_annotations_version_id ON annotations(version_id);
