package discord

import (
	"wednesday.wtf/godin/api/discord/file"
	"wednesday.wtf/godin/api/discord/service"
	"wednesday.wtf/godin/api/discord/session"
)

type DiscordAPIBase struct {
}

var Service DiscordAPIBase = DiscordAPIBase{}

func (service *DiscordAPIBase) Start() {

}

func (service *DiscordAPIBase) Stop() {

}

type DiscordAPI struct {
	DiscordAPIBase

	service.DiscordAPIService
	file.DiscordAPIFile
	session.DiscordAPISession
}

var Discord DiscordAPI = DiscordAPI{}
