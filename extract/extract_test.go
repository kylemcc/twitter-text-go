package extract

import (
	"os"
	"path"
)

type Conformance struct {
	Tests map[string][]*Test
}

type Test struct {
	Description string
	Text        string
	Expected    interface{}
}

var cwd, _ = os.Getwd()
var parentDir = path.Dir(cwd)
var extractYmlPath = path.Join(parentDir, "twitter-text-conformance", "extract.yml")
