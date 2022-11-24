package binding

import (
	"encoding/xml"
	"io"
	"net/http"
)

type xmlBinding struct {
}

func (b xmlBinding) Name() string {
	return "xml"
}

func (xmlBinding) Bind(r *http.Request, obj any) error {
	return decodeXML(r.Body, obj)
}

func decodeXML(r io.Reader, obj any) error {
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validate(obj)
}
