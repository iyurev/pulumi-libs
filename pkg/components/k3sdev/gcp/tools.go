package gcp

import (
	"bytes"
	"strings"
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

func CutOutDot(s string) string {
	return strings.TrimSuffix(s, ".")
}
