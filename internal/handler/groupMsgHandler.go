package handler

import (
	"fmt"
	"goantisc/internal/core"
	"goantisc/internal/data"
	"goantisc/internal/log"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	tele "gopkg.in/telebot.v4"
)

type ForbiddenWordsChecker interface {
	FilterForbiddenWords(sentence string) ([]string, int)
}

type DeleteRequester interface {
	PushDeleteRequest(chatId int64, msgId int64)
}

type DeleteReason struct {
	NGRate int
	Reason string
	Field  string
}

func MountGroupMessageHandler(
	bot *tele.Bot,
	serviceCfg *core.ServiceConfig,
	checker ForbiddenWordsChecker,
	delReuqester DeleteRequester,
	chatStore *data.ChatInfoStore,
) *GroupMessageHandler {
	ret := &GroupMessageHandler{
		bot:           bot,
		checker:       checker,
		delRequester:  delReuqester,
		servConfig:    serviceCfg,
		chatInfoStore: chatStore,
	}
	// OnGroupMessage
	bot.Handle(tele.OnText, ret.handleGroupMessage)
	return ret
}

type GroupMessageHandler struct {
	bot           *tele.Bot
	checker       ForbiddenWordsChecker
	chatInfoStore *data.ChatInfoStore
	delRequester  DeleteRequester
	servConfig    *core.ServiceConfig

	fg singleflight.Group
}

func (h *GroupMessageHandler) handleGroupMessage(ctx tele.Context) error {
	// if ctx.Chat().Type != tele.ChatGroup || ctx.Chat().Type != tele.ChatSuperGroup {
	// 	return nil
	// }

	now := time.Now().UTC()
	chatId := ctx.Chat().ID
	chatIdStr := strconv.Itoa(int(ctx.Chat().ID))
	// Use singleflight to refresh chatinfo data
	chatInfo, err, _ := h.fg.Do(chatIdStr, func() (any, error) {
		// try to get the chat info.
		saved := h.chatInfoStore.GetChatInfo(chatId)
		if now.Sub(saved.UpdateAt) > time.Minute*10 {
			if err := h.refreshChatInfo(chatId); err != nil {
				return saved, err
			}
		}

		ret := h.chatInfoStore.GetChatInfo(chatId)
		return ret, nil
	})

	if err != nil {
		log.Logger().Errorf("[GroupMessage] refresh chatinfo failed. err: %s", err.Error())
	}
	return h.handleText(ctx, chatInfo.(*data.ChatInfo))
}

func (h *GroupMessageHandler) handleText(ctx tele.Context, chatInfo *data.ChatInfo) error {
	reason := h.checkDeleteReason(ctx, chatInfo)

	// Try to delete message.
	if err := ctx.Delete(); err != nil {
		// handle the message delete failed.
		return nil
	}

	userName := ctx.Sender().FirstName + ctx.Sender().LastName
	// handle the message be deleted.
	tipMsg, err := h.bot.Send(
		ctx.Recipient(),
		fmt.Sprintf("Deleted message from [%s](tg://user?id=%d). Reason: [%s]", userName, ctx.Sender().ID, reason.Reason),
		tele.ModeMarkdown,
	)
	if err != nil {
		log.Logger().Errorf("[Goantib] Send the tips failed err: %v", err)
		return err
	}
	h.delRequester.PushDeleteRequest(tipMsg.Chat.ID, int64(tipMsg.ID))

	chatName := ctx.Chat().FirstName + ctx.Chat().LastName
	log.Logger().Infof("[Goantib] Deleted a message from %s in chat %s. Reason: [%s]", userName, chatName, reason.Reason)
	return nil
}

func (h *GroupMessageHandler) checkDeleteReason(ctx tele.Context, chatInfo *data.ChatInfo) *DeleteReason {
	// Calculate the content NGRate
	contentDiff, originNum := h.checker.FilterForbiddenWords(ctx.Text())
	contentNGRate := h.calcNGRate(h.calcNGPoint(contentDiff, chatInfo), originNum)

	if contentNGRate >= 20 {
		if len(contentDiff) <= 15 {
			return &DeleteReason{
				NGRate: contentNGRate,
				Field:  "content",
				Reason: strings.Join(contentDiff, ","),
			}
		} else {
			return &DeleteReason{
				NGRate: contentNGRate,
				Field:  "content",
				Reason: "Too many",
			}
		}
	}

	// Calculate the username NGRate
	senderName := ctx.Sender().FirstName + ctx.Sender().LastName
	nameDiff, nameOriginNum := h.checker.FilterForbiddenWords(senderName)
	nameNGRate := h.calcNGRate(h.calcNGPoint(nameDiff, chatInfo), nameOriginNum)
	if contentNGRate >= 20 {
		return &DeleteReason{
			NGRate: nameNGRate,
			Field:  "name",
			Reason: strings.Join(contentDiff, ","),
		}
	}

	return nil
}

func (h *GroupMessageHandler) calcNGPoint(diff []string, chatInfo *data.ChatInfo) int {
	ngPoint := len(diff)
	for _, text := range diff {
		offset := 0

		if _, ok := h.servConfig.DefaultAllowWords[text]; ok {
			offset -= 1
		}

		ngPoint += offset
	}
	return ngPoint
}

func (h *GroupMessageHandler) calcNGRate(ngPoint int, originLength int) int {
	ret := float32(ngPoint) / float32(originLength) * 100
	return int(ret)
}

func (h *GroupMessageHandler) refreshChatInfo(chatId int64) error {
	adminlist, err := h.bot.AdminsOf(&tele.Chat{ID: chatId})
	if err != nil {
		return err
	}

	storeMembers := make([]data.ChatMember, 0, len(adminlist))
	for _, item := range adminlist {
		storeMembers = append(storeMembers, data.ChatMember{
			Id:   item.User.ID,
			Name: item.User.FirstName + item.User.LastName,
		})
	}

	h.chatInfoStore.Update(chatId, func(info *data.ChatInfo) error {
		info.UpdateAdministrators(storeMembers)
		return nil
	})

	h.chatInfoStore.Save()
	return nil
}
