package handler

import (
	"fmt"
	"goantisc/internal/log"
	"strings"

	tele "gopkg.in/telebot.v4"
)

type ForbiddenWordsChecker interface {
	FilterForbiddenWords(sentence string) []string
}

type DeleteRequester interface {
	PushDeleteRequest(chatId int64, msgId int64)
}

func MountGroupMessageHandler(bot *tele.Bot, checker ForbiddenWordsChecker, delReuqester DeleteRequester) *GroupMessageHandler {
	ret := &GroupMessageHandler{
		bot:          bot,
		checker:      checker,
		delRequester: delReuqester,
	}
	bot.Handle(tele.OnText, ret.handleText)
	return ret
}

type GroupMessageHandler struct {
	bot          *tele.Bot
	checker      ForbiddenWordsChecker
	delRequester DeleteRequester
}

func (h *GroupMessageHandler) handleText(ctx tele.Context) error {
	diff := h.checker.FilterForbiddenWords(ctx.Text())
	if len(diff) == 0 {
		return nil
	}

	// Try to delete message.
	if err := ctx.Delete(); err != nil {
		// handle the message delete failed.
		return nil
	}

	if len(diff) <= 10 {
		// handle the message be deleted.
		tipMsg, err := h.bot.Send(ctx.Recipient(), fmt.Sprintf("Message deleted successfully. Reason: [%s]", strings.Join(diff, ",")))
		if err != nil {
			log.Logger().Errorf("[Goantib] Send the tips failed err: %v", err)
			return err
		}

		h.delRequester.PushDeleteRequest(tipMsg.Chat.ID, int64(tipMsg.ID))
	}
	return nil
}
