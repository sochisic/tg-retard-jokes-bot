package pictures

import (
	"errors"
	"time"

	"github.com/opesun/goquery"
	"github.com/rs/zerolog"
)

//Pictures provide methods for get pictures from url, changing pages and save seen history for each id
type Pictures struct {
	Items       []string
	ExpiresAt   time.Time
	nextPageURL string
	urlHistory  map[int][]string
	Logger      *zerolog.Logger
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
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
	p.Logger.Debug().Msg("Pictures Updating...")
	x, err := goquery.ParseUrl(domain + firstPageURL)
	if err != nil {
		panic(err)
	}
	p.Items = x.Find("#post_list .postContainer .article div.post_top div.post_content div.image img").Attrs("src")
	p.nextPageURL = x.Find("#Pagination .pagination_main a").Attr("href")
	p.ExpiresAt = time.Now().Add(1 * time.Hour)
	p.urlHistory = make(map[int][]string, 0)

	if len(p.Items) != 0 {
		p.Logger.Print("Pictures Updated successfully")
	}
}

// GetPicture represents new picture url, and initiate NextPage update if all pictures from current Items slice is taken
func (p *Pictures) GetPicture(id int) (string, error) {
	p.Logger.Debug().Msgf("Getting picture... forId: %v", id)

	if len(p.Items) == 0 || p.IsExpired() {
		p.Update()
		if len(p.Items) == 0 {
			p.Logger.Error().Msg("No pictures after update()")
			return "", errors.New("Нет картинок почему то :/")
		}
	}

	if urlHistory, ok := p.urlHistory[id]; ok {
		p.Logger.Debug().Msgf("Id: %v already stored", id)

		for _, pic := range p.Items {
			if contains(urlHistory, pic) {
				continue
			} else {
				p.urlHistory[id] = append(p.urlHistory[id], pic)
				return pic, nil
			}
		}

		for {
			newPage := p.nextPage()
			for _, pic := range newPage {
				if contains(urlHistory, pic) {
					continue
				} else {
					p.urlHistory[id] = append(p.urlHistory[id], pic)
					return pic, nil
				}
			}
		}
	}

	p.Logger.Debug().Msgf("Id: %v is new - store to history", id)
	p.urlHistory[id] = append(p.urlHistory[id], p.Items[0])
	return p.Items[0], nil
}

//GetHistory provide history map
func (p *Pictures) GetUrlHistory() map[int][]string {
	return p.urlHistory
}

// NextPage request new Items and change nextPageUrl as well
func (p *Pictures) nextPage() (newItems []string) {
	p.Logger.Debug().Int("Getting new page... items len", len(p.Items)).Send()

	x, err := goquery.ParseUrl(domain + p.nextPageURL)
	if err != nil {
		p.Logger.Panic().Err(err).Send()
		panic(err)
	}

	newItems = x.Find("#post_list .postContainer .article div.post_top div.post_content div.image img").Attrs("src")

	p.Items = append(p.Items, newItems...)
	p.nextPageURL = x.Find("#Pagination .pagination_main a").Attrs("href")[1]

	p.Logger.Debug().Int("Successfully got new page... items len", len(p.Items)).Send()

	return
}
