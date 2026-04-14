package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  *recurrenceDTO    `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	Recurrence   *recurrenceDTO    `json:"recurrence,omitempty"`
	SourceTaskID *int64            `json:"source_task_id,omitempty"`
	ScheduledFor *time.Time        `json:"scheduled_for,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type recurrenceDTO struct {
	Type          taskdomain.RecurrenceType `json:"type"`
	IntervalDays  int                       `json:"interval_days,omitempty"`
	MonthlyDay    int                       `json:"monthly_day,omitempty"`
	SpecificDates []time.Time               `json:"specific_dates,omitempty"`
}

type generateTasksDTO struct {
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Recurrence:  newRecurrenceDTO(task.Recurrence),
		SourceTaskID: task.SourceTaskID,
		ScheduledFor: task.ScheduledFor,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func newRecurrenceDTO(recurrence *taskdomain.Recurrence) *recurrenceDTO {
	if recurrence == nil {
		return nil
	}
	return &recurrenceDTO{
		Type:          recurrence.Type,
		IntervalDays:  recurrence.IntervalDays,
		MonthlyDay:    recurrence.MonthlyDay,
		SpecificDates: recurrence.SpecificDates,
	}
}

func newRecurrenceDomain(recurrence *recurrenceDTO) *taskdomain.Recurrence {
	if recurrence == nil {
		return nil
	}
	return &taskdomain.Recurrence{
		Type:          recurrence.Type,
		IntervalDays:  recurrence.IntervalDays,
		MonthlyDay:    recurrence.MonthlyDay,
		SpecificDates: recurrence.SpecificDates,
	}
}
