package network

import (
	"strings"
)

//redisHasMovedError check redis has moved or ask
func redisHasMovedError(buf []byte) (moved bool, ask bool, addr string) {

	s := string(buf)
	if strings.HasPrefix(s, "-MOVED ") {
		moved = true
	} else if strings.HasPrefix(s, "-ASK ") {
		ask = true
	} else {
		return
	}

	ind := strings.LastIndex(s, " ")
	if ind == -1 {
		return false, false, ""
	}
	addr = strings.TrimSuffix(s[ind+1:], "\r\n")
	return
}
