package pictures

import (
	"errors"
	"fmt"
	"time"

	"github.com/opesun/goquery"
)

//Pictures provide methods for get pictures from url, changing pages and save seen history for each id
type Pictures struct {
	Items       []string
	ExpiresAt   time.Time
	nextPageURL string
	history     map[int]int
}

const firstPageURL = "/tag/%23%D0%9F%D1%80%D0%B8%D0%BA%D0%BE%D0%BB%D1%8B+%D0%B4%D0%BB%D1%8F+%D0%B4%D0%B0%D1%83%D0%BD%D0%BE%D0%B2"
const domain = "http://joyreactor.cc"

//SetExpiresIn provide method for update Pictures expiresAt time manually
func (p *Pictures) SetExpiresIn(t time.Duration) {
	p.ExpiresAt = time.Now().Add(t)
}

//IsExpired provide method for checkeking expired set of items or not
func (p *Pictures) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

//Update initiate updating Items
func (p *Pictures) Update() {
	fmt.Println("Pictures Updating...")
	x, err := goquery.ParseUrl(domain + firstPageURL)
	if err != nil {
		panic(err)
	}
	p.Items = x.Find("#post_list .postContainer .article div.post_top div.post_content div.image img").Attrs("src")
	p.nextPageURL = x.Find("#Pagination .pagination_main a").Attr("href")
	p.ExpiresAt = time.Now().Add(5 * time.Hour)
	p.history = map[int]int{}

	if len(p.Items) != 0 {
		fmt.Println("Pictures Updated successfully")
	}
}

// GetPicture represents new picture url, and initiate NextPage update if all pictures from current Items slice is taken
func (p *Pictures) GetPicture(id int) (string, error) {
	fmt.Println("Getting picture... forId: ", id)

	if len(p.Items) == 0 || p.IsExpired() {
		p.Update()
		if len(p.Items) == 0 {
			return "", errors.New("Нет картинок почему то :/")
		}
	}

	if val, ok := p.history[id]; ok {
		fmt.Println("This id already stored", id, val)
		if len(p.Items)-1 == val {
			p.NextPage()
		}

		p.history[id] = val + 1
		return p.Items[val+1], nil
	}

	fmt.Println("This id is new", id)
	p.history[id] = 0
	return p.Items[0], nil
}

//GetHistory provide history map
func (p *Pictures) GetHistory() map[int]int {
	return p.history
}

// NextPage request new Items and change nextPageUrl as well
func (p *Pictures) NextPage() {
	fmt.Println("Getting new page...", len(p.Items))
	x, err := goquery.ParseUrl(domain + p.nextPageURL)
	if err != nil {
		panic(err)
	}

	p.Items = append(p.Items, x.Find("#post_list .postContainer .article div.post_top div.post_content div.image img").Attrs("src")...)
	p.nextPageURL = x.Find("#Pagination .pagination_main a").Attrs("href")[1]

	fmt.Println("Successfully got new page", len(p.Items))
	// fmt.Println("nextPageUrl", p.nextPageUrl)
}
