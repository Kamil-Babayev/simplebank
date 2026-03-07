package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAccount(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		chackResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "Success",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.chackResponse(t, recorder)
		})
	}
}

func TestCreateAccount(t *testing.T) {
	request := createAccountRequest{
		Owner:    util.RandomOwner(),
		Currency: util.RandomCurrency(),
	}

	expectedRes := db.Account{
		ID:        util.RandomInt(1, 1000),
		Owner:     request.Owner,
		Currency:  request.Currency,
		Balance:   0,
		CreatedAt: time.Now(),
	}

	testCases := []struct {
		name          string
		request       createAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		chackResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "Success",
			request: request,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{
					Owner:    request.Owner,
					Balance:  0,
					Currency: request.Currency})).
					Times(1).
					Return(expectedRes, nil)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, expectedRes)
			},
		},
		{
			name:    "BadRequest",
			request: createAccountRequest{},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:    "InternalError",
			request: request,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts")

			encodedRequest, err := json.Marshal(tc.request)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(encodedRequest))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.chackResponse(t, recorder)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	request := updateAccountRequest{
		Owner:    util.RandomOwner(),
		Currency: util.RandomCurrency(),
	}

	expectedRes := db.Account{
		ID:        util.RandomInt(1, 1000),
		Owner:     request.Owner,
		Currency:  request.Currency,
		Balance:   util.RandomBalance(),
		CreatedAt: time.Now(),
	}

	testCases := []struct {
		name          string
		accountID     int64
		request       updateAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		chackResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "Success",
			accountID: expectedRes.ID,
			request:   request,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateAccountData(gomock.Any(), gomock.Eq(db.UpdateAccountDataParams{
					ID:       expectedRes.ID,
					Owner:    sql.NullString{String: request.Owner, Valid: true},
					Currency: sql.NullString{String: request.Currency, Valid: true},
				})).Times(1).Return(expectedRes, nil)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, expectedRes)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			request:   updateAccountRequest{},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateAccountData(gomock.Any(), gomock.Any()).Times(0)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: expectedRes.ID,
			request:   request,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateAccountData(gomock.Any(), gomock.Eq(db.UpdateAccountDataParams{
					ID:       expectedRes.ID,
					Owner:    sql.NullString{String: request.Owner, Valid: true},
					Currency: sql.NullString{String: request.Currency, Valid: true},
				})).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: expectedRes.ID,
			request:   request,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateAccountData(gomock.Any(), gomock.Eq(db.UpdateAccountDataParams{
					ID:       expectedRes.ID,
					Owner:    sql.NullString{String: request.Owner, Valid: true},
					Currency: sql.NullString{String: request.Currency, Valid: true},
				})).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)

			encodedRequest, err := json.Marshal(tc.request)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(encodedRequest))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.chackResponse(t, recorder)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		chackResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "Success",
			accountID: 1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(int64(1))).Times(1).Return(nil)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: 1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(int64(1))).Times(1).Return(sql.ErrNoRows)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: 1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(int64(1))).Times(1).Return(sql.ErrConnDone)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.chackResponse(t, recorder)
		})
	}
}

func TestListAccounts(t *testing.T) {
	request := listAccountsRequest{
		PageID:   1,
		PageSize: 5,
	}

	expectedRes := []db.Account{
		{
			ID:        util.RandomInt(1, 1000),
			Owner:     util.RandomOwner(),
			Currency:  util.RandomCurrency(),
			Balance:   util.RandomBalance(),
			CreatedAt: time.Now(),
		},
	}

	testCases := []struct {
		name          string
		request       listAccountsRequest
		buildStubs    func(store *mockdb.MockStore)
		chackResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "Success",
			request: request,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Any()).Times(1).Return(expectedRes, nil)
			},
			chackResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		//TODO: more cases
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/accounts", nil)
			require.NoError(t, err)

			q := req.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.request.PageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.request.PageSize))
			req.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, req)
			tc.chackResponse(t, recorder)
		})
	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomBalance(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, expected db.Account) {
	t.Helper()
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)

	expected.CreatedAt = expected.CreatedAt.Round(0)
	gotAccount.CreatedAt = gotAccount.CreatedAt.Round(0)
	require.Equal(t, expected, gotAccount)
}
