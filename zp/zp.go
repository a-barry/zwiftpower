package zp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type club struct {
	Riders []Rider `json:"data"`
}

// Rider shows data about a rider
type Rider struct {
	Name   string
	Zwid   int
	Weight interface{} `json:"w"`
	Div    int         `json:"div"`  //ZP cat 5 = A+, 10 = A, 20 = B, 30 = C, 40 = D
	DivW   int         `json:"divw"` //ZP womens car 5 = A+, 10 = A, 20 = B, 30 = C, 40 = D
}

// Rider shows data about a rider
type RiderDetail struct {
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

	Wkg20min30Days float64
	Wkg5min30Days  float64
	Wkg2min30Days  float64
	Wkg1min30Days  float64
	Wkg30sec30Days float64
	Wkg15sec30Days float64
	Wkg5sec30Days  float64

	W20min30Days float64
	W5min30Days  float64
	W2min30Days  float64
	W1min30Days  float64
	W30sec30Days float64
	W15sec30Days float64
	W5sec30Days  float64

	Weight float64

	Div  int //ZP cat 5 = A+, 10 = A, 20 = B, 30 = C, 40 = D
	DivW int //ZP womens car 5 = A+, 10 = A, 20 = B, 30 = C, 40 = D

}

type riderEvents struct {
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
	Weight        interface{} `json:"weight"`
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

var (
	CloudFrontPolicy    string
	CloudFrontSignature string
	CloudFrontKeyPairId string
)

const zpTeamURL = "https://zwiftpower.com/cache3/teams/%d_riders.json"
const zpRiderURL = "https://zwiftpower.com/cache3/profile/%d_all.json"

func newClient() (*http.Client, error) {
	log.Printf("NewClient")
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	/*phpbb3_lswlk_u=57526;
	phpbb3_lswlk_k=ec01eab84919a3dd;
	phpbb3_lswlk_sid=cdeaf271fcffbdf0be7499e33193d133*/

	//jar.SetCookies()
	// req.AddCookie(&http.Cookie{Name: "CloudFront-Policy", Value: CloudFrontPolicy})
	// req.AddCookie(&http.Cookie{Name: "CloudFront-Signature", Value: CloudFrontSignature})
	// req.AddCookie(&http.Cookie{Name: "CloudFront-Key-Pair-Id", Value: CloudFrontKeyPairId})

	var cookies []*http.Cookie

	cookies = append(cookies, &http.Cookie{
		Name:   "CloudFront-Policy",
		Value:  CloudFrontPolicy,
		Path:   "/",
		Domain: "zwiftpower.com",
	})

	cookies = append(cookies, &http.Cookie{
		Name:   "CloudFront-Signature",
		Value:  CloudFrontSignature,
		Path:   "/",
		Domain: "zwiftpower.com",
	})

	cookies = append(cookies, &http.Cookie{
		Name:   "CloudFront-Key-Pair-Id",
		Value:  CloudFrontKeyPairId,
		Path:   "/",
		Domain: "zwiftpower.com",
	})

	u, _ := url.Parse("https://zwiftpower.com")
	jar.SetCookies(u, cookies)

	client := &http.Client{
		Jar: jar,
	}

	return client, nil
}

// ImportTeam imports data about the team with this ID
func ImportTeam(clubID int, limit int) ([]RiderDetail, error) {
	client, err := newClient()
	if err != nil {
		return nil, fmt.Errorf("error getting client: %v", err)
	}

	data, err := getJSON(client, fmt.Sprintf(zpTeamURL, clubID))
	if err != nil {
		return nil, fmt.Errorf("getting club data: %v", err)
	}

	var c club
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling club data: %v", err)
	}

	//return c.Data, nil
	output := make([]RiderDetail, len(c.Riders))

	for i, rider := range c.Riders {
		var err error

		riderDetail, err := importRider(client, rider)
		if err != nil {
			return nil, fmt.Errorf("loading data for %s (%d): %v", rider.Name, rider.Zwid, err)
		}
		output[i] = riderDetail

		if limit > 0 && i >= (limit-1) {
			log.Printf("Limiting output to %d riders", limit)
			break
		}
	}

	return output, nil
}
func ImportRider(riderID int) (riderDetail RiderDetail, err error) {
	var rider Rider
	rider.Zwid = riderID

	client, err := newClient()
	if err != nil {
		log.Printf("error getting client: %v", err)
		return riderDetail, err
	}

	return importRider(client, rider)
}

// ImportRider imports data about the rider with this ID
func importRider(client *http.Client, rider Rider) (riderDetail RiderDetail, err error) {
	// I think hitting the profile URL loads the data into the cache
	log.Printf("ImportRider(%d)", rider.Zwid)

	riderDetail.Zwid = rider.Zwid
	riderDetail.Name = strings.TrimSpace(rider.Name)
	riderDetail.Div = rider.Div
	riderDetail.DivW = rider.DivW
	riderDetail.Weight = convertPowerValue(rider.Weight)

	data, err := getJSON(client, fmt.Sprintf(zpRiderURL, rider.Zwid))
	if err != nil {
		log.Printf("loading data for %s (%d): %v", rider.Name, rider.Zwid, err)
		return riderDetail, err
	}

	var r riderEvents
	err = json.Unmarshal(data, &r)
	if err != nil {
		log.Printf("Error unmarshalling data: %v", err)
		log.Printf(string(data))
		return riderDetail, err
	}

	riderDetail.Zwid = rider.Zwid
	if len(r.Data) < 1 {
		log.Printf("No event data for rider %d", rider.Zwid)
		return riderDetail, nil
	}

	var latestEventDate time.Time
	var latestRaceDate time.Time
	for _, e := range r.Data {
		e.EventDate = time.Unix(int64(e.EventDateSecs), 0)
		daysAgo := int(time.Now().Sub(e.EventDate).Hours() / 24)
		// log.Printf("date %v, from %v is %d days ago\n", e.EventDate, e.EventDateSecs, daysAgo)
		isRace := strings.Contains(e.EventType, "RACE")

		if daysAgo <= 365 {
			riderDetail.Rides++
			if isRace {
				riderDetail.Races++
			}
		}

		var eventWkgFtp float64
		var eventAvgWkg float64

		eventWkgFtp = convertPowerValue(e.WkgFtp)
		eventAvgWkg = convertPowerValue(e.AvgWkg)

		//eventWkgFtp := e.WkgFtp.([]interface{})
		//wkgFtp, ok := eventWkgFtp[0].(float64)
		// if !ok {
		// 	wkgFtp, err = strconv.ParseFloat(eventWkgFtp[0].(string), 64)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// }

		// avgWkg, err = strconv.ParseFloat(e.AvgWkg.([]interface{})[0].(string), 64)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// Last three months?
		if daysAgo <= 90 {
			// if eventWkgFtp > riderDetail.Ftp90Days {
			// 	riderDetail.Ftp90Days = eventWkgFtp
			// }

			replaceIfGreater(&riderDetail.Ftp90Days, eventWkgFtp)

			if isRace {
				riderDetail.Races90Days++
			}
		}

		// Last two months?
		if daysAgo <= 60 {
			replaceIfGreater(&riderDetail.Ftp60Days, eventWkgFtp)
			// if eventWkgFtp > riderDetail.Ftp60Days {
			// 	riderDetail.Ftp60Days = eventWkgFtp
			// }
		}

		// Last month?
		if daysAgo <= 30 {
			if isRace {
				riderDetail.Races30Days++
			}
			replaceIfGreater(&riderDetail.Ftp30Days, eventWkgFtp)
			// if eventWkgFtp > riderDetail.Ftp30Days {
			// 	riderDetail.Ftp30Days = eventWkgFtp
			// }

			// 20 min power
			replaceIfGreater(&riderDetail.W20min30Days, convertPowerValue(e.W20min))
			replaceIfGreater(&riderDetail.Wkg20min30Days, convertPowerValue(e.Wkg20min))

			// 5 minute power
			replaceIfGreater(&riderDetail.W5min30Days, convertPowerValue(e.W5min))
			replaceIfGreater(&riderDetail.Wkg5min30Days, convertPowerValue(e.Wkg5min))

			// 2 minute power
			replaceIfGreater(&riderDetail.W2min30Days, convertPowerValue(e.W2min))
			replaceIfGreater(&riderDetail.Wkg2min30Days, convertPowerValue(e.Wkg2min))

			// 1 minute power
			replaceIfGreater(&riderDetail.W1min30Days, convertPowerValue(e.W1min))
			replaceIfGreater(&riderDetail.Wkg1min30Days, convertPowerValue(e.Wkg1min))

			// 30 second power
			replaceIfGreater(&riderDetail.W30sec30Days, convertPowerValue(e.W30sec))
			replaceIfGreater(&riderDetail.Wkg30sec30Days, convertPowerValue(e.Wkg30sec))

			// 15 second power
			replaceIfGreater(&riderDetail.W15sec30Days, convertPowerValue(e.W15sec))
			replaceIfGreater(&riderDetail.Wkg15sec30Days, convertPowerValue(e.Wkg15sec))

			// 5 second power
			replaceIfGreater(&riderDetail.W5sec30Days, convertPowerValue(e.W5sec))
			replaceIfGreater(&riderDetail.Wkg5sec30Days, convertPowerValue(e.Wkg5sec))
		}

		if e.EventDate.After(latestEventDate) {
			latestEventDate = e.EventDate
			riderDetail.LatestEvent = e.EventTitle
		}

		if isRace && e.EventDate.After(latestRaceDate) {
			latestRaceDate = e.EventDate
			riderDetail.LatestRace = e.EventTitle
			riderDetail.LatestRaceAvgWkg = eventAvgWkg
			riderDetail.LatestRaceWkgFtp = eventWkgFtp
		}
	}

	riderDetail.LatestEventDate = latestEventDate
	riderDetail.LatestRaceDate = latestRaceDate
	return riderDetail, nil
}

func convertPowerValue(sourceVal interface{}) float64 {
	sourceValArr := sourceVal.([]interface{})
	pwrVal, ok := sourceValArr[0].(float64) // try to convert as float64
	if !ok {
		// if float 64 fails then try as string
		var err error
		pwrVal, err = strconv.ParseFloat(sourceValArr[0].(string), 64)
		if err != nil {
			//log.Fatal(err)
		}
	}
	return pwrVal
}

func replaceIfGreater(current *float64, newVal float64) {
	if newVal > *current {
		*current = newVal
	}
}

func getJSON(client *http.Client, url string) ([]byte, error) {
	//resp, err := client.Get(url)
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return []byte{}, err
	}

	// req.AddCookie(&http.Cookie{Name: "CloudFront-Policy", Value: CloudFrontPolicy})
	// req.AddCookie(&http.Cookie{Name: "CloudFront-Signature", Value: CloudFrontSignature})
	// req.AddCookie(&http.Cookie{Name: "CloudFront-Key-Pair-Id", Value: CloudFrontKeyPairId})

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

// // MonthsAgo describes how many months since the rider's latest event
// func (r Rider) MonthsAgo() string {
// 	if r.LatestEventDate.IsZero() {
// 		return "No latest event"
// 	}

// 	if time.Now().Sub(r.LatestEventDate) > (time.Hour * 24 * 365) {
// 		return "Over a year ago"
// 	}

// 	monthDiff := time.Now().Month() - r.LatestEventDate.Month()
// 	if monthDiff < 0 {
// 		monthDiff += 12
// 	}

// 	switch monthDiff {
// 	case 0:
// 		return "This month"
// 	case 1:
// 		return "Last month"
// 	default:
// 		return fmt.Sprintf("%d months ago", monthDiff)
// 	}
// }

// Strings turns a rider struct into []string headers
func ColumnHeaders() []string {
	output := make([]string, 25)
	output[0] = "Name"
	output[1] = "Zwid"
	output[2] = "Profile"
	output[3] = "Category"
	output[4] = "Womens Category"
	output[5] = "Ftp30Days"
	output[6] = "Ftp60Days"
	output[7] = "Ftp90Days"
	output[8] = "Races30Days"
	output[9] = "Races90Days"

	output[10] = "W20min30Days"
	output[11] = "Wkg20min30Days"

	output[12] = "W5min30Days"
	output[13] = "Wkg5min30Days"

	output[14] = "W2min30Days"
	output[15] = "Wkg2min30Days"

	output[16] = "W1min30Days"
	output[17] = "Wkg1min30Days"

	output[18] = "W30sec30Days"
	output[19] = "Wkg30sec30Days"

	output[20] = "W15sec30Days"
	output[21] = "Wkg15sec30Days"

	output[22] = "W5sec30Days"
	output[23] = "Wkg5sec30Days"

	output[24] = "Weight"

	return output
}

func catValToString(cat int) (catString string) {
	switch cat {
	case 5:
		catString = "A+"
	case 10:
		catString = "A"
	case 20:
		catString = "B"
	case 30:
		catString = "C"
	case 40:
		catString = "D"
	default:
		catString = ""
	}
	return
}

// Strings turns a rider struct into []string
func (r RiderDetail) Strings() []string {
	output := make([]string, 25)
	output[0] = r.Name
	output[1] = strconv.Itoa(r.Zwid)
	output[2] = fmt.Sprintf("https://zwiftpower.com/profile.php?z=%d", r.Zwid)

	output[3] = catValToString(r.Div)
	output[4] = catValToString(r.DivW)

	output[5] = strconv.FormatFloat(r.Ftp30Days, 'f', 1, 64)
	output[6] = strconv.FormatFloat(r.Ftp60Days, 'f', 1, 64)
	output[7] = strconv.FormatFloat(r.Ftp90Days, 'f', 1, 64)
	output[8] = strconv.Itoa(r.Races30Days)
	output[9] = strconv.Itoa(r.Races90Days)

	output[10] = strconv.Itoa(int(r.W20min30Days))
	output[11] = strconv.FormatFloat(r.Wkg20min30Days, 'f', 1, 64)

	output[12] = strconv.Itoa(int(r.W5min30Days))
	output[13] = strconv.FormatFloat(r.Wkg5min30Days, 'f', 1, 64)

	output[14] = strconv.Itoa(int(r.W2min30Days))
	output[15] = strconv.FormatFloat(r.Wkg2min30Days, 'f', 1, 64)

	output[16] = strconv.Itoa(int(r.W1min30Days))
	output[17] = strconv.FormatFloat(r.Wkg1min30Days, 'f', 1, 64)

	output[18] = strconv.Itoa(int(r.W30sec30Days))
	output[19] = strconv.FormatFloat(r.Wkg30sec30Days, 'f', 1, 64)

	output[20] = strconv.Itoa(int(r.W15sec30Days))
	output[21] = strconv.FormatFloat(r.Wkg15sec30Days, 'f', 1, 64)

	output[22] = strconv.Itoa(int(r.W5sec30Days))
	output[23] = strconv.FormatFloat(r.Wkg5sec30Days, 'f', 1, 64)

	output[24] = strconv.FormatFloat(r.Weight, 'f', 1, 64)
	//output[2] = r.LatestEventDate.Format("2006-01-02")
	//output[3] = r.MonthsAgo()
	//output[4] = r.LatestEvent
	//output[5] = strconv.Itoa(r.Rides)
	//output[6] = fmt.Sprintf("https://www.zwiftpower.com/profile.php?z=%d", r.Zwid)
	//output[7] = strconv.FormatFloat(r.Ftp30Days, 'f', 1, 64)
	//output[8] = strconv.FormatFloat(r.Ftp90Days, 'f', 1, 64)
	//output[9] = strconv.Itoa(r.Races30Days)
	//output[10] = strconv.Itoa(r.Races90Days)
	//output[11] = strconv.Itoa(r.Races)
	//output[12] = r.LatestRace
	//output[13] = r.LatestRaceDate.Format("2006-01-02")
	return output
}
