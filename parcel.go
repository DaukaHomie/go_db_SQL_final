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
	res, err := s.db.Exec(`
		INSERT INTO parcel (client, status, address, created_at)
		VALUES (?, ?, ?, ?)
	`, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	// верните идентификатор последней добавленной записи
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	row := s.db.QueryRow(`
		SELECT number, client, status, address, created_at
		FROM parcel
		WHERE number = ?
	`, number)
	// здесь из таблицы должна вернуться только одна строка

	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)

	return p, err
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	rows, err := s.db.Query(`
		SELECT number, client, status, address, created_at
		FROM parcel
		WHERE client = ?
	`, client)
	// здесь из таблицы может вернуться несколько строк
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// заполните срез Parcel данными из таблицы
	var res []Parcel
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec(`
		UPDATE parcel
		SET status = ?
		WHERE number = ?
	`, status, number)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	row := s.db.QueryRow(`
		SELECT status FROM parcel
		WHERE number = ?
	`, number)
	// менять адрес можно только если значение статуса registered
	var status string
	err := row.Scan(&status)
	if err != nil {
		return err
	}
	if status != ParcelStatusRegistered {
		return errors.New("нельзя изменить адрес: посылка уже отправлена или доставлена")
	}

	_, err = s.db.Exec(`
		UPDATE parcel
		SET address = ?
		WHERE number = ?
	`, address, number)

	return err
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	row := s.db.QueryRow(`
		SELECT status FROM parcel
		WHERE number = ?
	`, number)

	var status string
	err := row.Scan(&status)
	if err != nil {
		return err
	}

	// удалять строку можно только если значение статуса registered
	if status != ParcelStatusRegistered {
		return errors.New("нельзя удалить: посылка уже отправлена или доставлена")
	}

	_, err = s.db.Exec(`
		DELETE FROM parcel
		WHERE number = ?
	`, number)
	return err
}
