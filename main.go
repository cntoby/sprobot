package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	playerList = make([]*SoccerPlayer, 0)
)

func main() {
	fmt.Println("Fetch Soccer Player List")
	err := fetchPageList("https://sofifa.com/players")
	if err != nil {
		fmt.Println(err.Error())
	}
	total := len(playerList)
	current := 0
	success := 0
	failure := 0
	fmt.Println("Got", len(playerList), "Items")
	completeNotice := make(chan bool)
	completeThreadNotice := make(chan bool)
	thread := 10
	completeThread := 0

	per := total / thread
	for i := 0; i < thread; i++ {
		start := per * i
		end := start + per
		if end > total {
			end = total
		}
		go func(compNotice chan<- bool, compThreadNotice chan<- bool, players []*SoccerPlayer) {
			for _, v := range players {
				doc, err := fetchPage(v.Url)
				if err != nil {
					compNotice <- false
					continue
				}
				if fetchPlayer(doc, v) != nil {
					compNotice <- false
					continue
				}
				compNotice <- true
			}
			completeThreadNotice <- true
		}(completeNotice, completeThreadNotice, playerList[start:end])
	}
	for {
		select {
		case v := <-completeNotice:
			current++
			if v {
				success++
			} else {
				failure++
			}
			fmt.Printf("\rTotal %d, Complete %d Failure %d, Threads Used %d Threads Completed %d", total, current, failure, thread, completeThread)
		case <-completeThreadNotice:
			completeThread++
			fmt.Printf("\rTotal %d, Complete %d Failure %d, Threads Used %d Threads Completed %d", total, current, failure, thread, completeThread)
			if completeThread >= thread {
				goto COMPLETE
			}
		}
	}
COMPLETE:
	d, _ := json.Marshal(playerList)
	fmt.Println("\nComplete")
	ioutil.WriteFile("data.json", d, 0644)
}

// fetchPageList ...
func fetchPageList(listPage string) error {
	for {
		fmt.Printf("Got %d Items, Fetching Page %s\r", len(playerList), listPage)
		p, err := fetchPage(listPage)
		if err != nil {
			return err
		}
		if s, err := fetchPlayerList(p); err != nil {
			return err
		} else {
			playerList = append(playerList, s...)
		}
		np := p.Find(".pagination .page-item a").Last()
		if np.HasClass("disabled") {
			break
		}
		if href, exists := np.Attr("href"); !exists {
			break
		} else {
			listPage = rebuildUrl(p.Url, href).String()
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

// rebuildUrl build a new Url
func rebuildUrl(orig *url.URL, href string) *url.URL {
	newUrl, _ := url.Parse(orig.String())
	if strings.HasPrefix(href, "/") {
		newUrl.Path = href
	} else {
		tu := strings.Split(newUrl.Path, "/")
		tu[len(tu)-1] = href
		newUrl.Path = strings.Join(tu, "/")
	}
	return newUrl
}

// fetchPage fetch pages
func fetchPage(url string) (*goquery.Document, error) {
	return goquery.NewDocument(url)
}

// fetchPlayerList ...
func fetchPlayerList(page *goquery.Document) ([]*SoccerPlayer, error) {
	playerList := make([]*SoccerPlayer, 0)
	page.Find("article #pjax-container table tbody tr").Each(func(no int, item *goquery.Selection) {
		anchorList := item.Find("td div.col-name").First().Find("a")
		if anchorList.Length() <= 0 {
			return
		}
		anchor := anchorList.Next().First()
		href, exists := anchor.Attr("href")
		if !exists {
			return
		}
		href = rebuildUrl(page.Url, href).String()
		name := anchor.Text()
		player := new(SoccerPlayer)
		player.Properties = make([]PlayerPropertyContainer, 0)
		player.Name = name
		player.Url = href
		playerList = append(playerList, player)
	})
	return playerList, nil
}

var nameAgeRegex *regexp.Regexp

// fetchPlayer ...
func fetchPlayer(page *goquery.Document, player *SoccerPlayer) error {
	// meta info
	meta := page.Find("article .player .info .meta span")
	metaStr, _ := meta.Html()
	nameAgeRegex, _ = regexp.Compile("^([^<>]*) <.*> *Age ([0-9]+) \\((.*)\\) ([^ ]+) ([^ ]+)")
	metaArr := nameAgeRegex.FindStringSubmatch(metaStr)
	player.FullName = metaArr[1]
	age, _ := strconv.Atoi(metaArr[2])
	player.Age = uint(age)
	player.Birthday, _ = time.Parse("Jan 2, 2006", metaArr[3])
	player.Height = metaArr[4]
	player.Weight = metaArr[5]

	// stats
	page.Find("article .player .stats .text-center span").Each(func(seq int, s *goquery.Selection) {
		_val, _ := strconv.Atoi(s.Text())
		switch seq {
		case 0:
			player.Overall = _val
		case 1:
			player.Potential = _val
		case 2:
			player.Value = s.Text()
		case 3:
			player.Wage = s.Text()
		}
	})

	// teams
	page.Find("article .player .teams tr td").Each(func(seq int, s *goquery.Selection) {
		reTag, _ := regexp.Compile("<[^<]+?>[^<]*</[^<]+?>")
		switch seq {
		case 1:
			return
		case 0:
			s.Find("ul li").Each(func(i int, li *goquery.Selection) {
				_li, _ := li.Html()
				_t := reTag.ReplaceAllString(_li, "")
				_v, _ := strconv.Atoi(_t)
				switch i {
				case 0:
					player.Foot = strings.Replace(_t, "\n", "", -1)
				case 1:
					player.Reputation = _v
				case 2:
					player.WeakFoot = _v
				case 3:
					player.SkillMoves = _v
				default:
					return
				}
			})
		case 2:
			s.Find("ul li").Each(func(i int, li *goquery.Selection) {
				switch i {
				case 0:
					player.Team = strings.Replace(li.Text(), "\n", "", -1)
				case 2:
					player.TeamPosition = li.Find("span").Text()
				case 3:
					_li, _ := li.Html()
					num, _ := strconv.Atoi(strings.Replace(reTag.ReplaceAllString(_li, ""), "\n", "", -1))
					player.TeamNumber = num
				}
			})
		case 3:
			s.Find("ul li").Each(func(i int, li *goquery.Selection) {
				switch i {
				case 0:
					player.Country = strings.Replace(li.Text(), "\n", "", -1)
				case 2:
					player.CountryPosition = li.Find("span").Text()
				case 3:
					_li, _ := li.Html()
					num, _ := strconv.Atoi(strings.Replace(reTag.ReplaceAllString(_li, ""), "\n", "", -1))
					player.CountryNumber = num
				}
			})
		}
	})

	player.Properties = make([]PlayerPropertyContainer, 0)
	// columns
	columnRe, _ := regexp.Compile("^([0-9]*) *(.+)$")
	page.Find("article .columns .column div").Each(func(seq int, s *goquery.Selection) {
		containers := make([]PlayerPropertyContainer, 0)
		s.Find("h5").Each(func(_seq int, _s *goquery.Selection) {
			containers = append(containers, PlayerPropertyContainer{Name: _s.Text(), Properties: make([]PlayerProperty, 0)})
		})
		s.Find("ul").Each(func(_seq int, _s *goquery.Selection) {
			//containers[_seq]
			_s.Find("li").Each(func(_ int, _li *goquery.Selection) {
				var score int
				iStr := strings.TrimSpace(_li.Text())
				c := columnRe.FindStringSubmatch(iStr)
				name := c[2]
				if c[1] == "" {
					score = -1
				} else {
					score, _ = strconv.Atoi(c[1])
				}
				containers[_seq].Properties = append(containers[_seq].Properties, PlayerProperty{Name: name, Score: score})
			})
		})
		player.Properties = append(player.Properties, containers...)
	})
	return nil
}
