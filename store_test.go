package imagine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/risico/imagine"
)


func TestStores(t *testing.T) {
    t.Run("local", testLocalStore)
}

func testLocalStore(t *testing.T) {
    store, err := imagine.NewLocalStorage(imagine.LocalStoreParams{
        Path: "/tmp/imagine",
    })
    assert.NoError(t, err)

    err = store.Set("test", []byte("test"))
    assert.NoError(t, err)

    data, ok, err := store.Get("test")
    assert.NoError(t, err)
    assert.True(t, ok)
    assert.Equal(t, []byte("test"), data)

    err = store.Delete("test")
    assert.NoError(t, err)

    data, ok, err = store.Get("test")
    assert.ErrorIs(t, err, imagine.ErrKeyNotFound)
    assert.False(t, ok)
    assert.Nil(t, data)

    err = store.Close()
    assert.NoError(t, err)
}
