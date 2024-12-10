package log

func Demo(level ...Level) {
	loggerMutex.RLock()
	logRaw := *logger
	loggerMutex.RUnlock()
	log := &logRaw
	if len(level) > 0 {
		log.Level = level[0]
	}
	log.printWithLevel("Log Demo Debug", LevelDebug)
	log.printWithLevel("Log Demo Info", LevelInfo)
	log.printWithLevel("Log Demo Notice", LevelNotice)
	log.printWithLevel("Log Demo Warning", LevelWarning)
	log.printWithLevel("Log Demo Error", LevelError)
	log.printWithLevel("Log Demo Critical", LevelCritical)
}
