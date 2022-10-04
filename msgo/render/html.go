package render

import "html/template"

type HTMLRender struct {
	//需要修改的值就用指针
	Template *template.Template
}
