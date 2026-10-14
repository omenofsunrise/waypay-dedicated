package persons

import "waypay_dedicated/sdk"

func migrations() []sdk.Migration {
	return []sdk.Migration{
		{
			Version: 1,
			Name:    "create_persons",
			UpSQL: `
				CREATE EXTENSION IF NOT EXISTS pgcrypto;

				CREATE TABLE IF NOT EXISTS persons (
					id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					full_name            TEXT NOT NULL,
					phone_number         TEXT,
					email                TEXT,
					birth_date           DATE,
					registration_address TEXT,
					fact_address         TEXT,
					is_deleted           BOOLEAN NOT NULL DEFAULT FALSE,
					created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
					updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
				);

				CREATE UNIQUE INDEX IF NOT EXISTS ux_persons_phone
					ON persons (phone_number)
					WHERE phone_number IS NOT NULL AND is_deleted = FALSE;

				CREATE UNIQUE INDEX IF NOT EXISTS ux_persons_email
					ON persons (email)
					WHERE email IS NOT NULL AND is_deleted = FALSE;

				CREATE INDEX IF NOT EXISTS ix_persons_full_name
					ON persons (full_name) WHERE is_deleted = FALSE;

				CREATE OR REPLACE FUNCTION persons_set_updated_at()
				RETURNS TRIGGER AS $$
				BEGIN
					NEW.updated_at := now();
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;

				DROP TRIGGER IF EXISTS trg_persons_updated_at ON persons;
				CREATE TRIGGER trg_persons_updated_at
					BEFORE UPDATE ON persons
					FOR EACH ROW
					EXECUTE FUNCTION persons_set_updated_at();
				`,
			DownSQL: `
				DROP TRIGGER IF EXISTS trg_persons_updated_at ON persons;
				DROP FUNCTION IF EXISTS persons_set_updated_at();
				DROP TABLE IF EXISTS persons;
				`,
		},
	}
}
