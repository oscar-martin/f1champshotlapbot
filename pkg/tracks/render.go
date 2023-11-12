package tracks

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (tm *Manager) renderShowTracksCallback(data []string) func(ctx context.Context, query *tgbotapi.CallbackQuery) {
	return func(ctx context.Context, query *tgbotapi.CallbackQuery) {
		tracks, err := tm.GetTracks(ctx)
		if err != nil {
			log.Printf("An error occured: %s", err.Error())
			message := "No hay circuitos disponibles"
			msg := tgbotapi.NewMessage(query.Message.Chat.ID, message)
			_, err = tm.bot.Send(msg)
			if err != nil {
				log.Printf("An error occured: %s", err.Error())
			}
			return
		}
		maxPages := len(tracks) / tracksPerPage
		HandleTrackDataCallbackQuery(query.Message.Chat.ID, query.Message.MessageID, maxPages, tm, data[1:]...)
	}
}

func (tm *Manager) renderSessionsCallback(data []string) func(ctx context.Context, query *tgbotapi.CallbackQuery) {
	return func(ctx context.Context, query *tgbotapi.CallbackQuery) {
		HandleSessionDataCallbackQuery(query.Message.Chat.ID, &query.Message.MessageID, tm, data[1:]...)
	}
}

func (tm *Manager) renderTracks() func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		tracks, err := tm.GetTracks(ctx)
		if err != nil {
			return err
		}

		if len(tracks) > 0 {
			err := SendTracksData(chatId, 0, tracksPerPage, len(tracks)/tracksPerPage, nil, tm)
			if err != nil {
				return err
			}
		} else {
			message := "No hay circuitos disponibles"
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = tm.bot.Send(msg)
			return err
		}
		return nil
	}
}

func (tm *Manager) renderCategoriesForTrackId(trackId int) func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		track, found := tm.GetTrackByID(fmt.Sprint(trackId))
		if !found {
			return tm.renderTrackNotFound(chatId)
		}
		cats, err := track.GetCategories(ctx, tm.apiDomain)
		if err != nil {
			return err
		}

		message := fmt.Sprintf("Elige categoría para %s:\n\n", track.Name)
		if len(cats) > 0 {
			categoriesStrings := make([]string, len(cats))
			for i, cat := range cats {
				categoriesStrings[i] = cat.CommandString(track.ID)
			}

			message += strings.Join(categoriesStrings, "\n")
		} else {
			message = "No hay categorías para este circuito"
		}
		msg := tgbotapi.NewMessage(chatId, message)
		_, err = tm.bot.Send(msg)

		return err
	}
}

func (tm *Manager) renderSessionForCategoryAndTrack(trackId string, categoryId string) func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		t, found := tm.GetTrackByID(trackId)
		if !found {
			return tm.renderTrackNotFound(chatId)
		}
		_, _ = t.GetCategories(ctx, tm.apiDomain)

		err := SendSessionData(chatId, nil, trackId, categoryId, inlineKeyboardTimes, tm)
		if err != nil {
			log.Printf("An error occured: %s", err.Error())
		}
		return nil
	}
}

func (tm *Manager) renderTrackNotFound(chatId int64) error {
	message := fmt.Sprintf("El circuito seleccionado no se ha encontrado. Vuelve a listarlos con %s", MenuTracks)
	msg := tgbotapi.NewMessage(chatId, message)
	_, err := tm.bot.Send(msg)
	return err
}

func (tm *Manager) renderCurrentSession() func(ctx context.Context, chatId int64) error {
	return func(ctx context.Context, chatId int64) error {
		tracks, err := tm.GetTracks(ctx)
		if err != nil {
			return err
		}

		if len(tracks) > 0 {
			track := tracks[0]
			cats, err := track.GetCategories(ctx, tm.apiDomain)
			if err != nil {
				return err
			}

			if len(cats) > 0 {
				selectedCat := Category{}
				for _, cat := range cats {
					if len(cat.Sessions) > 0 {
						selectedCat = cat
						break
					}
				}
				if len(selectedCat.Sessions) == 0 {
					message := "No hay sesiones disponibles"
					msg := tgbotapi.NewMessage(chatId, message)
					_, err = tm.bot.Send(msg)
					return err
				}

				for _, cat := range cats {
					if len(cat.Sessions) > 0 {
						if selectedCat.Sessions[0].Time < cat.Sessions[0].Time {
							selectedCat = cat
						}
					}
				}
				return tm.renderSessionForCategoryAndTrack(track.ID, selectedCat.ID)(ctx, chatId)
			} else {
				message := "No hay sesiones disponibles"
				msg := tgbotapi.NewMessage(chatId, message)
				_, err = tm.bot.Send(msg)
				return err
			}

		} else {
			message := "No hay circuitos disponibles"
			msg := tgbotapi.NewMessage(chatId, message)
			_, err = tm.bot.Send(msg)
			return err
		}
	}
}
