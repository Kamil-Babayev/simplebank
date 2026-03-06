package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStore_TransferTx(t *testing.T) {
	store := NewStore(testDB)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParam{
				FromAccountID: acc1.ID,
				ToAccountID:   acc2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	//Map for storing instances of account balance differences
	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// Check transfer

		transfer := result.Transfer
		//Struct is not empty
		require.NotEmpty(t, transfer)
		//Ids are equal
		require.Equal(t, acc1.ID, transfer.FromAccountID)
		require.Equal(t, acc2.ID, transfer.ToAccountID)
		//Amount is equal
		require.Equal(t, amount, transfer.Amount)
		//Generated values are not zero
		require.NotZero(t, transfer.CreatedAt)
		require.NotZero(t, transfer.ID)
		//Data fetches without error
		_, err = testQueries.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// Check entries

		fromEntry := result.FromEntry
		toEntry := result.ToEntry
		//Struct is not empty
		require.NotEmpty(t, fromEntry)
		require.NotEmpty(t, toEntry)
		//Ids are equal
		require.Equal(t, acc1.ID, fromEntry.AccountID)
		require.Equal(t, acc2.ID, toEntry.AccountID)
		//Amount is set correctly
		require.Equal(t, amount, -fromEntry.Amount)
		require.Equal(t, amount, toEntry.Amount)
		//Generated values are not zero
		require.NotZero(t, fromEntry.CreatedAt)
		require.NotZero(t, toEntry.CreatedAt)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, toEntry.ID)
		//Data fetches without error
		_, err = testQueries.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)
		_, err = testQueries.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// Check Accounts

		fromAccount := result.FromAccount
		toAccount := result.ToAccount
		//Struct is not empty
		require.NotEmpty(t, fromAccount)
		require.NotEmpty(t, toAccount)
		//Ids are equal
		require.Equal(t, acc1.ID, fromAccount.ID)
		require.Equal(t, acc2.ID, toAccount.ID)
		//Owners are equal
		require.Equal(t, acc1.Owner, fromAccount.Owner)
		require.Equal(t, acc2.Owner, toAccount.Owner)
		//Timestamps are equal
		require.WithinDuration(t, acc1.CreatedAt, fromAccount.CreatedAt, time.Second)
		require.WithinDuration(t, acc2.CreatedAt, toAccount.CreatedAt, time.Second)
		//Balance updated correctly
		diff1 := acc1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - acc2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)
		k := int(diff1 / amount)
		require.True(t, k >= 0 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
		//Data fetches without error
		_, err = testQueries.GetAccount(context.Background(), fromAccount.ID)
		require.NoError(t, err)
		_, err = testQueries.GetAccount(context.Background(), toAccount.ID)
		require.NoError(t, err)
	}

	updateAccount1, err := testQueries.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)
	updateAccount2, err := testQueries.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)
	require.Equal(t, acc1.Balance-int64(n)*amount, updateAccount1.Balance)
	require.Equal(t, acc2.Balance+int64(n)*amount, updateAccount2.Balance)
}

func TestStore_TransferTxDeadLock(t *testing.T) {
	store := NewStore(testDB)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	n := 10
	amount := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccount := acc1.ID
		toAccount := acc2.ID

		if i%2 == 1 {
			fromAccount = acc2.ID
			toAccount = acc1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParam{
				FromAccountID: fromAccount,
				ToAccountID:   toAccount,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	updateAccount1, err := testQueries.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)
	updateAccount2, err := testQueries.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)
	require.Equal(t, acc1.Balance, updateAccount1.Balance)
	require.Equal(t, acc2.Balance, updateAccount2.Balance)
}
