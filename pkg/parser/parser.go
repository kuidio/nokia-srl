package parser

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	netwv1alpha1 "github.com/kuidio/kuidapps/apis/network/v1alpha1"
)

func New(path string) (*Parser, error) {
	tmpl := template.New("main").Funcs(templateHelperFunctions)
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tmpl") {
			_, err := tmpl.ParseFiles(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Parser{
		tmpl: tmpl,
	}, nil
}

type Parser struct {
	tmpl *template.Template
}

func (r *Parser) Render(ctx context.Context, nd *netwv1alpha1.NetworkDevice, w io.Writer) error {
	if w != nil {
		return r.tmpl.ExecuteTemplate(w, "main.tmpl", nd.Spec)
	}
	return r.tmpl.ExecuteTemplate(os.Stdout, "main.tmpl", nd.Spec)
}

/*
func (r *Parser) processTemplate(templateFile string, templateName string, data interface{}) string {
	p := template.New(templateName)
	p.Funcs(templateHelperFunctions)
	t := template.Must(p.ParseFiles(templateFile))
	buf := new(bytes.Buffer)
	err := t.ExecuteTemplate(buf, templateName, data)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

func processInterface(nodename string, islinterfaces *netwv1alpha1.NetworkDeviceInterface) string {
	templateFile := path.Join("templates", "interface.tmpl")
	return GeneralTemplateProcessing(templateFile, "srlinterface", islinterfaces)
}
*/
