package pictures

import (
	"errors"
	"fmt"
	"time"

	"github.com/opesun/goquery"
)

type Pictures struct {
	Items         []string
	ExpiresAt     time.Time
	pictureNumber int
	nextPageUrl   string
}

const firstPageUrl = "/tag/%23%D0%9F%D1%80%D0%B8%D0%BA%D0%BE%D0%BB%D1%8B+%D0%B4%D0%BB%D1%8F+%D0%B4%D0%B0%D1%83%D0%BD%D0%BE%D0%B2"
const domain = "http://joyreactor.cc"

func (p *Pictures) SetExpiresIn(t time.Duration) {
	p.ExpiresAt = time.Now().Add(t)
}

func (p *Pictures) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

func (p *Pictures) Update() {
	fmt.Println("Pictures Updating...")
	x, err := goquery.ParseUrl(domain + firstPageUrl)
	if err != nil {
		panic(err)
	}
	p.Items = x.Find("#post_list .postContainer .article div.post_top div.post_content div.image img").Attrs("src")
	p.nextPageUrl = x.Find("#Pagination .pagination_main a").Attr("href")
	p.ExpiresAt = time.Now().Add(5 * time.Hour)
	p.pictureNumber = 0

	if len(p.Items) != 0 {
		fmt.Println("Pictures Updated successfully")
	}
}

// GetPicture represents new picture url, and initiate NextPage update if all pictures from current Items slice is taken
func (p *Pictures) GetPicture() (string, error) {
	fmt.Println("Getting picture...")

	if len(p.Items) == 0 || p.IsExpired() {
		p.Update()
		if len(p.Items) == 0 {
			return "", errors.New("Нет картинок почему то :/")
		}
	}

	if len(p.Items)-1 == p.pictureNumber {
		p.NextPage()
	}

	pic := p.Items[p.pictureNumber]
	p.pictureNumber++

	return pic, nil
}

// NextPage request new Items and change nextPageUrl as well
func (p *Pictures) NextPage() {
	fmt.Println("Getting new page...", len(p.Items))
	x, err := goquery.ParseUrl(domain + p.nextPageUrl)
	if err != nil {
		panic(err)
	}

	p.Items = append(p.Items, x.Find("#post_list .postContainer .article div.post_top div.post_content div.image img").Attrs("src")...)
	p.nextPageUrl = x.Find("#Pagination .pagination_main a").Attrs("href")[1]

	fmt.Println("Successfully got new page", len(p.Items))
	// fmt.Println("nextPageUrl", p.nextPageUrl)
}
