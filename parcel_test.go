package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Создаем таблицу
	_, err = db.Exec(`
		CREATE TABLE parcel (
			number INTEGER PRIMARY KEY AUTOINCREMENT,
			client INTEGER,
			status TEXT,
			address TEXT,
			created_at TEXT
		)
	`)
	require.NoError(t, err)

	return db
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {

	db := setupTestDB(t)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// get
	parcel.Number = id // обновляем ID для сравнения
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel, storedParcel)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	// check deleted
	_, err = store.Get(id)
	require.Error(t, err)

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewParcelStore(db)

	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	updated, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, updated.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewParcelStore(db)

	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)

	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	updated, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, updated.Status)

}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	for _, sp := range storedParcels {
		expected, ok := parcelMap[sp.Number]
		require.True(t, ok, "unexpected parcel number")
		require.Equal(t, expected, sp)
	}
}
