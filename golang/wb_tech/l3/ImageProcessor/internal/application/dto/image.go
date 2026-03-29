package dto

type UploadImageInput struct {
	Filename string
	Data     []byte
}

type ImageJob struct {
	ImageID string `json:"image_id"`
}

type ProcessedImage struct {
	Processed []byte
	Thumbnail []byte
	Format    string
}
