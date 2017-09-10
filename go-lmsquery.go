package lmsquery

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Lmsquery struct {
	host      string
	port      int
	serverUrl string
}

type lmsRequest struct {
	Id     int            `json:"id"`
	Method string         `json:"method"`
	Params lmsCommandArgs `json:"params"`
}

type lmsCommandArgs struct {
	PlayerId string
	Args     []string
}

func (ca lmsCommandArgs) MarshalJSON() ([]byte, error) {
	var args string = ""
	for _, arg := range ca.Args {
		if args != "" {
			args += ", "
		}
		args += "\"" + string(arg) + "\""
	}
	return []byte(fmt.Sprintf("[\"%s\", [%s]]", ca.PlayerId, args)), nil
}

type lmsResponse struct {
	Id     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	Result Result        `json:"result"`
}

type Result struct {
	PlayerCount      string       `json:"player_count"`
	PlayersLoop      []playerInfo `json:"players_loop"`
	AlarmsLoop       []Alarm      `json:"alarms_loop"`
	InfoTotalAlbums  string       `json:"info_total_albums"`
	InfoTotalArtists string       `json:"info_total_artists"`
	InfoTotalSongs   string       `json:"info_total_songs"`
	InfoTotalGenres  string       `json:"info_total_genres"`
	Power            string       `json:"_power"`
	Param0           string       `json:"_p0"`
	Param1           string       `json:"_p1"`
	Param2           string       `json:"_p2"`
	Param3           string       `json:"_p3"`
	Param4           string       `json:"_p4"`
	Count            int          `json:"count"`
}

type Alarm struct {
	Enabled       string `json:"enabled"`
	Dow           string `json:"dow"`
	Time          string `json:"time"`
	Volume        string `json:"volume"`
	Repeat        string `json:"repeat"`
	Url           string `json:"url"`
	NextExecution time.Time
}
type playerInfo struct {
	IsPlaying int    `json:"isplaying"`
	PlayerId  string `json:"playerid"`
	Power     int    `json:"power"`
	Name      string `json:"name"`
	Ip        string `json:"ip"`
}

// send a generic query to lms
func (l *Lmsquery) query(params []string, playerId string) Result {
	args := lmsCommandArgs{playerId, params}
	req := lmsRequest{Id: 1, Method: "slim.request", Params: args}

	js, err := json.Marshal(req)
	if err != nil {
		panic(err.Error())
	}
	logtime := []byte(time.Now().String())
	log.Printf("request [%x]: call %v with %v\n", sha1.Sum(logtime), l.serverUrl, string(js))

	res, err := http.Post(l.serverUrl, "application/json", bytes.NewBuffer(js))
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("response [%x]: %v\n", sha1.Sum(logtime), string(body))

	var result lmsResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatal(err)
	}

	return result.Result
}

/**
Server functions
*/
func (l *Lmsquery) GetServerStatus() Result {
	return l.query([]string{"serverstatus", "0", "99"}, "")
}

func (l *Lmsquery) GetPlayers() []playerInfo {
	return l.GetServerStatus().PlayersLoop
}

/**
Player functions
*/
func (l *Lmsquery) SetPower(playerId string, power int) {
	l.query([]string{"power", strconv.Itoa(power)}, playerId)
}
func (l *Lmsquery) SetPowerAll(power int) {
	for _, player := range l.GetPlayers() {
		l.SetPower(player.PlayerId, power)
	}
}
func (l *Lmsquery) IsPowered(playerId string) bool {
	return l.query([]string{"power", "?"}, playerId).Power == "1"
}

func (l *Lmsquery) GetPlayerPref(playerId string, pref string) string {
	return l.query([]string{"playerpref", pref, "?"}, playerId).Param2
}

func (l *Lmsquery) GetAlarms(playerId string, enabled bool) (int, []Alarm) {
	filter := ""
	if enabled {
		if l.GetPlayerPref(playerId, "alarmsEnabled") == "0" {
			return 0, nil
		}

		filter = "filter:enabled"
	} else {
		filter = "filter:all"
	}
	res := l.query([]string{"alarms", "0", "99", filter}, playerId)
	count := res.Count
	// fill next alarm times
	loc, _ := time.LoadLocation("Europe/Berlin")
	for idx, alarm := range res.AlarmsLoop {
		if alarm.Enabled == "0" || alarm.Dow == "" {
			continue
		}
		date := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, loc)
		sec, _ := strconv.Atoi(alarm.Time)
		date = date.Add(time.Duration(sec) * time.Second)
		if date.Before(time.Now()) {
			log.Println("Too late, lets search for the next execution")
			dows := strings.Split(alarm.Dow, ",")
			found := false
			for !found {
				date = date.Add(time.Duration(24) * time.Hour)
				for _, d := range dows {
					dow, _ := strconv.Atoi(d)
					if int(date.Weekday()) == dow {
						log.Println("Got it!")
						log.Println(date)
						found = true
						break
					}
				}
			}
		}
		res.AlarmsLoop[idx].NextExecution = date
	}
	return count, res.AlarmsLoop
}

func (l *Lmsquery) GetNextAlarm(playerId string) (bool, Alarm) {
	_, alarms := l.GetAlarms(playerId, true)
	nextAlarmIdx := 0
	found := false
	for idx := range alarms {
		found = true
		if alarms[idx].NextExecution.Before(alarms[nextAlarmIdx].NextExecution) {
			nextAlarmIdx = idx
		}
	}
	if found {
		return found, alarms[nextAlarmIdx]
	} else {
		return found, *new(Alarm)
	}
}

func CreateLms(host string, port int) Lmsquery {
	return Lmsquery{host: host, port: port, serverUrl: "http://" + host + ":" + strconv.Itoa(port) + "/jsonrpc.js"}
}
