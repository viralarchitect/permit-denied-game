package assets

import "embed"

//go:embed usable/tileset.png usable/tileset.tsj usable/sprites.png usable/sprites.json usable/sfx/*.wav
var FS embed.FS
