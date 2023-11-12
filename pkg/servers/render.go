package servers

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// func (sm *Manager) RenderServers() func(ctx context.Context, chatId int64) error {
// 	return func(ctx context.Context, chatId int64) error {
// 		servers, err := sm.GetServers(ctx)
// 		if err != nil {
// 			return err
// 		}
// 		err = SendServersData(chatId, serverPrefixCommand, servers, sm.bot)
// 		if err != nil {
// 			log.Printf("An error occured: %s", err.Error())
// 		}
// 		return nil
// 	}
// }

func (sm *Manager) RenderServerId(serverId string) func(ctx context.Context, chatId int64) (tgbotapi.MessageConfig, error) {
	return func(ctx context.Context, chatId int64) (tgbotapi.MessageConfig, error) {
		msg := tgbotapi.MessageConfig{}
		if len(sm.servers) == 0 {
			_, err := sm.GetServers(ctx)
			if err != nil {
				log.Printf("An error occured: %s", err.Error())
				return msg, err
			}
		}
		s, found := sm.GetServerById(serverId)
		if !found {
			message := "El server seleccionado no se ha encontrado"
			msg = tgbotapi.NewMessage(chatId, message)
			return msg, nil
		}

		if !s.Online {
			message := "El server seleccionado no está online"
			msg = tgbotapi.NewMessage(chatId, message)
			return msg, nil
		}
		si, err := s.GetSessionInfo(ctx)
		if err != nil {
			log.Printf("An error occured: %s", err.Error())
			return msg, err
		}

		msg = tgbotapi.NewMessage(chatId,
			fmt.Sprintf(`Server %s:

‣ Circuito: %s (%0.fm)
‣ Sesión: %s (Vueltas: %d)
‣ Coches: %d
‣ Lluvia: %d%% (min: %d%%. max: %d%%)
‣ Temperatura (Pista/Ambiente): %0.fºC/%0.fºC
`, serverId, si.TrackName, si.LapDistance, si.Session, si.MaximumLaps, si.NumberOfVehicles, si.Raining, si.MinPathWetness, si.MaxPathWetness, si.TrackTemp, si.AmbientTemp))
		// msg.ParseMode = tgbotapi.ModeMarkdownV2

		return msg, nil
	}
}
