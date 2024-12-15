package services

import (
	"context"
	"strconv"
	"sync"
	"time"

	"gopkg.in/telebot.v4"
)

type DeleteMessageIndex struct {
	chatId int64
	msgId  int64
}

type DeleteMessagePayload struct {
	chatId     int64
	msgId      int64
	hasDeleted bool

	ts time.Time

	mu sync.Mutex
}

func NewDeleteWatcher(bot *telebot.Bot) *DeleteWatcher {
	return &DeleteWatcher{
		bot: bot,
		wg:  sync.WaitGroup{},
		mu:  sync.Mutex{},

		reqMap: make(map[DeleteMessageIndex]*DeleteMessagePayload),
	}
}

type DeleteWatcher struct {
	bot        *telebot.Bot
	ticker     *time.Ticker
	cancelFunc context.CancelFunc

	reqMap map[DeleteMessageIndex]*DeleteMessagePayload // value as retry number.

	wg sync.WaitGroup
	mu sync.Mutex
}

func (dw *DeleteWatcher) Start() {
	serviceCtx, cancelFunc := context.WithCancel(context.Background())
	dw.cancelFunc = cancelFunc
	dw.ticker = time.NewTicker(time.Second * 2)
	dw.wg.Add(1)
	go dw.watch(serviceCtx)
}

func (dw *DeleteWatcher) Stop() {
	if dw.cancelFunc != nil {
		dw.cancelFunc()
	}

	dw.wg.Wait()
}

func (dw *DeleteWatcher) watch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// on stop
			dw.wg.Done()
			return
		case <-dw.ticker.C:
			dw.doCheckDeleteRequest()
		}
	}
}

func (dw *DeleteWatcher) PushDeleteRequest(chatId int64, msgId int64) {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	if _, ok := dw.reqMap[DeleteMessageIndex{chatId: chatId, msgId: msgId}]; !ok {
		dw.reqMap[DeleteMessageIndex{chatId: chatId, msgId: msgId}] = &DeleteMessagePayload{
			chatId:     chatId,
			msgId:      msgId,
			hasDeleted: false,
			ts:         time.Now().UTC(),
		}
	}
}

func (dw *DeleteWatcher) doCheckDeleteRequest() {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	now := time.Now().UTC()
	for index, nPayload := range dw.reqMap {
		payload := nPayload

		if payload.hasDeleted {
			delete(dw.reqMap, index)
			continue
		}

		if now.Sub(payload.ts) < time.Second*8 {
			continue
		}

		go func() {
			payload.mu.Lock()
			defer payload.mu.Unlock()

			if payload.hasDeleted {
				return
			}

			err := dw.bot.Delete(&telebot.StoredMessage{
				MessageID: strconv.Itoa(int(index.msgId)),
				ChatID:    index.chatId,
			})

			if err != nil {
				if _, ok := err.(*telebot.Error); !ok {
					return
				}
			}
			payload.hasDeleted = true
		}()
	}

}
