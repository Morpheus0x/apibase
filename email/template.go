package email

import (
	"bytes"
	"embed"
	"html/template"
	"os"

	"gopkg.cc/apibase/errx"
)

type EmailTemplate struct {
	FilePath string `toml:"file"`
}

func (c EmailConfig) SendTemplate(to []string, tmpl EmailTemplate, data any) error {
	if c.From == "" {
		return errx.NewWithType(ErrInvalidConfig, "from email address not specified in config")
	}
	return c.SendTemplateFrom(c.From, to, tmpl, data)
}

func (c EmailConfig) SendTemplateFrom(from string, to []string, tmpl EmailTemplate, data any) error {
	if stat, err := os.Stat(tmpl.FilePath); err != nil || stat.IsDir() {
		return errx.WrapWithTypef(ErrInvalidTemplate, err, "unable to get file '%s'", tmpl.FilePath)

	}
	t, err := template.New("email").ParseFiles(tmpl.FilePath)
	if err != nil {
		return errx.WrapWithTypef(ErrInvalidTemplate, err, "unable to parse template from file '%s'", tmpl.FilePath)
	}
	return c.executeAndSend(t, from, to, data)
}

func (c EmailConfig) SendTemplateEmbedFs(embedFS embed.FS, to []string, tmpl EmailTemplate, data any) error {
	if c.From == "" {
		return errx.NewWithType(ErrInvalidConfig, "from email address not specified in config")
	}
	return c.SendTemplateEmbedFsFrom(embedFS, c.From, to, tmpl, data)
}

func (c EmailConfig) SendTemplateEmbedFsFrom(embedFS embed.FS, from string, to []string, tmpl EmailTemplate, data any) error {
	t, err := template.ParseFS(embedFS, tmpl.FilePath)
	if err != nil {
		return errx.WrapWithType(ErrTemplateNoFileInEmbedFS, err, "")
	}
	return c.executeAndSend(t, from, to, data)
}

func (c EmailConfig) executeAndSend(t *template.Template, from string, to []string, data any) error {
	var msg bytes.Buffer
	err := t.Execute(&msg, data)
	if err != nil {
		return errx.WrapWithType(ErrTemplateExec, err, "")
	}

	return c.SendMailFrom(from, to, msg.Bytes())
}
