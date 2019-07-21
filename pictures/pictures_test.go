package pictures

import (
	"testing"
	"time"
	"unicode/utf8"

	"github.com/rs/zerolog"
)

func SliceUniqMap(s []string) []string {
	seen := make(map[string]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}

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

	list := make([]string, 0)

	for i := 0; i < 100; i++ {
		pic, err := pics.GetPicture(id)
		if err != nil {
			t.Error("TestGetPictures - GetPicture cycle returns error:", err)
		}

		if utf8.RuneCountInString(pic) == 0 {
			t.Error("TestGetPictures - GetPicture returns empty url", pic)
		}

		list = append(list, pic)
	}

	uniqList := SliceUniqMap(list)

	if len(uniqList) != len(list) {
		t.Errorf("TestGetPictures - GetPicture returns duplicates expected: %v, recieved: %v", len(uniqList), len(list))
	}
}

func TestGetUrlHistory(t *testing.T) {
	pics := Pictures{Logger: &zerolog.Logger{}}

	id := 1

	pic, err := pics.GetPicture(id)

	history := pics.GetUrlHistory()

	if err != nil {
		t.Error("TestGetHistory - GetPicture returns error:", err)
	}

	if len(history) != 1 {
		t.Errorf("TestGetHistory - GetHistory returns incorect history length: returns - [%v] expected - [%v]", len(history), 1)
	}

	if _, ok := history[id]; !ok {
		t.Error("TestGetHistory - GetHistory doesn't contain necessary item")
	}

	if value, ok := history[id]; ok {
		if !contains(value, pic) {
			t.Error("TestGetHistory - History doesn't contain previous url")
		}
	}

	if value, ok := history[2]; ok {
		t.Errorf("TestGetHistory - GetHistory does contain necessary item: [%v] expected: not exist", value)
	}

	id2 := 2

	pic2, err2 := pics.GetPicture(id2)

	if err2 != nil {
		t.Error("TestGetHistory - err in GetPicture")
		panic(err2)
	}

	history = pics.GetUrlHistory()

	if len(history) != 2 {
		t.Errorf("TestGetHistory - GetHistory after second request returns incorect history length: returns - [%v] expected - [%v]", len(history), 2)
	}

	_, err3 := pics.GetPicture(id)
	if err3 != nil {
		t.Error("TestGetHistory - err in GetPicture")
		panic(err3)
	}

	if _, ok := history[id2]; !ok {
		t.Error("TestGetHistory - History for second user doesn't exist")
	}

	if historyItem, ok := history[id2]; ok {
		if !contains(historyItem, pic2) {
			t.Error("TestGetHistory - History for second user doesn't contain correct url")
		}
	}
}

func TestPicturesOtherFuncs(t *testing.T) {
	pics := Pictures{Logger: &zerolog.Logger{}}
	id := 1

	_, err := pics.GetPicture(id)
	if err != nil {
		t.Error("TestPicturesOtherFuncs - GetPicture returns error:", err)
		panic(err)
	}

	if pics.IsExpired() {
		t.Errorf("TestPicturesOtherFuncs - IsExpired returns incorrect value: expected: false, returns: %v", pics.IsExpired())
	}

	pics.SetExpiresIn(1 * time.Second)
	time.Sleep(3 * time.Second)

	if !pics.IsExpired() {
		t.Errorf("TestPicturesOtherFuncs - IsExpired returns incorrect value: expected: true, returns: %v", pics.IsExpired())
	}
}
