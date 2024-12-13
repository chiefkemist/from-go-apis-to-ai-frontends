import "strings"

// Image upload contract
#ImageUpload: {
	// Unique identifier
	id: string & =~"^[0-9a-zA-Z -]{36}$"

	// Image title
	title: string & =~"^.{3,100}$" & =~"^[A-Za-z0-9 -_.]+$"

	// Base64 encoded image
	blob: string & strings.MinRunes(3) & strings.MaxRunes(13_900_000) & =~"^data:image/(jpeg|png|gif|webp);base64,[A-Za-z0-9+/]+=*$"
}

// Image upload status
#ImageUploadStatus: {
	// Unique identifier
	id: string & =~"^[0-9a-zA-Z -]{36}$"

	// Image title
	title: string & strings.MinRunes(3) & strings.MaxRunes(100) & =~"^[A-Za-z0-9 -_.]+$"

	// Image upload status
	status: string & strings.MinRunes(3) & strings.MaxRunes(300) & =~"^[A-Za-z0-9 -_.]+$"
}
