package natssrv

import "errors"

// ErrEventMissingGroupID is returned when a group event is missing the group ID
var ErrEventMissingGroupID = errors.New("event missing group ID")
