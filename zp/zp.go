package zp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

type club struct {
	Data []Rider
}

// Rider shows data about a rider
type Rider struct {
	Name             string
	Zwid             int
	LatestEventDate  time.Time
	Rides            int
	Races            int
	Races90Days      int
	Races30Days      int
	Ftp90Days        float64
	Ftp60Days        float64
	Ftp30Days        float64
	LatestRace       string
	LatestRaceDate   time.Time
	LatestEvent      string
	LatestRaceAvgWkg float64
	LatestRaceWkgFtp float64
}

type riderData struct {
	Data []Event
}

// Event is a ZwiftPower event
type Event struct {
	EventType     string        `json:"f_t"`
	EventDateSecs EventDateType `json:"event_date"`
	EventDate     time.Time
	EventTitle    string      `json:"event_title"`
	AvgWkg        interface{} `json:"avg_wkg"`
	WkgFtp        interface{} `json:"wkg_ftp"`
	Wkg20min      interface{} `json:"wkg1200"`
	Wkg5min       interface{} `json:"wkg300"`
	Wkg2min       interface{} `json:"wkg120"`
	Wkg1min       interface{} `json:"wkg60"`
	Wkg30sec      interface{} `json:"wkg30"`
	Wkg15sec      interface{} `json:"wkg15"`
	Wkg5sec       interface{} `json:"wkg5"`
	WFtp          interface{} `json:"wftp"`
	W20min        interface{} `json:"w1200"`
	W5min         interface{} `json:"w300"`
	W2min         interface{} `json:"w120"`
	W1min         interface{} `json:"w60"`
	W30sec        interface{} `json:"w30"`
	W15sec        interface{} `json:"w15"`
	W5sec         interface{} `json:"w5"`
	Weight        interface{} `json:"weight[0]"`
}

// EventDateType so we can use a custom unmarshaller
type EventDateType int64

// UnmarshalJSON custom because usually EventDateSecs is a number, but sometimes it's an empty string
func (e *EventDateType) UnmarshalJSON(data []byte) error {
	var v int64

	// recklessly ignoring the error, because we'll get an error if the JSON is an empty string
	// and in that case we want to return 0
	json.Unmarshal(data, &v)
	*e = EventDateType(v)
	return nil
}

func NewClient() (*http.Client, error) {
	log.Printf("NewClient")
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	/*phpbb3_lswlk_u=57526;
	phpbb3_lswlk_k=ec01eab84919a3dd;
	phpbb3_lswlk_sid=cdeaf271fcffbdf0be7499e33193d133*/

	//jar.SetCookies()

	client := &http.Client{
		Jar: jar,
	}

	return client, nil
}

// ImportZP imports data about the club with this ID
func ImportZP(client *http.Client, clubID int) ([]Rider, error) {
	//	https://zwiftpower.com/api3.php?do=team_riders&id=%d
	//https://www.zwiftpower.com/cache3/teams/%d_riders.json
	data, err := getJSON(client, fmt.Sprintf("https://zwiftpower.com/cache3/teams/%d_riders.json", clubID))
	if err != nil {
		return nil, fmt.Errorf("getting club data: %v", err)
	}

	var c club
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling club data: %v", err)
	}

	return c.Data, nil
}

// ImportRider imports data about the rider with this ID
func ImportRider(client *http.Client, riderID int) (rider Rider, err error) {
	// I think hitting the profile URL loads the data into the cache
	log.Printf("ImportRider(%d)", riderID)
	//_, _ = client.Get(fmt.Sprintf("https://www.zwiftpower.com/profile.php?z=%d", riderID))
	//https://www.zwiftpower.com/cache3/profile/%d_all.json
	data, err := getJSON(client, fmt.Sprintf("https://zwiftpower.com/cache3/profile/%d_all.json", riderID))
	if err != nil {
		return rider, err
	}

	var r riderData
	err = json.Unmarshal(data, &r)
	if err != nil {
		log.Printf("Error unmarshalling data: %v", err)
		log.Printf(string(data))
		return rider, err
	}

	rider.Zwid = riderID
	if len(r.Data) < 1 {
		log.Printf("No event data for rider %d", riderID)
		return rider, nil
	}

	var latestEventDate time.Time
	var latestRaceDate time.Time
	for _, e := range r.Data {
		e.EventDate = time.Unix(int64(e.EventDateSecs), 0)
		daysAgo := int(time.Now().Sub(e.EventDate).Hours() / 24)
		// log.Printf("date %v, from %v is %d days ago\n", e.EventDate, e.EventDateSecs, daysAgo)
		isRace := strings.Contains(e.EventType, "RACE")

		if daysAgo <= 365 {
			rider.Rides++
			if isRace {
				rider.Races++
			}
		}

		var wkgFtp float64
		var avgWkg float64

		eventWkgFtp := e.WkgFtp.([]interface{})
		wkgFtp, ok := eventWkgFtp[0].(float64)
		if !ok {
			wkgFtp, err = strconv.ParseFloat(eventWkgFtp[0].(string), 64)
			if err != nil {
				log.Fatal(err)
			}
		}

		avgWkg, err = strconv.ParseFloat(e.AvgWkg.([]interface{})[0].(string), 64)
		if err != nil {
			log.Fatal(err)
		}

		// Last three months?
		if daysAgo <= 90 {
			if wkgFtp > rider.Ftp90Days {
				rider.Ftp90Days = wkgFtp
			}

			if isRace {
				rider.Races90Days++
			}
		}

		// Last two months?
		if daysAgo <= 60 {
			if wkgFtp > rider.Ftp60Days {
				rider.Ftp60Days = wkgFtp
			}
		}

		// Last month?
		if daysAgo <= 30 {
			if isRace {
				rider.Races30Days++
			}

			if wkgFtp > rider.Ftp30Days {
				rider.Ftp30Days = wkgFtp
			}
		}

		if e.EventDate.After(latestEventDate) {
			latestEventDate = e.EventDate
			rider.LatestEvent = e.EventTitle
		}

		if isRace && e.EventDate.After(latestRaceDate) {
			latestRaceDate = e.EventDate
			rider.LatestRace = e.EventTitle
			rider.LatestRaceAvgWkg = avgWkg
			rider.LatestRaceWkgFtp = wkgFtp
		}
	}

	rider.LatestEventDate = latestEventDate
	rider.LatestRaceDate = latestRaceDate
	return rider, nil
}

func getJSON(client *http.Client, url string) ([]byte, error) {
	//resp, err := client.Get(url)
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return []byte{}, err
	}

	/*phpbb3_lswlk_u=57526;
	phpbb3_lswlk_k=ec01eab84919a3dd;
	phpbb3_lswlk_sid=cdeaf271fcffbdf0be7499e33193d133*/
	//req.AddCookie(&http.Cookie{Name: "phpbb3_lswlk_u", Value: "57526"})
	//req.AddCookie(&http.Cookie{Name: "phpbb3_lswlk_k", Value: "ec01eab84919a3dd"})
	//req.AddCookie(&http.Cookie{Name: "phpbb3_lswlk_sid", Value: "cdeaf271fcffbdf0be7499e33193d133"})

	req.AddCookie(&http.Cookie{Name: "CloudFront-Policy", Value: "eyJTdGF0ZW1lbnQiOlt7IlJlc291cmNlIjoiaHR0cHM6Ly96d2lmdHBvd2VyLmNvbS8qIiwiQ29uZGl0aW9uIjp7IkRhdGVMZXNzVGhhbiI6eyJBV1M6RXBvY2hUaW1lIjoyMTQ3NDgzNjQ3fX19XX0_"})
	req.AddCookie(&http.Cookie{Name: "CloudFront-Signature", Value: "7f6sOCtlQZRe90rpvVq9l84rRNbxl4HBsyEqJByfkMsXWzVe2JkpWGGMZ~KZ1W5w9lm~wYTaQM8ZHFO-MCBokO9EsyeeAcxFOd7gIu5OTcFFFankW4L6RBR-o3oS48dIi22zpWjLZAyku4SXlXbhtMUJrLsASioVG~UDe00Kvc3OWw~WAf0a8LkrdMxG4kIOlb7vvKYEKONo7IO5-JxNVq5WDwMJqB43qRrI2dSK3TMDUsugJHd6QmWpaPxQPHDxrH7MkuBaumdUS1138NbwLL8KwVfcLlV~VnooURL~ID~9lLQHfQ11tkASNzeJgYczkBXTGiKeOh9npbElvqlfOg__"})
	req.AddCookie(&http.Cookie{Name: "CloudFront-Key-Pair-Id", Value: "K2HE75OK1CK137"})

	resp, err := client.Do(req)

	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return []byte{}, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

// MonthsAgo describes how many months since the rider's latest event
func (r Rider) MonthsAgo() string {
	if r.LatestEventDate.IsZero() {
		return "No latest event"
	}

	if time.Now().Sub(r.LatestEventDate) > (time.Hour * 24 * 365) {
		return "Over a year ago"
	}

	monthDiff := time.Now().Month() - r.LatestEventDate.Month()
	if monthDiff < 0 {
		monthDiff += 12
	}

	switch monthDiff {
	case 0:
		return "This month"
	case 1:
		return "Last month"
	default:
		return fmt.Sprintf("%d months ago", monthDiff)
	}
}

// Strings turns a rider struct into []string
func (r Rider) Strings() []string {
	output := make([]string, 14)
	output[0] = r.Name
	output[1] = strconv.Itoa(r.Zwid)
	output[2] = r.LatestEventDate.Format("2006-01-02")
	output[3] = r.MonthsAgo()
	output[4] = r.LatestEvent
	output[5] = strconv.Itoa(r.Rides)
	output[6] = fmt.Sprintf("https://www.zwiftpower.com/profile.php?z=%d", r.Zwid)
	output[7] = strconv.FormatFloat(r.Ftp30Days, 'f', 1, 64)
	output[8] = strconv.FormatFloat(r.Ftp90Days, 'f', 1, 64)
	output[9] = strconv.Itoa(r.Races30Days)
	output[10] = strconv.Itoa(r.Races90Days)
	output[11] = strconv.Itoa(r.Races)
	output[12] = r.LatestRace
	output[13] = r.LatestRaceDate.Format("2006-01-02")
	return output
}
