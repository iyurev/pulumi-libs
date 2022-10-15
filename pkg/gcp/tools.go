package gcp

import (
	"bytes"
	"text/template"
)

func IsEmptyStr(s string) bool {
	return s == ""
}

func RenderingTmpl(tmpl string, data interface{}) (string, error) {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var textBuff bytes.Buffer
	if err := t.Execute(&textBuff, data); err != nil {
		return "", err
	}
	return textBuff.String(), nil
}
