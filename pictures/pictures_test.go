package pictures

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestGetPictures(t *testing.T) {
	pics := Pictures{Logger: &zerolog.Logger{}}

	id := 1

	pic, err := pics.GetPicture(id)
	if err != nil {
		t.Error("TestGetPictures - GetPicture returns error:", err)
	}

	if len(pic) == 0 {
		t.Error("TestGetPictures - GetPicture returns empty string:")
	}
}
