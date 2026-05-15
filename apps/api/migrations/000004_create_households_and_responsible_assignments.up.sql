CREATE TABLE households (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mfy_id UUID NOT NULL REFERENCES mfys(id) ON DELETE CASCADE,
    street_id UUID NOT NULL REFERENCES streets(id) ON DELETE CASCADE,
    house_number TEXT NOT NULL,
    total_numbers INTEGER NOT NULL DEFAULT 0,
    contacted_numbers INTEGER NOT NULL DEFAULT 0,
    voted_numbers INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'NEW',
    notes TEXT NULL,
    assigned_responsible_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT households_house_number_not_empty CHECK (length(trim(house_number)) > 0),
    CONSTRAINT households_total_numbers_non_negative CHECK (total_numbers >= 0),
    CONSTRAINT households_contacted_numbers_non_negative CHECK (contacted_numbers >= 0),
    CONSTRAINT households_voted_numbers_non_negative CHECK (voted_numbers >= 0),
    CONSTRAINT households_contacted_numbers_lte_total CHECK (contacted_numbers <= total_numbers),
    CONSTRAINT households_voted_numbers_lte_total CHECK (voted_numbers <= total_numbers),
    CONSTRAINT households_status_check CHECK (
        status IN (
            'NEW',
            'VISITED',
            'EXPLAINED',
            'PARTIALLY_VOTED',
            'FULLY_VOTED',
            'NOT_HOME',
            'CALLBACK_NEEDED',
            'REFUSED',
            'INVALID_INFO'
        )
    )
);

CREATE UNIQUE INDEX households_street_house_number_unique
ON households (street_id, lower(house_number));

CREATE INDEX households_mfy_id_idx ON households (mfy_id);
CREATE INDEX households_street_id_idx ON households (street_id);
CREATE INDEX households_assigned_responsible_user_id_idx ON households (assigned_responsible_user_id);
CREATE INDEX households_status_idx ON households (status);
CREATE INDEX households_created_at_idx ON households (created_at);

CREATE TABLE responsible_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    street_id UUID NOT NULL REFERENCES streets(id) ON DELETE CASCADE,
    responsible_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    from_house_number TEXT NOT NULL,
    to_house_number TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT responsible_assignments_from_not_empty CHECK (length(trim(from_house_number)) > 0),
    CONSTRAINT responsible_assignments_to_not_empty CHECK (length(trim(to_house_number)) > 0)
);

CREATE INDEX responsible_assignments_street_id_idx ON responsible_assignments (street_id);
CREATE INDEX responsible_assignments_responsible_user_id_idx ON responsible_assignments (responsible_user_id);
CREATE INDEX responsible_assignments_is_active_idx ON responsible_assignments (is_active);
CREATE INDEX responsible_assignments_street_user_active_idx
ON responsible_assignments (street_id, responsible_user_id)
WHERE is_active = true;

CREATE TABLE household_change_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    household_id UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    changed_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    field_name TEXT NOT NULL,
    old_value TEXT NULL,
    new_value TEXT NULL,
    note TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX household_change_logs_household_id_idx ON household_change_logs (household_id);
CREATE INDEX household_change_logs_changed_by_user_id_idx ON household_change_logs (changed_by_user_id);
CREATE INDEX household_change_logs_created_at_idx ON household_change_logs (created_at);
