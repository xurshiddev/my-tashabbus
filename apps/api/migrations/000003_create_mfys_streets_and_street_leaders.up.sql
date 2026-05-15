CREATE TABLE mfys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    region TEXT NULL,
    district TEXT NULL,
    target_votes INTEGER NULL,
    season TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT mfys_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT mfys_target_votes_non_negative CHECK (target_votes IS NULL OR target_votes >= 0)
);

CREATE INDEX mfys_name_idx ON mfys (name);
CREATE INDEX mfys_region_district_idx ON mfys (region, district);
CREATE INDEX mfys_is_active_idx ON mfys (is_active);

ALTER TABLE users
    ADD CONSTRAINT users_mfy_id_fkey
    FOREIGN KEY (mfy_id) REFERENCES mfys(id) ON DELETE SET NULL;

CREATE TABLE streets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mfy_id UUID NOT NULL REFERENCES mfys(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    planned_households_count INTEGER NOT NULL DEFAULT 0,
    notes TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT streets_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT streets_planned_households_count_non_negative CHECK (planned_households_count >= 0)
);

CREATE UNIQUE INDEX streets_mfy_lower_name_active_unique
ON streets (mfy_id, lower(name))
WHERE is_active = true;

CREATE INDEX streets_mfy_id_idx ON streets (mfy_id);
CREATE INDEX streets_is_active_idx ON streets (is_active);
CREATE INDEX streets_name_idx ON streets (name);

CREATE TABLE street_leader_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    street_id UUID NOT NULL REFERENCES streets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX street_leader_assignments_one_active_per_street
ON street_leader_assignments (street_id)
WHERE is_active = true;

CREATE INDEX street_leader_assignments_street_id_idx ON street_leader_assignments (street_id);
CREATE INDEX street_leader_assignments_user_id_idx ON street_leader_assignments (user_id);
CREATE INDEX street_leader_assignments_is_active_idx ON street_leader_assignments (is_active);
