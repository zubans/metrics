package config

import "time"

type Config struct {
	AddressServer string
	SendInterval  time.Duration
	PollInterval  time.Duration
}
