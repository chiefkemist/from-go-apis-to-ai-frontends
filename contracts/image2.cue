import "strings"

// Image upload contract
#ImageUpload: {
	// Unique identifier
	id: string & =~"^[0-9a-zA-Z -]{36}$"

	// Image prompt
	prompt: string & =~"^.{3,100}$" & =~"^[A-Za-z0-9 -_.]+$"

	// Stream enabled
	stream: bool

	// Base64 encoded image
	blob: string & strings.MinRunes(3) & strings.MaxRunes(13_900_000) & =~"^data:image/(jpeg|png|gif|webp);base64,[A-Za-z0-9+/]+=*$"
}

// Image info contract
#ImageInfo: {
	// Image info
	info: string
}
