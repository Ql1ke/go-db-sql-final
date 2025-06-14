package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	schema := `
    CREATE TABLE IF NOT EXISTS parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER NOT NULL,
        status TEXT NOT NULL,
        address TEXT NOT NULL,
        created_at TEXT NOT NULL
    );`
	if _, err := db.Exec(schema); err != nil {
		panic(fmt.Sprintf("cannot create table: %v", err))
	}
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec(
		`INSERT INTO parcel(client,status,address,created_at) VALUES(?,?,?,?);`,
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	var p Parcel
	row := s.db.QueryRow(
		`SELECT number,client,status,address,created_at 
         FROM parcel 
         WHERE number = ?;`, number,
	)
	if err := row.Scan(
		&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return p, fmt.Errorf("parcel %d not found", number)
		}
		return p, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(
		`SELECT number,client,status,address,created_at 
         FROM parcel 
         WHERE client = ? 
         ORDER BY number;`, client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(
			&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (s ParcelStore) SetStatus(number int, status string) error {
	res, err := s.db.Exec(
		`UPDATE parcel SET status = ? WHERE number = ?;`,
		status, number,
	)
	if err != nil {
		return err
	}
	if cnt, _ := res.RowsAffected(); cnt == 0 {
		return fmt.Errorf("parcel %d not found", number)
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	res, err := s.db.Exec(
		`UPDATE parcel SET address = ? 
         WHERE number = ? AND status = ?;`,
		address, number, ParcelStatusRegistered,
	)
	if err != nil {
		return err
	}
	if cnt, _ := res.RowsAffected(); cnt == 0 {
		return errors.New("cannot change address: only registered parcel")
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	res, err := s.db.Exec(
		`DELETE FROM parcel 
         WHERE number = ? AND status = ?;`,
		number, ParcelStatusRegistered,
	)
	if err != nil {
		return err
	}
	if cnt, _ := res.RowsAffected(); cnt == 0 {
		return errors.New("cannot delete parcel: only registered parcel")
	}
	return nil
}
