//go:build embed_main
// +build embed_main

package runner

import (
	"embed"
)

const Option_Embed_Main bool = true

//go:embed buildtemp/main.rye
var Rye_files embed.FS
