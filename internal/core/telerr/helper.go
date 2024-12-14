package telerr

import tele "gopkg.in/telebot.v4"

func ErrIs(err error, teleErr *tele.Error) bool {
	nerr, ok := err.(*tele.Error)
	if !ok {
		return false
	}

	return nerr.Description == teleErr.Description
}
