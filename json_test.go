package jsrest_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/gopatchy/jsrest"
	"github.com/stretchr/testify/require"
)

type testType struct {
	Text1 string
}

func TestRead(t *testing.T) {
	t.Parallel()

	body := bytes.NewBufferString(`{"text1":"foo"}`)

	req, err := http.NewRequest(http.MethodGet, "xyz", body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	obj := &testType{}

	err = jsrest.Read(req, obj)
	require.NoError(t, err)
	require.Equal(t, "foo", obj.Text1)
}

func TestReadContentTypeParams(t *testing.T) {
	t.Parallel()

	body := bytes.NewBufferString(`{"text1":"bar"}`)

	req, err := http.NewRequest(http.MethodGet, "xyz", body) //nolint:noctx
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	obj := &testType{}

	err = jsrest.Read(req, obj)
	require.NoError(t, err)
	require.Equal(t, "bar", obj.Text1)
}
