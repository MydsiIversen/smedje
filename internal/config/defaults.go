package config

// Defaults holds every configurable value with its built-in default.
var Defaults = map[string]string{
	"password.length":  "24",
	"password.charset": "full",

	"totp.issuer":  "Smedje",
	"totp.account": "user@example.com",
	"totp.digits":  "6",
	"totp.period":  "30",

	"tls.cn":   "localhost",
	"tls.days": "825",

	"snowflake.worker": "0",

	"mac.format": "colon",

	"bulk.max-count": "100000000",

	"bench.duration": "2s",
	"bench.warmup":   "500ms",
	"bench.repeat":   "1",
	"bench.cores":    "0",
}
