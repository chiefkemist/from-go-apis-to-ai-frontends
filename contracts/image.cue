// Image upload contract
#ImageUpload: {
    // Unique identifier
    id:    string & =~"^[0-9a-fA-F]{24}$"

    // Image title
    title: string & {
        minLength: 3
        maxLength: 100
        =~"^[A-Za-z0-9 -_.]+$"
    }

    // Base64 encoded image
    image: string & {
        // Must be base64 encoded
        =~"^data:image/(jpeg|png|gif);base64,[A-Za-z0-9+/]+=*$"

        // Reasonable size limit (10MB in base64)
        maxLength: 13_900_000
    }
}
