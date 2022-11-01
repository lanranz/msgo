package render

import (
	"html/template"
	"msgo/internal/bytesconv"
	"net/http"
)

type HTML struct {
	Data       any
	Name       string
	Template   *template.Template
	IsTemplate bool
}
type HTMLRender struct {
	//需要修改的值就用指针
	Template *template.Template
}

func (h *HTML) Render(w http.ResponseWriter) error {
	h.WriteContentType(w)
	if h.IsTemplate {
		err := h.Template.ExecuteTemplate(w, h.Name, h.Data)
		return err
	}
	_, err := w.Write(bytesconv.StringToBytes(h.Data.(string)))
	return err
}

func (h *HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html; charset=utf-8")
}
