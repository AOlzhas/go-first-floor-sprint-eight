package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func newTestStore(t *testing.T) (ParcelStore, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE parcel (
			number     INTEGER PRIMARY KEY AUTOINCREMENT,
			client     INTEGER NOT NULL,
			status     TEXT    NOT NULL,
			address    TEXT    NOT NULL,
			created_at TEXT    NOT NULL
		);
	`)
	require.NoError(t, err)

	return NewParcelStore(db), db
}

func TestAddAndGetParcel(t *testing.T) {
	store, db := newTestStore(t)
	defer db.Close()

	createdAt := time.Now().UTC().Format(time.RFC3339)

	parcel := Parcel{
		Client:    123,
		Status:    ParcelStatusRegistered,
		Address:   "Some street, 1",
		CreatedAt: createdAt,
	}

	id, err := store.Add(parcel)
	require.NoError(t, err)

	stored, err := store.Get(id)
	require.NoError(t, err)

	// заполняем Number, чтобы сравнивать целиком структуру
	parcel.Number = id
	assert.Equal(t, parcel, stored)
}

func TestDeleteParcel_RemovesRow(t *testing.T) {
	store, db := newTestStore(t)
	defer db.Close()

	createdAt := time.Now().UTC().Format(time.RFC3339)
	parcel := Parcel{
		Client:    1,
		Status:    ParcelStatusRegistered,
		Address:   "Address",
		CreatedAt: createdAt,
	}

	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.Delete(id)
	require.NoError(t, err)

	// проверьте, что посылку больше нельзя получить из БД
	_, err = store.Get(id)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress_Registered(t *testing.T) {
	store, db := newTestStore(t)
	defer db.Close()

	createdAt := time.Now().UTC().Format(time.RFC3339)
	parcel := Parcel{
		Client:    42,
		Status:    ParcelStatusRegistered,
		Address:   "Old address",
		CreatedAt: createdAt,
	}

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "New address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	stored, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, stored.Address)
}

func TestSetAddress_NotRegistered(t *testing.T) {
	store, db := newTestStore(t)
	defer db.Close()

	createdAt := time.Now().UTC().Format(time.RFC3339)
	parcel := Parcel{
		Client:    42,
		Status:    ParcelStatusSent,
		Address:   "Old address",
		CreatedAt: createdAt,
	}

	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.SetAddress(id, "New address")
	require.Error(t, err)
}

func TestSetStatus(t *testing.T) {
	store, db := newTestStore(t)
	defer db.Close()

	createdAt := time.Now().UTC().Format(time.RFC3339)
	parcel := Parcel{
		Client:    7,
		Status:    ParcelStatusRegistered,
		Address:   "Address",
		CreatedAt: createdAt,
	}

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	stored, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newStatus, stored.Status)
}

func TestGetByClient(t *testing.T) {
	store, db := newTestStore(t)
	defer db.Close()

	clientID := 100
	now := time.Now().UTC().Format(time.RFC3339)

	parcels := []Parcel{
		{
			Client:    clientID,
			Status:    ParcelStatusRegistered,
			Address:   "Address 1",
			CreatedAt: now,
		},
		{
			Client:    clientID,
			Status:    ParcelStatusRegistered,
			Address:   "Address 2",
			CreatedAt: now,
		},
	}

	parcelMap := make(map[int]Parcel)
	for _, p := range parcels {
		id, err := store.Add(p)
		require.NoError(t, err)
		p.Number = id
		parcelMap[id] = p
	}

	// посылка другого клиента — не должна попасть в выборку
	_, err := store.Add(Parcel{
		Client:    999,
		Status:    ParcelStatusRegistered,
		Address:   "Other client",
		CreatedAt: now,
	})
	require.NoError(t, err)

	storedParcels, err := store.GetByClient(clientID)
	require.NoError(t, err)

	assert.Len(t, storedParcels, len(parcels))

	for _, parcel := range storedParcels {
		expected, ok := parcelMap[parcel.Number]
		assert.True(t, ok, "unexpected parcel with number %d", parcel.Number)
		assert.Equal(t, expected, parcel)
	}
}
