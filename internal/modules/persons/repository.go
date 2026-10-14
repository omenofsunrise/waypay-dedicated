package persons

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo { return &Repo{db: db} }

func scanPerson(s interface{ Scan(dest ...any) error }) (*Person, error) {
	var (
		p        Person
		ptype    string
		phone    sql.NullString
		email    sql.NullString
		birth    sql.NullString
		regAddr  sql.NullString
		factAddr sql.NullString
	)
	err := s.Scan(
		&p.ID,
		&ptype,
		&p.FullName,
		&phone,
		&email,
		&birth,
		&regAddr,
		&factAddr,
		&p.IsDeleted,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if phone.Valid {
		p.PhoneNumber = &phone.String
	}
	if email.Valid {
		p.Email = &email.String
	}
	if birth.Valid {
		p.BirthDate = &birth.String
	}
	if regAddr.Valid {
		p.RegistrationAddress = &regAddr.String
	}
	if factAddr.Valid {
		p.FactAddress = &factAddr.String
	}
	return &p, nil
}

const selectColumns = `
	id::text,
	full_name,
	phone_number,
	email,
	to_char(birth_date, 'YYYY-MM-DD'),
	registration_address,
	fact_address,
	is_deleted,
	created_at,
	updated_at`

// ---- Create --------------------------------------------------------------

func (r *Repo) Create(ctx context.Context, in CreateInput) (*Person, error) {
	in.Normalize()
	if err := in.Validate(); err != nil {
		return nil, err
	}

	const q = `
		INSERT INTO persons
			(full_name, phone_number, email, birth_date,
			 registration_address, fact_address)
		VALUES ($1, $2, $3, $4, $5::date, $6, $7)
		RETURNING ` + selectColumns

	row := r.db.QueryRowContext(ctx, q,
		in.FullName,
		in.PhoneNumber,
		in.Email,
		in.BirthDate,
		in.RegistrationAddress,
		in.FactAddress,
	)
	p, err := scanPerson(row)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("person with same phone/email already exists")
		}
		return nil, err
	}
	return p, nil
}

// ---- Read ----------------------------------------------------------------

func (r *Repo) GetByID(ctx context.Context, id string) (*Person, error) {
	const q = `SELECT ` + selectColumns + `
		FROM persons WHERE id = $1::uuid AND is_deleted = FALSE`

	p, err := scanPerson(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *Repo) List(ctx context.Context, f ListFilter) ([]Person, error) {
	where := []string{"is_deleted = FALSE"}
	args := []any{}
	i := 1

	if f.Search != "" {
		where = append(where,
			fmt.Sprintf(`(full_name ILIKE $%d OR phone_number ILIKE $%d OR email ILIKE $%d)`, i, i, i))
		args = append(args, "%"+f.Search+"%")
		i++
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	q := fmt.Sprintf(
		`SELECT %s FROM persons WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		selectColumns, strings.Join(where, " AND "), i, i+1,
	)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Person
	for rows.Next() {
		p, err := scanPerson(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// ---- Update --------------------------------------------------------------

func (r *Repo) Update(ctx context.Context, id string, in UpdateInput) (*Person, error) {
	in.Normalize()
	if err := in.Validate(); err != nil {
		return nil, err
	}

	sets := []string{}
	args := []any{}
	i := 1

	add := func(expr string, v any) {
		sets = append(sets, fmt.Sprintf(expr, i))
		args = append(args, v)
		i++
	}

	if in.FullName != nil {
		add("full_name = $%d", *in.FullName)
	}
	if in.PhoneNumber != nil {
		add("phone_number = $%d", nullableStr(*in.PhoneNumber))
	}
	if in.Email != nil {
		add("email = $%d", nullableStr(*in.Email))
	}
	if in.BirthDate != nil {
		add("birth_date = $%d::date", nullableStr(*in.BirthDate))
	}
	if in.RegistrationAddress != nil {
		add("registration_address = $%d", nullableStr(*in.RegistrationAddress))
	}
	if in.FactAddress != nil {
		add("fact_address = $%d", nullableStr(*in.FactAddress))
	}

	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}

	args = append(args, id)
	q := fmt.Sprintf(
		`UPDATE persons SET %s WHERE id = $%d::uuid AND is_deleted = FALSE RETURNING %s`,
		strings.Join(sets, ", "), i, selectColumns,
	)

	p, err := scanPerson(r.db.QueryRowContext(ctx, q, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil && isUniqueViolation(err) {
		return nil, fmt.Errorf("person with same phone/email already exists")
	}
	return p, err
}

// ---- Delete (soft) -------------------------------------------------------

func (r *Repo) SoftDelete(ctx context.Context, id string) error {
	const q = `UPDATE persons SET is_deleted = TRUE
		WHERE id = $1::uuid AND is_deleted = FALSE`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func findBy(ctx context.Context, tx *sql.Tx, cond string, args ...any) (*Person, error) {
	q := `SELECT ` + selectColumns + ` FROM persons
		WHERE ` + cond + ` AND is_deleted = FALSE
		ORDER BY created_at LIMIT 1`
	p, err := scanPerson(tx.QueryRowContext(ctx, q, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func commitAndReturn(tx *sql.Tx, p *Person, created bool) (*Person, bool, error) {
	if err := tx.Commit(); err != nil {
		return nil, false, err
	}
	return p, created, nil
}

// ---- helpers -------------------------------------------------------------

func nullableStr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func deref(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}
