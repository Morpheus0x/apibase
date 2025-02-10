package helper

import (
	"fmt"
	"strconv"
	"strings"
)

func PercentageToFloat32(toConvert string) (float32, error) {

	splitMargin := strings.Split(toConvert, "%")
	if len(splitMargin) != 2 || splitMargin[1] != "" {
		// log.Logf(log.LevelCritical, "unable to parse token_cookie_expiry_margin percentage: %s, assuming default %s", ac.TokenCookieExpiryMargin, TOKEN_COOKIE_EXPIRY_MARGIN)
		// ac.tokenCookieExpiryMargin = TOKEN_COOKIE_EXPIRY_MARGIN * 0.01
		return 0, fmt.Errorf("unable to parse string, no valid percentage")
	}
	parsedMargin, err := strconv.ParseFloat(splitMargin[0], 32)
	if err != nil {
		// log.Logf(log.LevelCritical, "unable to parse token_cookie_expiry_margin percentage: %s, assuming default %s", ac.TokenCookieExpiryMargin, TOKEN_COOKIE_EXPIRY_MARGIN)
		// ac.tokenCookieExpiryMargin = TOKEN_COOKIE_EXPIRY_MARGIN
		return 0, fmt.Errorf("unable to parse float: %v", err)
	}
	return float32(parsedMargin) * 0.01, nil
}
