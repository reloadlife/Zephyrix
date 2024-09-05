package zephyrix

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

type JobInterface interface {
	Run()
	Spec() string
}

type ScheduleInterface interface {
	Run()
	Next(time.Time) time.Time
}

func (z *zephyrix) scheduleInvoke(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			z.crond.Start()
			Logger.Debug("Scheduler started")
			return nil
		},
		OnStop: func(context.Context) error {
			c := z.crond.Stop()
			Logger.Debug("Waiting for scheduler's tasks to end")
			// wait for the cron to stop
			<-c.Done()
			return nil
		},
	})
}

func (z *zephyrix) RegisterJob(jobs ...JobInterface) {
	for _, job := range jobs {
		if _, err := z.crond.AddJob(job.Spec(), job); err != nil {
			Logger.Error("Failed to register job: %v", err)
		}
	}
}

// RegisterSchedule registers one or more custom schedules with the scheduler.
func (z *zephyrix) RegisterSchedule(schedules ...ScheduleInterface) {
	for _, schedule := range schedules {
		_ = z.crond.Schedule(schedule, schedule)
	}
}

// RegisterCronFunc registers a function to be executed on a cron schedule.
func (z *zephyrix) RegisterCronFunc(spec string, f func()) {
	if _, err := z.crond.AddFunc(spec, f); err != nil {
		Logger.Error("Failed to register cron function: %v", err)
	}
}

func (z *zephyrix) RegisterExecuteLaterFunc(duration time.Duration, f func()) {
	scheduleSpec := z.durationToScheduleSpec(duration)
	_, err := z.crond.AddFunc(scheduleSpec, func() {
		z.executeFuncOnce(f)
	})

	if err != nil {
		Logger.Error("Failed to schedule function: %v", err)
	}
}

func (z *zephyrix) durationToScheduleSpec(duration time.Duration) string {
	seconds := int(duration.Seconds())
	return fmt.Sprintf("@every %ds", seconds)
}

func (z *zephyrix) executeFuncOnce(f func()) {
	defer func() {
		if r := recover(); r != nil {
			Logger.Error("Panic in scheduled function: %v", r)
		}
	}()

	f()

	// Remove the scheduled job after execution
	z.crond.Remove(cron.EntryID(z.crond.Entries()[len(z.crond.Entries())-1].ID))
}
