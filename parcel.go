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
	// добавление строки в таблицу parcel
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
	// чтение строки по заданному number
	p := Parcel{}

	row := s.db.QueryRow(
		`SELECT number, client, status, address, created_at 
         FROM parcel WHERE number = ?`,
		number,
	)

	// важно: при ошибке возвращаем нулевой Parcel{}, а не p
	if err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
		return Parcel{}, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// чтение строк по client
	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at 
         FROM parcel WHERE client = ? ORDER BY number`,
		client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	// обновление статуса
	_, err := s.db.Exec(
		`UPDATE parcel SET status = ? WHERE number = ?`,
		status, number,
	)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// меняем адрес только если статус registered — делаем это в одном запросе
	res, err := s.db.Exec(
		`UPDATE parcel 
         SET address = ?
         WHERE number = ? AND status = ?`,
		address, number, ParcelStatusRegistered,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		// либо не найдено, либо статус не registered
		return errors.New("address can be changed only for registered parcels")
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// удалять можно только если статус registered — тоже одним запросом
	res, err := s.db.Exec(
		`DELETE FROM parcel 
         WHERE number = ? AND status = ?`,
		number, ParcelStatusRegistered,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		// либо записи нет, либо статус не registered
		return errors.New("parcel can be deleted only in registered status")
	}

	return nil
}
