package jsrest_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gopatchy/jsrest"
	"github.com/stretchr/testify/require"
)

func TestFromError(t *testing.T) {
	t.Parallel()

	e1 := errors.New("error 1") //nolint:goerr113
	e2 := jsrest.Errorf(jsrest.ErrBadGateway, "error 2: %w", e1)

	require.ErrorIs(t, e2, e1)
	require.ErrorIs(t, e2, jsrest.ErrBadGateway)

	je := jsrest.ToJSONError(e2)

	require.Equal(t, http.StatusBadGateway, je.Code)
	require.Contains(t, je.Messages, "error 1")
	require.Contains(t, je.Messages, "error 2: error 1")
	require.Contains(t, je.Messages, "[502] Bad Gateway")
}

func TestFromErrors(t *testing.T) {
	t.Parallel()

	e1 := jsrest.Errorf(jsrest.ErrForbidden, "error 1")
	e2 := errors.New("error 2") //nolint:goerr113
	e3 := jsrest.Errorf(jsrest.ErrBadGateway, "error 3: %w + %w", e1, e2)

	require.ErrorIs(t, e3, e1)
	require.ErrorIs(t, e3, e2)
	require.ErrorIs(t, e3, jsrest.ErrForbidden)

	je := jsrest.ToJSONError(e3)

	require.Equal(t, http.StatusForbidden, je.Code)
	require.Contains(t, je.Messages, "error 1")
	require.Contains(t, je.Messages, "error 2")
	require.Contains(t, je.Messages, "error 3: error 1 + error 2")
	require.Contains(t, je.Messages, "[403] Forbidden")
}
