package bank

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrCompanyNotFound = errors.New("company not found")
var ErrCompanyRegisteredIDExists = errors.New("company with registered id already exists")
var ErrCompanyOwnerNotFound = errors.New("company owner not found")
var ErrCompanyActivityCodeNotFound = errors.New("company activity code not found")

func scanCompany(scanner interface {
	Scan(dest ...any) error
}) (*Company, error) {
	var company Company
	var activityCodeID sql.NullInt64
	err := scanner.Scan(
		&company.Id,
		&company.Registered_id,
		&company.Name,
		&company.Tax_code,
		&activityCodeID,
		&company.Address,
		&company.Owner_id,
	)
	if err != nil {
		return nil, err
	}
	if activityCodeID.Valid {
		company.Activity_code_id = activityCodeID.Int64
	}
	return &company, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (s *Server) CreateCompanyRecord(company Company) (*Company, error) {
	tx, err := s.database.Begin()
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var ownerExists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1)`, company.Owner_id).Scan(&ownerExists); err != nil {
		return nil, fmt.Errorf("checking owner existence: %w", err)
	}
	if !ownerExists {
		return nil, ErrCompanyOwnerNotFound
	}

	if company.Activity_code_id != 0 {
		var activityCodeExists bool
		if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM activity_codes WHERE id = $1)`, company.Activity_code_id).Scan(&activityCodeExists); err != nil {
			return nil, fmt.Errorf("checking activity code existence: %w", err)
		}
		if !activityCodeExists {
			return nil, ErrCompanyActivityCodeNotFound
		}
	}

	var row *sql.Row
	if company.Activity_code_id == 0 {
		row = tx.QueryRow(`
			INSERT INTO companies (registered_id, name, tax_code, activity_code_id, address, owner_id)
			VALUES ($1, $2, $3, NULL, $4, $5)
			RETURNING id, registered_id, name, tax_code, activity_code_id, address, owner_id
		`, company.Registered_id, company.Name, company.Tax_code, company.Address, company.Owner_id)
	} else {
		row = tx.QueryRow(`
			INSERT INTO companies (registered_id, name, tax_code, activity_code_id, address, owner_id)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, registered_id, name, tax_code, activity_code_id, address, owner_id
		`, company.Registered_id, company.Name, company.Tax_code, company.Activity_code_id, company.Address, company.Owner_id)
	}

	created, err := scanCompany(row)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCompanyRegisteredIDExists
		}
		return nil, fmt.Errorf("creating company: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return created, nil
}

func (s *Server) GetCompanyByIDRecord(companyID int64) (*Company, error) {
	row := s.database.QueryRow(`
		SELECT id, registered_id, name, tax_code, activity_code_id, address, owner_id
		FROM companies
		WHERE id = $1
	`, companyID)

	company, err := scanCompany(row)
	if err == sql.ErrNoRows {
		return nil, ErrCompanyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting company by id: %w", err)
	}

	return company, nil
}

func (s *Server) GetCompaniesRecords() ([]*Company, error) {
	rows, err := s.database.Query(`
		SELECT id, registered_id, name, tax_code, activity_code_id, address, owner_id
		FROM companies
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("listing companies: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var companies []*Company
	for rows.Next() {
		company, err := scanCompany(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning company: %w", err)
		}
		companies = append(companies, company)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating companies: %w", err)
	}

	return companies, nil
}

func (s *Server) UpdateCompanyRecord(company Company) (*Company, error) {
	tx, err := s.database.Begin()
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var companyExists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)`, company.Id).Scan(&companyExists); err != nil {
		return nil, fmt.Errorf("checking company existence: %w", err)
	}
	if !companyExists {
		return nil, ErrCompanyNotFound
	}

	var ownerExists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1)`, company.Owner_id).Scan(&ownerExists); err != nil {
		return nil, fmt.Errorf("checking owner existence: %w", err)
	}
	if !ownerExists {
		return nil, ErrCompanyOwnerNotFound
	}

	if company.Activity_code_id != 0 {
		var activityCodeExists bool
		if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM activity_codes WHERE id = $1)`, company.Activity_code_id).Scan(&activityCodeExists); err != nil {
			return nil, fmt.Errorf("checking activity code existence: %w", err)
		}
		if !activityCodeExists {
			return nil, ErrCompanyActivityCodeNotFound
		}
	}

	var row *sql.Row
	if company.Activity_code_id == 0 {
		row = tx.QueryRow(`
			UPDATE companies
			SET name = $1, activity_code_id = NULL, address = $2, owner_id = $3
			WHERE id = $4
			RETURNING id, registered_id, name, tax_code, activity_code_id, address, owner_id
		`, company.Name, company.Address, company.Owner_id, company.Id)
	} else {
		row = tx.QueryRow(`
			UPDATE companies
			SET name = $1, activity_code_id = $2, address = $3, owner_id = $4
			WHERE id = $5
			RETURNING id, registered_id, name, tax_code, activity_code_id, address, owner_id
		`, company.Name, company.Activity_code_id, company.Address, company.Owner_id, company.Id)
	}

	updated, err := scanCompany(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCompanyNotFound
		}
		return nil, fmt.Errorf("updating company: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return updated, nil
}

func scanCard(scanner interface{ Scan(dest ...any) error }) (*Card, error) {
	var card Card
	err := scanner.Scan(
		&card.Id,
		&card.Number,
		&card.Type,
		&card.Brand,
		&card.Creation_date,
		&card.Valid_until,
		&card.Account_number,
		&card.Cvv,
		&card.Card_limit,
		&card.Status,
	)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func scanCardRequest(scanner interface{ Scan(dest ...any) error }) (*CardRequest, error) {
	var req CardRequest
	err := scanner.Scan(
		&req.Id,
		&req.Account_number,
		&req.Type,
		&req.Brand,
		&req.Token,
		&req.ExpirationDate,
		&req.Complete,
		&req.Email,
	)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *Server) CreateCardRecord(card Card) (*Card, error) {
	row := s.database.QueryRow(`
		INSERT INTO cards (number, type, brand, creation_date, valid_until, account_number, cvv, card_limit, status)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, $4, $5, $6, $7, $8)
		RETURNING id, number, type, brand, creation_date, valid_until, account_number, cvv, card_limit, status
	`, card.Number, card.Type, card.Brand, card.Valid_until, card.Account_number, card.Cvv, card.Card_limit, card.Status)
	return scanCard(row)
}

func (s *Server) GetCardsRecords() ([]*Card, error) {
	rows, err := s.database.Query(`
		SELECT id, number, type, brand, creation_date, valid_until, account_number, cvv, card_limit, status
		FROM cards
	`)
	if err != nil {
		return nil, fmt.Errorf("listing cards: %w", err)
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func (s *Server) BlockCardRecord(cardID int64) error {
	res, err := s.database.Exec(`UPDATE cards SET status = $1 WHERE id = $2`, Blocked, cardID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("card not found")
	}
	return nil
}

func (s *Server) CreateCardRequestRecord(req CardRequest) (*CardRequest, error) {
	row := s.database.QueryRow(`
		INSERT INTO card_requests (account_number, type, brand, token, expiration_date, complete, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, account_number, type, brand, token, expiration_date, complete, email
	`, req.Account_number, req.Type, req.Brand, req.Token, req.ExpirationDate, req.Complete, req.Email)
	return scanCardRequest(row)
}

func (s *Server) GetCardRequestByToken(token string) (*CardRequest, error) {
	row := s.database.QueryRow(`
		SELECT id, account_number, type, brand, token, expiration_date, complete, email
		FROM card_requests
		WHERE token = $1 AND complete = false
	`, token)
	return scanCardRequest(row)
}

func (s *Server) MarkCardRequestFulfilled(id int64) error {
	_, err := s.database.Exec(`UPDATE card_requests SET complete = true WHERE id = $1`, id)
	return err
}

func (s *Server) GetAccountByNumberRecord(number string) (*Account, error) {
	var acc Account
	err := s.database.QueryRow(`
		SELECT id, number, name, owner, balance, currency, active, owner_type, account_type,
		       maintainance_cost, daily_limit, monthly_limit, daily_expenditure, monthly_expenditure,
		       created_by, created_at, valid_until
		FROM accounts WHERE number = $1
	`, number).Scan(
		&acc.Id, &acc.Number, &acc.Name, &acc.Owner, &acc.Balance, &acc.Currency, &acc.Active, &acc.Owner_type, &acc.Account_type,
		&acc.Maintainance_cost, &acc.Daily_limit, &acc.Monthly_limit, &acc.Daily_expenditure, &acc.Monthly_expenditure,
		&acc.Created_by, &acc.Created_at, &acc.Valid_until,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("account not found")
	}
	return &acc, err
}

func (s *Server) CountActiveCardsByAccountNumber(accountNumber string) (int, error) {
	var count int
	err := s.database.QueryRow(`
		SELECT COUNT(*) FROM cards
		WHERE account_number = $1 AND status != $2
	`, accountNumber, Deactivated).Scan(&count)
	return count, err
}

func (s *Server) IsAuthorizedParty(email string, accountNumber string) (bool, error) {
	var exists bool
	err := s.database.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM authorized_party ap
			WHERE ap.email = $1 AND EXISTS (
				SELECT 1 FROM accounts a WHERE a.number = $2
			)
		)
	`, email, accountNumber).Scan(&exists)
	return exists, err
}

func (s *Server) GetCardByNumberRecord(cardNumber string) (*Card, error) {
	row := s.database.QueryRow(`
		SELECT id, number, type, brand, creation_date, valid_until, account_number, cvv, card_limit, status
		FROM cards WHERE number = $1
	`, cardNumber)
	return scanCard(row)
}

func (s *Server) GetCardByIDRecord(id int64) (*Card, error) {
	row := s.database.QueryRow(`
		SELECT id, number, type, brand, creation_date, valid_until, account_number, cvv, card_limit, status
		FROM cards WHERE id = $1
	`, id)
	return scanCard(row)
}
