package tuner

import "github.com/rs/zerolog/log"

var BLACKLISTED_OPTS map[string]string = map[string]string{
	"-XX:MaxRAMFraction=1": "This option requests whole container memory, not leaving room for system buffers, which can lead to OOM errors.",
	// https://blog.csanchez.org/2017/05/31/running-a-jvm-in-a-container-without-getting-killed/
	"-XX:+UseCGroupMemoryLimitForHeap": "This option will only request 25% of the container memory limit, which is often too low for production workloads.",
}

func FilterBlacklisted(opts []string) []string {
	var filtered []string
	for _, opt := range opts {
		blacklisted := false
		for blacklistedOpt, reason := range BLACKLISTED_OPTS {
			if opt == blacklistedOpt {
				blacklisted = true
				log.Warn().Str("option", opt).Str("reason", reason).Msg("Filtered out blacklisted JVM option")
				break
			}
		}
		if !blacklisted {
			filtered = append(filtered, opt)
		}
	}
	return filtered
}
