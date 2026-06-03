package repository

import "database/sql"

type postgresAdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) AdminRepository {
	return &postgresAdminRepository{db: db}
}

func (r *postgresAdminRepository) GetPasswordByLogin(login string) (string, error) {
	var password string
	err := r.db.QueryRow(`SELECT password FROM admins WHERE login = $1`, login).Scan(&password)
	return password, err
}

func (r *postgresAdminRepository) GetLoginByID(id string) (string, error) {
	var login string
	err := r.db.QueryRow(`SELECT login FROM admins WHERE id = $1`, id).Scan(&login)
	return login, err
}

func (r *postgresAdminRepository) GetAll() ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`SELECT id, login FROM admins ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id int
		var login string
		if err := rows.Scan(&id, &login); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"ID":    id,
			"Логин": login,
		})
	}
	return result, nil
}

func (r *postgresAdminRepository) Create(login, hashedPassword string) error {
	_, err := r.db.Exec(`INSERT INTO admins (login, password) VALUES ($1, $2)`, login, hashedPassword)
	return err
}

func (r *postgresAdminRepository) Update(id, login string, hashedPassword *string) error {
	if hashedPassword != nil {
		_, err := r.db.Exec(`UPDATE admins SET login = $1, password = $2 WHERE id = $3`, login, *hashedPassword, id)
		return err
	}
	_, err := r.db.Exec(`UPDATE admins SET login = $1 WHERE id = $2`, login, id)
	return err
}

func (r *postgresAdminRepository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(`DELETE FROM admins WHERE id = $1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`SELECT setval('admins_id_seq', COALESCE((SELECT MAX(id) FROM admins), 0) + 1, false)`); err != nil {
		return err
	}
	return tx.Commit()
}
