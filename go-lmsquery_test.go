package lmsquery

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestLmsquery_GetServerStatus(t *testing.T) {
	host := os.Getenv("TEST_HOST")
	port, _ := strconv.Atoi(os.Getenv("TEST_PORT"))
	lms := CreateLms(host, port)
	status := lms.GetServerStatus()
	t.Log(status)
}

func TestLmsquery_GetPlayers(t *testing.T) {
	host := os.Getenv("TEST_HOST")
	port, _ := strconv.Atoi(os.Getenv("TEST_PORT"))
	lms := CreateLms(host, port)
	players := lms.GetPlayers()
	t.Log(players)
}

func TestLmsquery_GetAlarms(t *testing.T) {
	host := os.Getenv("TEST_HOST")
	port, _ := strconv.Atoi(os.Getenv("TEST_PORT"))
	lms := CreateLms(host, port)
	players := lms.GetPlayers()
	count, alarms := lms.GetAlarms(players[0].PlayerId, true)
	t.Log(count)
	t.Log(alarms)
}

func TestLmsquery_GetNextAlarm(t *testing.T) {
	host := os.Getenv("TEST_HOST")
	port, _ := strconv.Atoi(os.Getenv("TEST_PORT"))
	lms := CreateLms(host, port)

	var playerId string
	players := lms.GetPlayers()
	for _, player := range players {
		if strings.Contains(player.Name, os.Getenv("TEST_PLAYER")) {
			playerId = player.PlayerId
		}
	}
	found, alarm := lms.GetNextAlarm(playerId)
	t.Log(found)
	t.Log(alarm)
}
