package inbound

import "github.com/robfig/cron/v3"

type Task interface {
	cron.Job
}
