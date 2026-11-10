package media

import "github.com/cloudinary/cloudinary-go/v2/transformation"

func transformationOf(raw string) transformation.RawTransformation {
	return transformation.RawTransformation(raw)
}
