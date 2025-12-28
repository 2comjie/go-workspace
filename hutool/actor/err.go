package actor

import "errors"

var ErrMsgChanFull error = errors.New("msg chan is full")
var ErrMsgActorClosed error = errors.New("actor is closed")
