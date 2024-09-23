package assets

import "embed"

//go:embed templates
var Templates embed.FS

//go:embed static
var Static embed.FS

//go:embed survey.yaml
var SurveyYAML []byte
