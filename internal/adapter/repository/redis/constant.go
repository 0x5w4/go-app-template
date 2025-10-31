package redisrepository

const (
	KeyPatternUserLock       = "lock:user:%s"
	KeyPatternBlockIP        = "block:ip:%s"
	KeyPatternUserAttempts   = "attempts:user:%s"
	KeyPatternIPAttempts     = "attempts:ip:%s"
	KeyPatternBlockCountIP   = "blockcount:ip:%s"
	KeyPatternBlacklistToken = "blacklist:token:%s"
	KeyPatternResetPassword  = "reset:password:%s"
)
