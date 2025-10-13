package email

import "gopkg.cc/apibase/errx"

var (
	ErrInvalidConfig           = errx.NewType("invalid email sender config")
	ErrInvalidTemplate         = errx.NewType("invalide email template")
	ErrTemplateExec            = errx.NewType("error during template execution")
	ErrTemplateNoFileInEmbedFS = errx.NewType("template file not found in embedfs")
)
