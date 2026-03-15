package attachment

type Attachment struct {
	ID           int64
	PublicID     string
	OriginalName string
	StoredName   string
	FileExt      string
	ContentType  string
	FileSize     int64
	SHA256       string
	SourceURL    *string
	Note         *string
	CreatedAt    int64
	UpdatedAt    int64
}

type CreateInput struct {
	OriginalName string
	StoredName   string
	FileExt      string
	ContentType  string
	FileSize     int64
	SHA256       string
	SourceURL    *string
	Note         *string
}

type MetadataPatch struct {
	URL  *string
	Note *string
}
