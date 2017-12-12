package main

import (
	"time"
)

type SoccerPlayer struct {
	Url             string                    `json:"url"`
	Name            string                    `json:"name"`
	FullName        string                    `json:"fullname"`
	Age             uint                      `json:"age"`
	Birthday        time.Time                 `json:"birthday"`
	Height          string                    `json:"height"`
	Weight          string                    `json:"weight"`
	Overall         int                       `json:"overall"`
	Potential       int                       `json:"potential"`
	Value           string                    `json:"value"`
	Wage            string                    `json:"wage"`
	Foot            string                    `json:"foot"`
	Reputation      int                       `json:"reputation"`
	WeakFoot        int                       `json:"weak_foot"`
	SkillMoves      int                       `json:"skill_moves"`
	Team            string                    `json:"team"`
	Country         string                    `json:"country"`
	TeamPosition    string                    `json:"team_position"`
	CountryPosition string                    `json:"json:country_position"`
	TeamNumber      int                       `json:"team_number"`
	CountryNumber   int                       `json:"country_number"`
	Properties      []PlayerPropertyContainer `json:"properties"`
}

type PlayerPropertyContainer struct {
	Name       string           `json:"name"`
	Properties []PlayerProperty `json:"properties"`
}

type PlayerProperty struct {
	Name  string `json:"name"`
	Score string `json:"score"`
}
