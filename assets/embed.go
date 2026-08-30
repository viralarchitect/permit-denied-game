package assets

import "embed"

//go:embed usable/lot.json usable/tileset.png usable/sprites.png usable/sprites.json usable/sfx/*.wav
var FS embed.FS
