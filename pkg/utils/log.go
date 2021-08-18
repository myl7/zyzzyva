package utils

import "log"

func InitLog() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC)
}
