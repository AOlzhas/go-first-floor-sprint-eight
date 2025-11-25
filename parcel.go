package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec(
		`INSERT INTO parcel (client, status, address, created_at) 
         VALUES (?, ?, ?, ?)`,
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	// верните идентификатор последней добавленной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка

	p := Parcel{}

	row := s.db.QueryRow(
		`SELECT number, client, status, address, created_at 
         FROM parcel WHERE number = ?`,
		number,
	)

	// заполните объект Parcel данными из таблицы
	if err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
		return p, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк

	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at 
         FROM parcel WHERE client = ? ORDER BY number`,
		client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// заполните срез Parcel данными из таблицы
	var res []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec(
		`UPDATE parcel SET status = ? WHERE number = ?`,
		status, number,
	)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered

	var status string
	err := s.db.QueryRow(
		`SELECT status FROM parcel WHERE number = ?`,
		number,
	).Scan(&status)
	if err != nil {
		return err
	}

	if status != ParcelStatusRegistered {
		return errors.New("address can be changed only for registered parcels")
	}

	_, err = s.db.Exec(
		`UPDATE parcel SET address = ? WHERE number = ?`,
		address, number,
	)
	return err
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered

	var status string
	err := s.db.QueryRow(
		`SELECT status FROM parcel WHERE number = ?`,
		number,
	).Scan(&status)
	if err != nil {
		return err
	}

	if status != ParcelStatusRegistered {
		return errors.New("parcel can be deleted only in registered status")
	}

	_, err = s.db.Exec(
		`DELETE FROM parcel WHERE number = ?`,
		number,
	)
	return err
}
