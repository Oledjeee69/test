package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      Status      `json:"status"`
	Recurrence  *Recurrence `json:"recurrence,omitempty"`
	SourceTaskID *int64     `json:"source_task_id,omitempty"`
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthlyDay    RecurrenceType = "monthly_day"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceEvenDays      RecurrenceType = "even_days"
	RecurrenceOddDays       RecurrenceType = "odd_days"
)

type Recurrence struct {
	Type          RecurrenceType `json:"type"`
	IntervalDays  int            `json:"interval_days,omitempty"`
	MonthlyDay    int            `json:"monthly_day,omitempty"`
	SpecificDates []time.Time    `json:"specific_dates,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (t RecurrenceType) Valid() bool {
	switch t {
	case RecurrenceDaily, RecurrenceMonthlyDay, RecurrenceSpecificDates, RecurrenceEvenDays, RecurrenceOddDays:
		return true
	default:
		return false
	}
}
