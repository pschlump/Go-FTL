package InMemoryCache

import (
	"net/http"
	"strconv"
	"strings"
)

const TenYears int64 = 60 * 60 * 24 * 365 * 10 // 10 years

// -----------------------------------------------------------------------------------------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------------------------------------------------------------------------------------
// Return if this item can be cached and if so for how long.
//
// Exampels of setting non-caching headers are:
//		www.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
//		www.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
//		www.Header().Set("Expires", "0")                                         // Proxies.
func HeadersAllowCaching(hdr http.Header, duration int) (seconds int64, rawSeconds int64, mayCache bool) {
	rawSeconds = TenYears // just about forever
	seconds = int64(duration)
	mayCache = true
	for ii, vv := range hdr {
		switch ii {
		case "Pragma":
			for _, ww := range vv {
				if ww == "no-cache" {
					return 0, 0, false
				}
			}
		case "Expires":
			seconds = TenYears
			for _, ww := range vv {
				if ww == "0" {
					return 0, 0, false
				} else {
					rawSeconds, err := strconv.ParseInt(ww, 10, 64)
					if err != nil {
						return 0, 0, false
					}
					if rawSeconds < seconds {
						seconds = rawSeconds
					}
				}
			}
		case "Cache-Control":
			// xyzzy304 if "must-revalidate" cache - then -- regenerate data - and if not changed send 304
			for _, ww := range vv {
				if strings.Contains(ww, "no-cache") || strings.Contains(ww, "no-store") || strings.Contains(ww, "must-revalidate") {
					return 0, 0, false
				}
			}
		}
	}
	return
}

/* vim: set noai ts=4 sw=4: */
