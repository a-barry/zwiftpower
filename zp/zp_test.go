package zp

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRiderMonthsAgoString(t *testing.T) {
	cases := []struct {
		d        time.Time
		expected string
	}{
		{d: time.Now(), expected: "This month"},
		{d: time.Now().Add(-1 * 30 * 24 * time.Hour), expected: "Last month"},
		{d: time.Now().Add(-2 * 30 * 24 * time.Hour), expected: "2 months ago"},
		{d: time.Now().Add(-3 * 30 * 24 * time.Hour), expected: "3 months ago"},
		{d: time.Now().Add(-13 * 30 * 24 * time.Hour), expected: "Over a year ago"},
		{expected: "No latest event"},
	}

	for i, c := range cases {
		r := Rider{
			LatestEventDate: c.d,
		}
		result := r.MonthsAgo()
		if result != c.expected {
			t.Fatalf("Case %d: got %s expected %s", i, result, c.expected)
		}
	}
}

func TestRiderStrings(t *testing.T) {
	ss := strings.Split("Liz Rice,98588,2020-04-15,This month,ZZRC SUB 2.0 Ride,160,https://www.zwiftpower.com/profile.php?z=98588,2.5,2.7,0,3,47,Stage 4 Race - Tour of Watopia 2020,2020-03-21", ",")

	r := Rider{
		Name:            "Liz Rice",
		Zwid:            98588,
		LatestEventDate: time.Now(),
	}
	rr := r.Strings()
	if len(rr) != len(ss) {
		t.Fatalf("Strings length %d, expected %d", len(rr), len(ss))
	}

	for i := range rr {
		switch i {
		case 0, 1:
			if rr[i] != ss[i] {
				t.Errorf("Got %s expected %s", rr[i], ss[i])
			}
			// TODO! Test more fields
		default:
			continue

		}

	}

}

func TestFormatAsExpected(t *testing.T) {
	ss := strings.Split("Some name,98588,2020-04-15,This month,ZZRC SUB 2.0 Ride,160,https://www.zwiftpower.com/profile.php?z=98588,2.5,2.7,0,3,47,Stage 4 Race - Tour of Watopia 2020,2020-03-21", ",")

	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to log in: %v", err)
	}

	rider, err := ImportRider(client, 98588)
	if err != nil {
		t.Fatalf("Failed to get data from ZwiftPower: %v", err)
	}

	result := rider.Strings()
	if len(result) != len(ss) {
		t.Errorf("Data length %d, expected %d", len(result), len(ss))
	}

	for i := range ss {
		switch i {
		case 0:
			// Name not included in the data from this URL
			continue
		case 1, 6: // ID, URL
			// Field should be identical
			if ss[i] != result[i] {
				t.Errorf("Unexpected data field %d, got %s expected %s", i, result[i], ss[i])
			}
		case 2, 13:
			// Check for a date format
			v := strings.Split(result[i], "-")
			if len(v) != 3 {
				t.Errorf("Unexpected format for field %d, got %s expected date in format 2006-01-02", i, result[i])
			}
		case 3:
			if !strings.Contains(result[i], "month") {
				t.Errorf("Unexpected data field %d, got %s expected %s", i, result[i], ss[i])
			}
		case 4, 12:
			// Strings that will vary wildly
			continue
		case 5, 9, 10, 11:
			// Check it's an integer number
			_, err := strconv.Atoi(result[i])
			if err != nil {
				t.Errorf("Error converting field %d containing %s to integer: %v", i, result[i], err)
			}
		case 7, 8:
			// Check for a decimal format that's plausible for W/kg
			v, err := strconv.ParseFloat(result[i], 32)
			if err != nil {
				t.Errorf("Error converting field %d containing %s to float: %v", i, result[i], err)
			}

			// Ftp30 will be zero if date was > 30 days ago
			switch result[3] {
			case "This month":
				if v < 0.1 || v > 6 {
					t.Errorf("Unexpected power value in field %d from string %s", i, result[i])
				}
			default:
				if i == 7 && result[i] != "0.0" {
					t.Errorf("Should have 0.0 for 30-day FTP since last event was over a month ago, got %s", result[i])
				}
			}
		}
	}
}

func TestUnmarshalEvent(t *testing.T) {
	var r riderEvents
	err := json.Unmarshal([]byte(testdata), &r)
	if err != nil {
		t.Errorf("Failed unmarshalling: %v", err)
	}

	var e Event
	err = json.Unmarshal([]byte(testevent), &e)
	if err != nil {
		t.Errorf("Failed unmarshalling event: %v", err)
	}
}

const testdata = `{"data":[{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"4","zid":"1096124","pos":107,"position_in_cat":2,"name":"&Ouml;zge Yazar [REVO]","cp":0,"zwid":1261784,"res_id":"1096124.107","lag":0,"uid":"3153245763137311192","time":[1557.351,1],"time_gun":1557.531,"gap":113.632,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"D","height":[165,1],"flag":"ca","avg_hr":[162,0],"max_hr":[177,0],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":"525.97","skill_b":0,"skill_gain":"14.81","np":[156,0],"hrr":["0.94",0],"hreff":["60",0],"avg_power":[152,0],"avg_wkg":["2.7",0],"wkg_ftp":["2.5",0],"wftp":[143,0],"wkg_guess":0,"wkg1200":["2.7",0],"wkg300":["2.9",0],"wkg120":["3.2",0],"wkg60":["3.7",0],"wkg30":["4.4",0],"wkg15":["5.2",0],"wkg5":["7.0",1],"w1200":["151",0],"w300":["164",0],"w120":["179",0],"w60":["211",0],"w30":["249",0],"w15":["293",0],"w5":["392",1],"is_guess":0,"upg":0,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"Crit City Race","f_t":"TYPE_RACE TYPE_RACE ","distance":16,"event_date":1601736300,"rt":"2875658892","laps":"8","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"3","zid":"1102266","pos":62,"position_in_cat":20,"name":"&Ouml;zge Yazar [REVO]","cp":0,"zwid":1261784,"res_id":"1102266.62","lag":0,"uid":"566503692966670752","time":[1902.401,0],"time_gun":1902.401,"gap":313.75,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"C","height":[0,0],"flag":"ca","avg_hr":[169,0],"max_hr":[184,1],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":0,"skill_b":"585.19","skill_gain":0,"np":[158,0],"hrr":["0.91",0],"hreff":["61",0],"avg_power":[154,0],"avg_wkg":["2.7",0],"wkg_ftp":["2.6",0],"wftp":[149,0],"wkg_guess":0,"wkg1200":["2.8",0],"wkg300":["2.9",0],"wkg120":["3.3",0],"wkg60":["3.7",0],"wkg30":["4.4",0],"wkg15":["5.9",1],"wkg5":["6.2",0],"w1200":["157",0],"w300":["163",0],"w120":["188",0],"w60":["206",0],"w30":["248",0],"w15":["334",1],"w5":["349",0],"is_guess":0,"upg":0,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"Sydkysten Cycling - Carl Ras Race","f_t":"TYPE_RACE TYPE_RACE ","distance":20,"event_date":1601994600,"rt":"947394567","laps":"10","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"5","zid":"1106655","pos":80,"position_in_cat":80,"name":"&Ouml;zge Yazar [REVO]","cp":0,"zwid":1261784,"res_id":"1106655.80","lag":24,"uid":"568258994141508128","time":[4930.385,0],"time_gun":4930.625,"gap":1306.278,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"V","height":[0,0],"flag":"ca","avg_hr":[132,0],"max_hr":[160,0],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":10,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":0,"skill_b":0,"skill_gain":0,"np":[123,0],"hrr":["0.89",0],"hreff":["62",0],"avg_power":[118,0],"avg_wkg":["2.1",0],"wkg_ftp":["2.2",0],"wftp":[129,0],"wkg_guess":0,"wkg1200":["2.4",0],"wkg300":["2.6",0],"wkg120":["2.8",0],"wkg60":["3.5",0],"wkg30":["3.7",0],"wkg15":["4.1",0],"wkg5":["4.2",0],"w1200":["136",0],"w300":["147",0],"w120":["155",0],"w60":["195",0],"w30":["211",0],"w15":["233",0],"w5":["239",0],"is_guess":0,"upg":1,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"WTRL Team Time Trial - Zone 7","f_t":"TYPE_RACE","distance":43,"event_date":1602200100,"rt":"604330868","laps":"2","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"1","zid":"1121958","pos":33,"position_in_cat":33,"name":"&Ouml;zge Yazar [REVO]","cp":0,"zwid":1261784,"res_id":"1121958.33","lag":7,"uid":"3156197863137311192","time":[5978.016,0],"time_gun":5978.016,"gap":1192.669,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"A","height":[165,1],"flag":"ca","avg_hr":[163,0],"max_hr":[175,0],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":0,"skill_b":"585.19","skill_gain":0,"np":[145,0],"hrr":["0.87",0],"hreff":["65",0],"avg_power":[141,0],"avg_wkg":["2.5",0],"wkg_ftp":["2.5",0],"wftp":[140,0],"wkg_guess":0,"wkg1200":["2.6",0],"wkg300":["2.9",0],"wkg120":["3.0",0],"wkg60":["3.2",0],"wkg30":["3.5",0],"wkg15":["4.1",0],"wkg5":["5.4",0],"w1200":["148",0],"w300":["164",0],"w120":["171",0],"w60":["181",0],"w30":["199",0],"w15":["229",0],"w5":["306",0],"is_guess":0,"upg":0,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"Zwift Racing League | WTRL - AMERICAS W (WOMEN)","f_t":"TYPE_RACE TYPE_RACE ","distance":50,"event_date":1602639900,"rt":"3921412335","laps":"","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"2","zid":"1139267","pos":40,"position_in_cat":40,"name":"&Ouml;zge Yazar [REVO]","cp":0,"zwid":1261784,"res_id":"1139267.40","lag":1,"uid":"3158201863137311192","time":[2447.491,0],"time_gun":2447.551,"gap":271.537,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"A","height":[165,1],"flag":"ca","avg_hr":[168,0],"max_hr":[184,1],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":0,"skill_b":0,"skill_gain":0,"np":[161,0],"hrr":["0.95",0],"hreff":["59",0],"avg_power":[160,0],"avg_wkg":["2.8",0],"wkg_ftp":["2.7",0],"wftp":[156,0],"wkg_guess":0,"wkg1200":["2.9",0],"wkg300":["3.2",0],"wkg120":["3.3",0],"wkg60":["3.6",0],"wkg30":["4.3",0],"wkg15":["4.7",0],"wkg5":["5.1",0],"w1200":["165",0],"w300":["178",0],"w120":["183",0],"w60":["202",0],"w30":["241",0],"w15":["266",0],"w5":["286",0],"is_guess":0,"upg":1,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"Zwift Racing League | WTRL - AMERICAS W (WOMEN) - TTT","f_t":"TYPE_RACE","distance":25,"event_date":1603244700,"rt":"1776635757","laps":"1","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"1","zid":"1154781","pos":40,"position_in_cat":40,"name":"&Ouml;zge Yazar [REVO]","cp":0,"zwid":1261784,"res_id":"1154781.40","lag":0,"uid":"3159972963137311192","time":[3790.579,0],"time_gun":3790.579,"gap":772.717,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"A","height":[165,1],"flag":"ca","avg_hr":[166,0],"max_hr":[177,0],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":0,"skill_b":"585.19","skill_gain":0,"np":[159,0],"hrr":["0.93",0],"hreff":["60",0],"avg_power":[155,0],"avg_wkg":["2.8",0],"wkg_ftp":["2.7",0],"wftp":[152,0],"wkg_guess":0,"wkg1200":["2.9",0],"wkg300":["3.1",0],"wkg120":["3.2",0],"wkg60":["3.6",0],"wkg30":["3.9",0],"wkg15":["4.3",0],"wkg5":["4.7",0],"w1200":["161",0],"w300":["172",0],"w120":["182",0],"w60":["202",0],"w30":["221",0],"w15":["242",0],"w5":["264",0],"is_guess":0,"upg":0,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"Zwift Racing League | WTRL - Womens AMERICAS W DIVISION 1","f_t":"TYPE_RACE TYPE_RACE ","distance":32,"event_date":1603849500,"rt":"2196019512","laps":"2","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"1","zid":"1228095","pos":25,"position_in_cat":25,"name":"&Ouml;zge Yazar [REVO]","cp":1,"zwid":1261784,"res_id":"1228095.25","lag":0,"uid":"3168135563137311192","time":[3159.605,0],"time_gun":3159.605,"gap":399.721,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"A","height":[165,1],"flag":"ca","avg_hr":[165,0],"max_hr":[183,0],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":0,"skill_b":0,"skill_gain":0,"np":[154,0],"hrr":["0.92",0],"hreff":["61",0],"avg_power":[152,0],"avg_wkg":["2.7",0],"wkg_ftp":["2.6",0],"wftp":[149,0],"wkg_guess":0,"wkg1200":["2.8",0],"wkg300":["3.2",0],"wkg120":["3.6",0],"wkg60":["4.2",0],"wkg30":["4.9",1],"wkg15":["5.2",0],"wkg5":["5.3",0],"w1200":["157",0],"w300":["178",0],"w120":["202",0],"w60":["238",0],"w30":["276",1],"w15":["291",0],"w5":["299",0],"is_guess":0,"upg":1,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"Zwift Racing League | WTRL - Womens AMERICAS W DIVISION 1","f_t":"TYPE_RACE","distance":31,"event_date":1605667500,"rt":"1880443431","laps":"1","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"1","zid":"1261637","pos":29,"position_in_cat":29,"name":"&Ouml;zge Yazar [REVO]","cp":1,"zwid":1261784,"res_id":"1261637.29","lag":0,"uid":"3171810563137311192","time":[4990.387,0],"time_gun":4990.387,"gap":663.148,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"A","height":[165,1],"flag":"ca","avg_hr":[169,0],"max_hr":[178,0],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":"595.88","skill_b":"585.19","skill_gain":"0.82","np":[155,0],"hrr":["0.91",0],"hreff":["62",0],"avg_power":[153,0],"avg_wkg":["2.7",0],"wkg_ftp":["2.6",0],"wftp":[149,0],"wkg_guess":0,"wkg1200":["2.8",0],"wkg300":["2.9",0],"wkg120":["3.2",0],"wkg60":["3.5",0],"wkg30":["3.8",0],"wkg15":["4.3",0],"wkg5":["5.1",0],"w1200":["157",0],"w300":["164",0],"w120":["180",0],"w60":["197",0],"w30":["215",0],"w15":["242",0],"w5":["285",0],"is_guess":0,"upg":0,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"Zwift Racing League | WTRL - Womens AMERICAS W DIVISION 1","f_t":"TYPE_RACE TYPE_RACE ","distance":47,"event_date":1606272300,"rt":"2852153296","laps":"","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"1","zid":"1289214","pos":25,"position_in_cat":25,"name":"&Ouml;zge Yazar [REVO]","cp":1,"zwid":1261784,"res_id":"1289214.25","lag":0,"uid":"3174779063137311192","time":[2999.432,0],"time_gun":2999.432,"gap":197.223,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"A","height":[165,1],"flag":"ca","avg_hr":[164,0],"max_hr":[182,0],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":"498.52","skill_b":"584.37","skill_gain":"20.30","np":[155,0],"hrr":["0.95",0],"hreff":["59",0],"avg_power":[156,0],"avg_wkg":["2.8",0],"wkg_ftp":["2.6",0],"wftp":[148,0],"wkg_guess":0,"wkg1200":["2.8",0],"wkg300":["3.0",0],"wkg120":["3.4",0],"wkg60":["3.7",0],"wkg30":["4.1",0],"wkg15":["4.3",0],"wkg5":["4.5",0],"w1200":["156",0],"w300":["169",0],"w120":["190",0],"w60":["206",0],"w30":["233",0],"w15":["240",0],"w5":["254",0],"is_guess":0,"upg":0,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"Zwift Racing League | WTRL - Womens AMERICAS W DIVISION 1","f_t":"TYPE_RACE TYPE_RACE ","distance":28,"event_date":1606877100,"rt":"1064303857","laps":"1","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"1","zid":"1347837","pos":37,"position_in_cat":37,"name":"&Ouml;zge Yazar [REVO]","cp":1,"zwid":1261784,"res_id":"1347837.37","lag":0,"uid":"3181093163137311192","time":[3881.264,0],"time_gun":3881.264,"gap":354.153,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"A","height":[165,1],"flag":"ca","avg_hr":[160,0],"max_hr":[178,0],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":0,"skill_b":0,"skill_gain":0,"np":[152,0],"hrr":["0.96",0],"hreff":["58",0],"avg_power":[153,0],"avg_wkg":["2.7",0],"wkg_ftp":["2.6",0],"wftp":[150,0],"wkg_guess":0,"wkg1200":["2.8",0],"wkg300":["3.0",0],"wkg120":["3.3",0],"wkg60":["3.7",0],"wkg30":["3.9",0],"wkg15":["4.0",0],"wkg5":["4.4",0],"w1200":["158",0],"w300":["167",0],"w120":["183",0],"w60":["206",0],"w30":["217",0],"w15":["225",0],"w5":["247",0],"is_guess":0,"upg":0,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"Zwift Racing League | WTRL - Womens AMERICAS W DIVISION 1","f_t":"TYPE_RACE","distance":36,"event_date":1608086700,"rt":"3366225080","laps":"2","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"1","zid":"1389185","pos":34,"position_in_cat":34,"name":"&Ouml;zge Yazar [REVO]","cp":1,"zwid":1261784,"res_id":"1389185.34","lag":5,"uid":"3185527763137311192","time":[4781.411,0],"time_gun":4781.411,"gap":816.312,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"V","height":[165,1],"flag":"ca","avg_hr":[151,0],"max_hr":[178,0],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":0,"skill_b":"564.07","skill_gain":0,"np":[144,0],"hrr":["0.88",0],"hreff":["63",0],"avg_power":[133,0],"avg_wkg":["2.4",0],"wkg_ftp":["2.5",0],"wftp":[143,0],"wkg_guess":0,"wkg1200":["2.7",0],"wkg300":["3.1",0],"wkg120":["3.3",0],"wkg60":["3.9",0],"wkg30":["4.2",0],"wkg15":["4.7",0],"wkg5":["6.6",0],"w1200":["151",0],"w300":["173",0],"w120":["188",0],"w60":["218",0],"w30":["239",0],"w15":["263",0],"w5":["374",0],"is_guess":0,"upg":1,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"WTRL Team Time Trial Platinum League","f_t":"TYPE_RACE TYPE_RACE ","distance":32,"event_date":1608835500,"rt":"2843604888","laps":"","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"3","zid":"1497992","pos":54,"position_in_cat":15,"name":"&Ouml;zge Yazar [REVO]","cp":1,"zwid":1261784,"res_id":"1497992.54","lag":0,"uid":"3197084463137311192","time":[3582.416,0],"time_gun":3582.536,"gap":238.969,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"C","height":[165,1],"flag":"ca","avg_hr":[163,0],"max_hr":[183,0],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"27","zada":0,"note":"","div":30,"divw":30,"skill":"582.68","skill_b":"578.88","skill_gain":"3.46","np":[164,0],"hrr":["0.93",0],"hreff":["60",0],"avg_power":[152,0],"avg_wkg":["2.7",0],"wkg_ftp":["2.6",0],"wftp":[148,0],"wkg_guess":0,"wkg1200":["2.8",0],"wkg300":["3.4",1],"wkg120":["3.8",1],"wkg60":["4.6",1],"wkg30":["4.8",0],"wkg15":["5.6",0],"wkg5":["6.6",0],"w1200":["156",0],"w300":["191",0],"w120":["213",1],"w60":["257",1],"w30":["271",0],"w15":["314",0],"w5":["371",0],"is_guess":0,"upg":0,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"Zwift Racing League | WTRL - Womens AMERICAS W DIVISION 1","f_t":"TYPE_RACE TYPE_RACE ","distance":32,"event_date":1610505900,"rt":"1039983620","laps":"2","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"3","zid":"1644250","pos":51,"position_in_cat":9,"name":"&Ouml;zge Yazar [REVO]","cp":1,"zwid":1261784,"res_id":"1644250.51","lag":0,"uid":"3212539963137311192","time":[2955.764,0],"time_gun":2955.884,"gap":45.874,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"C","height":[165,1],"flag":"ca","avg_hr":[171,1],"max_hr":[180,0],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":1,"age":"27","zada":0,"note":"","div":30,"divw":30,"skill":"503.87","skill_b":"575.42","skill_gain":"19.23","np":[179,1],"hrr":["1.00",1],"hreff":["56",1],"avg_power":[171,1],"avg_wkg":["3.0",1],"wkg_ftp":["2.9",1],"wftp":[163,1],"wkg_guess":0,"wkg1200":["3.1",1],"wkg300":["3.4",1],"wkg120":["3.6",0],"wkg60":["3.9",0],"wkg30":["4.5",0],"wkg15":["5.4",0],"wkg5":["6.6",0],"w1200":["172",1],"w300":["192",1],"w120":["204",0],"w60":["219",0],"w30":["256",0],"w15":["303",0],"w5":["369",0],"is_guess":0,"upg":0,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"Zwift Racing League | WTRL - Womens AMERICAS W DIVISION 1","f_t":"TYPE_RACE TYPE_RACE ","distance":28,"event_date":1612320300,"rt":"2007026433","laps":"2","dur":""},{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"4","zid":"1118313","pos":5,"position_in_cat":0,"name":"&Ouml;zge Yazar [REVO]","cp":0,"zwid":1261784,"res_id":"1118313.5","lag":31,"uid":"3155754963137311192","time":[3600,0],"time_gun":3600,"gap":0,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"N\/A","height":[165,1],"flag":"ca","avg_hr":[134,0],"max_hr":[152,1],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":10,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":0,"skill_b":0,"skill_gain":0,"np":[107,0],"hrr":["0.76",0],"hreff":["73",0],"avg_power":[102,0],"avg_wkg":["1.8",0],"wkg_ftp":["1.7",0],"wftp":[100,0],"wkg_guess":0,"wkg1200":["1.9",0],"wkg300":["2.0",0],"wkg120":["2.1",0],"wkg60":["2.5",0],"wkg30":["2.6",0],"wkg15":["2.7",0],"wkg5":["2.7",0],"w1200":["106",0],"w300":["114",0],"w120":["121",0],"w60":["141",0],"w30":["148",0],"w15":["150",0],"w5":["150",0],"is_guess":0,"upg":0,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"REVO Social SUB2","f_t":"TYPE_RIDE","distance":0,"event_date":"","rt":"1776635757","laps":"","dur":"3600"}]}`

const testevent = `{"DT_RowId":"","ftp":"170","friend":0,"pt":"","label":"4","zid":"1118313","pos":5,"position_in_cat":0,"name":"&Ouml;zge Yazar [REVO]","cp":0,"zwid":1261784,"res_id":"1118313.5","lag":31,"uid":"3155754963137311192","time":[3600,0],"time_gun":3600,"gap":0,"vtta":"","vttat":0,"male":0,"tid":"2672","topen":"","tname":"REVO","tc":"fc00e3","tbc":"000000","tbd":"fc00e3","zeff":0,"category":"N\/A","height":[165,1],"flag":"ca","avg_hr":[134,0],"max_hr":[152,1],"hrmax":[0,0],"hrm":1,"weight":["56.3",1],"power_type":3,"display_pos":1,"src":10,"age":"26","zada":0,"note":"","div":30,"divw":30,"skill":0,"skill_b":0,"skill_gain":0,"np":[107,0],"hrr":["0.76",0],"hreff":["73",0],"avg_power":[102,0],"avg_wkg":["1.8",0],"wkg_ftp":["1.7",0],"wftp":[100,0],"wkg_guess":0,"wkg1200":["1.9",0],"wkg300":["2.0",0],"wkg120":["2.1",0],"wkg60":["2.5",0],"wkg30":["2.6",0],"wkg15":["2.7",0],"wkg5":["2.7",0],"w1200":["106",0],"w300":["114",0],"w120":["121",0],"w60":["141",0],"w30":["148",0],"w15":["150",0],"w5":["150",0],"is_guess":0,"upg":0,"penalty":"","reg":1,"fl":"","pts":"","pts_pos":"","info":0,"info_notes":[],"strike":-1,"event_title":"REVO Social SUB2","f_t":"TYPE_RIDE","distance":0,"event_date":1602590400,"rt":"1776635757","laps":"","dur":"3600"}`
