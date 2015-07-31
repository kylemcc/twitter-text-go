package validate

import (
	"os"
	"path"
)

var cwd, _ = os.Getwd()
var parentDir = path.Dir(cwd)
var validateYmlPath = path.Join(parentDir, "twitter-text-mono", "conformance", "validate.yml")
