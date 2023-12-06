package pubsub

import (
	"f1champshotlapsbot/pkg/model"
	"f1champshotlapsbot/pkg/thumbnails"
)

var (
	LiveSessionInfoDataPubSub = NewPubSub[model.LiveSessionInfoData]()
	LiveStandingDataPubSub    = NewPubSub[model.LiveStandingData]()
	LiveStandingHistoryPubSub = NewPubSub[model.LiveStandingHistoryData]()
	TrackThumbnailPubSub      = NewPubSub[thumbnails.Thumbnail]()
)
