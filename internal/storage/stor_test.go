package storage_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/filatkinen/socialnet/internal/common"
	"github.com/filatkinen/socialnet/internal/storage"
	"github.com/stretchr/testify/require"
)

func newStringPointer(s string) *string {
	return &s
}

func newTimePointer(s time.Time) *time.Time {
	return &s
}

func toJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

const deleteSet = false

var (
	UUID1, _ = storage.UUID()
	UUID2, _ = storage.UUID()
	UUID3, _ = storage.UUID()
	users    = []*storage.User{
		{
			Id:        UUID1,
			BirthDate: newTimePointer(time.Now().Add(-time.Hour * 24 * 365 * 21).Truncate(time.Hour * 24).UTC()),
			Sex:       newStringPointer("male"),
		},
		{
			Id:         UUID2,
			FirstName:  "Mariya1",
			SecondName: newStringPointer("Filimonova1"),
			Sex:        newStringPointer("female"),
			BirthDate:  newTimePointer(time.Now().Add(-time.Hour * 34 * 365 * 20).Truncate(time.Hour * 24).UTC()),
		},
		{
			Id:         UUID3,
			FirstName:  "Mariya2",
			SecondName: newStringPointer("Filimonova2"),
			Sex:        newStringPointer("female"),
			BirthDate:  newTimePointer(time.Now().Add(-time.Hour * 34 * 365 * 18).Truncate(time.Hour * 24).UTC()),
		},
	}
)

func runTestStorage(t *testing.T, ctx context.Context, store storage.Storage) {
	for i := range users {
		err := store.UserAdd(ctx, users[i])
		require.NoError(t, err)
	}

	// update + get
	users[0].City = newStringPointer("Moskva")
	err := store.UserUpdate(ctx, users[0])
	require.NoError(t, err)

	user, err := store.UserGet(ctx, users[0].Id)
	require.NoError(t, err)
	require.Equal(t, users[0], user)

	// deleteSet
	err = store.UserDelete(ctx, users[0].Id)
	require.NoError(t, err)
	_, err = store.UserGet(ctx, users[0].Id)
	require.Equal(t, storage.ErrRecordNotFound, err)

	// tokens
	user = users[1]
	token1, _ := common.TokenGenerator()
	token2, _ := common.TokenGenerator()
	exp := time.Now().Add(time.Hour * 2).Round(time.Minute).UTC()

	t1 := &storage.Token{
		Hash:   common.Hasher(token1),
		Expire: exp,
		UserID: user.Id,
	}
	t2 := &storage.Token{
		Hash:   common.Hasher(token2),
		Expire: exp,
		UserID: user.Id,
	}
	err = store.TokenAdd(ctx, t1)
	require.NoError(t, err)
	err = store.TokenAdd(ctx, t2)
	require.NoError(t, err)

	tget1, err := store.TokenGet(ctx, t1.Hash)
	require.NoError(t, err)
	require.Equal(t, t1, tget1)
	require.Equal(t, common.Hasher(token1), tget1.Hash)
	tget2, err := store.TokenGet(ctx, t2.Hash)
	require.NoError(t, err)
	require.Equal(t, t2, tget2)
	require.Equal(t, common.Hasher(token2), tget2.Hash)

	// deleteSet
	err = store.TokenDelete(ctx, t1.Hash)
	require.NoError(t, err)
	_, err = store.TokenGet(ctx, t1.Hash)
	require.Equal(t, storage.ErrRecordNotFound, err)
	// deleteSet all
	err = store.TokenDeleteAllUser(ctx, user.Id)
	require.NoError(t, err)
	err = store.TokenDeleteAllUser(ctx, user.Id)
	require.Equal(t, storage.ErrRecordNotFound, err)

	// pass redo twice + update
	for i := 0; i < 2; i++ {
		pass := "password" + fmt.Sprintf("%d", i)
		encryptedPass, err := common.HashPassword(pass)
		require.NoError(t, err)

		cred := storage.UserCredential{
			UserID:     user.Id,
			PassBcrypt: encryptedPass,
		}
		err = store.UserCredentialSet(ctx, &cred)
		require.NoError(t, err)
		credGet, err := store.UserCredentialGet(ctx, user.Id)
		require.NoError(t, err)
		require.Equal(t, cred, *credGet)
		require.True(t, common.CheckPasswordHash(pass, credGet.PassBcrypt))
	}

	// deleteSet pass
	err = store.UserCredentialDelete(ctx, user.Id)
	require.NoError(t, err)
	_, err = store.UserCredentialGet(ctx, user.Id)
	require.Equal(t, storage.ErrRecordNotFound, err)

	// search
	users, _ := store.UserSearch(ctx, "Mar", "Fil")
	for i := range users {
		fmt.Println(toJSON(users[i]))
	}

	if deleteSet {
		err = store.UserDelete(ctx, user.Id)
		require.NoError(t, err)
	}
}
