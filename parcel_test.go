package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func setupTestDB(t *testing.T) (*sql.DB, ParcelStore) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM parcel") // очистка после теста
		db.Close()
	})
	return db, NewParcelStore(db)
}

func TestAddGetDelete(t *testing.T) {
	_, store := setupTestDB(t)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	got, err := store.Get(id)
	require.NoError(t, err)

	assert.Equal(t, id, got.Number)
	assert.Equal(t, parcel.Client, got.Client)
	assert.Equal(t, parcel.Status, got.Status)
	assert.Equal(t, parcel.Address, got.Address)
	assert.Equal(t, parcel.CreatedAt[:19], got.CreatedAt[:19]) // сравнение без миллисекунд

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	assert.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	_, store := setupTestDB(t)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	got, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, got.Address)
	assert.Equal(t, parcel.Client, got.Client)
	assert.Equal(t, parcel.Status, got.Status)
	assert.Equal(t, parcel.CreatedAt[:19], got.CreatedAt[:19])
}

func TestSetStatus(t *testing.T) {
	_, store := setupTestDB(t)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	got, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, got.Status)
}

func TestGetByClient(t *testing.T) {
	_, store := setupTestDB(t)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}
	client := randRange.Intn(10_000_000)

	for i := range parcels {
		parcels[i].Client = client
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))

	for _, parcel := range storedParcels {
		original, ok := parcelMap[parcel.Number]
		assert.True(t, ok)
		assert.Equal(t, original.Client, parcel.Client)
		assert.Equal(t, original.Status, parcel.Status)
		assert.Equal(t, original.Address, parcel.Address)
		assert.Equal(t, original.CreatedAt[:19], parcel.CreatedAt[:19])
		assert.Equal(t, original.Number, parcel.Number)
	}
}
