package live

import (
	"context"
	"f1champshotlapsbot/pkg/apps"
	"f1champshotlapsbot/pkg/caster"
	"f1champshotlapsbot/pkg/helper"
	"f1champshotlapsbot/pkg/menus"
	"f1champshotlapsbot/pkg/pubsub"
	"f1champshotlapsbot/pkg/servers"
	"f1champshotlapsbot/pkg/thumbnails"
	"fmt"
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	buttonStint = "Tanda"
	buttonGrid  = "Parrilla"
	buttonInfo  = "Info"
)

type ServerApp struct {
	bot                           *tgbotapi.BotAPI
	appMenu                       menus.ApplicationMenu
	menuKeyboard                  tgbotapi.ReplyKeyboardMarkup
	gridApp                       *GridApp
	stintApp                      *StintApp
	accepters                     []apps.Accepter
	serverID                      string
	liveSessionInfoData           servers.LiveSessionInfoData
	liveSessionInfoDataUpdateChan <-chan string
	liveSessionInfoDataCaster     caster.ChannelCaster[servers.LiveSessionInfoData]
	trackThumbnailData            thumbnails.Thumbnail
	trackThumbnailDataUpdateChan  <-chan string
	trackThumbnailCaster          caster.ChannelCaster[thumbnails.Thumbnail]
	mu                            sync.Mutex
}

func sanitizeServerName(name string) string {
	fixed := strings.TrimPrefix(name, servers.ServerStatusOnline)
	fixed = strings.TrimPrefix(fixed, servers.ServerStatusOffline)
	fixed = strings.TrimPrefix(fixed, servers.ServerStatusOnlineButNotData)
	return strings.TrimSpace(fixed)
}

func NewServerApp(bot *tgbotapi.BotAPI, appMenu menus.ApplicationMenu, pubsubMgr *pubsub.PubSub, serverID, serverURL string) *ServerApp {
	sa := &ServerApp{
		bot:                           bot,
		appMenu:                       appMenu,
		serverID:                      serverID,
		liveSessionInfoDataCaster:     caster.JSONChannelCaster[servers.LiveSessionInfoData]{},
		liveSessionInfoDataUpdateChan: pubsubMgr.Subscribe(servers.PubSubSessionInfoPreffix + serverID),
		trackThumbnailCaster:          caster.JSONChannelCaster[thumbnails.Thumbnail]{},
		trackThumbnailDataUpdateChan:  pubsubMgr.Subscribe(thumbnails.PubSubThumbnailPreffix + serverID),
	}

	go sa.liveSessionInfoUpdater()
	go sa.trackThumbnailUpdater()

	gridAppMenu := menus.NewApplicationMenu("", serverID, sa)
	gridApp := NewGridApp(bot, gridAppMenu, pubsubMgr, serverID)

	stintAppMenu := menus.NewApplicationMenu("", serverID, sa)
	stintApp := NewStintApp(bot, stintAppMenu, pubsubMgr, serverID, serverURL)

	accepters := []apps.Accepter{gridApp, stintApp}

	sa.accepters = accepters
	sa.gridApp = gridApp
	sa.stintApp = stintApp
	return sa
}

func (sa *ServerApp) update(lsid servers.LiveSessionInfoData, t thumbnails.Thumbnail) {
	stint := buttonStint + " " + lsid.ServerName
	grid := buttonGrid + " " + lsid.ServerName
	info := buttonInfo + " " + lsid.ServerName
	menuKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(stint),
			tgbotapi.NewKeyboardButton(grid),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(sa.appMenu.ButtonBackTo()),
			tgbotapi.NewKeyboardButton(info),
		),
	)

	sa.menuKeyboard = menuKeyboard
	sa.liveSessionInfoData = lsid
	sa.trackThumbnailData = t
}

func (sa *ServerApp) liveSessionInfoUpdater() {
	for payload := range sa.liveSessionInfoDataUpdateChan {
		// fmt.Println("Updating SessionInfo")
		s, err := sa.liveSessionInfoDataCaster.From(payload)
		if err != nil {
			log.Printf("Error casting SessionInfo: %s\n", err.Error())
			continue
		}
		sa.mu.Lock()
		sa.update(s, sa.trackThumbnailData)
		sa.mu.Unlock()
	}
}

func (sa *ServerApp) trackThumbnailUpdater() {
	for payload := range sa.trackThumbnailDataUpdateChan {
		// fmt.Println("Updating SessionInfo")
		t, err := sa.trackThumbnailCaster.From(payload)
		if err != nil {
			log.Printf("Error casting thumbnail: %s\n", err.Error())
			continue
		}
		sa.mu.Lock()
		sa.update(sa.liveSessionInfoData, t)
		sa.mu.Unlock()
	}
}

func (sa *ServerApp) Menu() tgbotapi.ReplyKeyboardMarkup {
	return sa.menuKeyboard
}

func (sa *ServerApp) AcceptCommand(command string) (bool, func(ctx context.Context, chatId int64) error) {
	for _, accepter := range sa.accepters {
		accept, handler := accepter.AcceptCommand(command)
		if accept {
			return true, handler
		}
	}

	return false, nil
}

func (sa *ServerApp) AcceptCallback(query *tgbotapi.CallbackQuery) (bool, func(ctx context.Context, query *tgbotapi.CallbackQuery) error) {
	for _, accepter := range sa.accepters {
		accept, handler := accepter.AcceptCallback(query)
		if accept {
			return true, handler
		}
	}

	return false, nil
}

func (sa *ServerApp) AcceptButton(button string) (bool, func(ctx context.Context, chatId int64) error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// fmt.Printf("SERVER: button: %s. appName: %s\n", button, sa.sessionInfo.ServerName)
	if sanitizeServerName(button) == sa.liveSessionInfoData.ServerName ||
		sanitizeServerName(button) == buttonInfo+" "+sa.liveSessionInfoData.ServerName {
		return true, func(ctx context.Context, chatId int64) error {
			if !sa.liveSessionInfoData.SessionInfo.WebSocketRunning {
				msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("El servidor %s está apagado", sa.liveSessionInfoData.ServerName))
				msg.ReplyMarkup = sa.appMenu.PrevMenu()
				_, err := sa.bot.Send(msg)
				return err
			} else if !sa.liveSessionInfoData.SessionInfo.RecevingData {
				msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("No se reciben datos de server %s", sa.liveSessionInfoData.ServerName))
				msg.ReplyMarkup = sa.appMenu.PrevMenu()
				_, err := sa.bot.Send(msg)
				return err
			}
			si := sa.liveSessionInfoData.SessionInfo
			laps := "No Limitado"
			if si.MaximumLaps < 100 {
				laps = fmt.Sprintf("%d", si.MaximumLaps)
			}
			text := fmt.Sprintf(`%s:
			‣ Circuito: %s (%0.fm)
			‣ Tiempo restante: %s
			‣ Sesión: %s (Vueltas: %s)
			‣ Coches: %d
			‣ Lluvia: %.1f%% (min: %.1f%%. max: %.1f%%)
			‣ Temperatura (Pista/Ambiente): %0.fºC/%0.fºC
			`,
				sa.liveSessionInfoData.ServerName,
				si.TrackName,
				si.LapDistance,
				helper.SecondsToHoursAndMinutes(si.EndEventTime-si.CurrentEventTime),
				si.Session,
				laps,
				si.NumberOfVehicles,
				si.Raining,
				si.MinPathWetness,
				si.MaxPathWetness,
				si.TrackTemp,
				si.AmbientTemp)
			err := fmt.Errorf("No track thumbnail available")
			var rfd tgbotapi.RequestFileData
			if !sa.trackThumbnailData.IsZero() {
				rfd, err = sa.trackThumbnailData.FileData()
			}
			var cfg tgbotapi.Chattable
			if err != nil {
				log.Printf("Error getting thumbnail data: %s\n", err.Error())
				msg := tgbotapi.NewMessage(chatId, text)
				msg.ReplyMarkup = sa.menuKeyboard
				cfg = msg
			} else {
				msg := tgbotapi.NewPhoto(chatId, rfd)
				msg.Caption = text
				msg.ReplyMarkup = sa.menuKeyboard
				cfg = msg
			}
			_, err = sa.bot.Send(cfg)
			return err
		}
	} else if button == sa.appMenu.ButtonBackTo() {
		return true, func(ctx context.Context, chatId int64) error {
			msg := tgbotapi.NewMessage(chatId, "OK")
			msg.ReplyMarkup = sa.appMenu.PrevMenu()
			_, err := sa.bot.Send(msg)
			return err
		}
	} else {
		for _, accepter := range sa.accepters {
			accept, handler := accepter.AcceptButton(button)
			if accept {
				return true, handler
			}
		}
		// fmt.Print("SERVER: FALSE\n")
		return false, nil
	}
}
