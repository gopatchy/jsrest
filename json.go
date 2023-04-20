package jsrest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gopatchy/metadata"
	"github.com/vfaronov/httpheader"
)

var ErrUnsupportedContentType = errors.New("unsupported Content-Type")

func Read(r *http.Request, obj any) error {
	contentType, _ := httpheader.ContentType(r.Header)

	switch contentType {
	case "":
		fallthrough
	case "application/json":
		break

	default:
		return Errorf(ErrUnsupportedMediaType, "Content-Type: %s", contentType)
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(obj)
	if err != nil {
		return Errorf(ErrBadRequest, "decode JSON request body failed (%w)", err)
	}

	return nil
}

func Write(w http.ResponseWriter, obj any) error {
	m := metadata.GetMetadata(obj)

	w.Header().Set("Content-Type", "application/json")
	httpheader.SetETag(w.Header(), httpheader.EntityTag{Opaque: m.ETag})

	enc := json.NewEncoder(w)

	err := enc.Encode(obj)
	if err != nil {
		return Errorf(ErrInternalServerError, "encode JSON response failed (%w)", err)
	}

	return nil
}

func WriteList(w http.ResponseWriter, list []any, etag string) error {
	w.Header().Set("Content-Type", "application/json")
	httpheader.SetETag(w.Header(), httpheader.EntityTag{Opaque: etag})

	enc := json.NewEncoder(w)

	err := enc.Encode(list)
	if err != nil {
		return Errorf(ErrInternalServerError, "encode JSON response failed (%w)", err)
	}

	return nil
}
