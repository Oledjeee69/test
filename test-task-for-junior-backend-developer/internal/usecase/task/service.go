package task

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Recurrence:  normalized.Recurrence,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Recurrence:  normalized.Recurrence,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) Generate(ctx context.Context, id int64, input GenerateInput) ([]taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	from := normalizeToDate(input.FromDate)
	to := normalizeToDate(input.ToDate)
	if from.IsZero() || to.IsZero() {
		return nil, fmt.Errorf("%w: from_date and to_date are required", ErrInvalidInput)
	}
	if to.Before(from) {
		return nil, fmt.Errorf("%w: to_date must be greater or equal to from_date", ErrInvalidInput)
	}

	template, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if template.Recurrence == nil {
		return nil, fmt.Errorf("%w: task has no recurrence settings", ErrInvalidInput)
	}

	occurrences := buildOccurrences(template.Recurrence, from, to)
	generated := make([]taskdomain.Task, 0, len(occurrences))

	for _, day := range occurrences {
		scheduledFor := day
		sourceTaskID := template.ID
		now := s.now()

		model := &taskdomain.Task{
			Title:        template.Title,
			Description:  template.Description,
			Status:       taskdomain.StatusNew,
			SourceTaskID: &sourceTaskID,
			ScheduledFor: &scheduledFor,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		created, err := s.repo.Create(ctx, model)
		if err != nil {
			return nil, err
		}

		generated = append(generated, *created)
	}

	return generated, nil
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	recurrence, err := normalizeRecurrence(input.Recurrence)
	if err != nil {
		return CreateInput{}, err
	}
	input.Recurrence = recurrence

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	recurrence, err := normalizeRecurrence(input.Recurrence)
	if err != nil {
		return UpdateInput{}, err
	}
	input.Recurrence = recurrence

	return input, nil
}

func normalizeRecurrence(recurrence *taskdomain.Recurrence) (*taskdomain.Recurrence, error) {
	if recurrence == nil {
		return nil, nil
	}

	if !recurrence.Type.Valid() {
		return nil, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	switch recurrence.Type {
	case taskdomain.RecurrenceDaily:
		if recurrence.IntervalDays <= 0 {
			return nil, fmt.Errorf("%w: interval_days must be positive", ErrInvalidInput)
		}
		recurrence.MonthlyDay = 0
		recurrence.SpecificDates = nil
	case taskdomain.RecurrenceMonthlyDay:
		if recurrence.MonthlyDay < 1 || recurrence.MonthlyDay > 30 {
			return nil, fmt.Errorf("%w: monthly_day must be in range 1..30", ErrInvalidInput)
		}
		recurrence.IntervalDays = 0
		recurrence.SpecificDates = nil
	case taskdomain.RecurrenceSpecificDates:
		if len(recurrence.SpecificDates) == 0 {
			return nil, fmt.Errorf("%w: specific_dates must not be empty", ErrInvalidInput)
		}
		recurrence.IntervalDays = 0
		recurrence.MonthlyDay = 0

		unique := make(map[string]time.Time, len(recurrence.SpecificDates))
		for _, date := range recurrence.SpecificDates {
			normalized := normalizeToDate(date)
			if normalized.IsZero() {
				return nil, fmt.Errorf("%w: specific_dates contains invalid date", ErrInvalidInput)
			}
			unique[normalized.Format("2006-01-02")] = normalized
		}
		recurrence.SpecificDates = make([]time.Time, 0, len(unique))
		for _, date := range unique {
			recurrence.SpecificDates = append(recurrence.SpecificDates, date)
		}
		sort.Slice(recurrence.SpecificDates, func(i, j int) bool {
			return recurrence.SpecificDates[i].Before(recurrence.SpecificDates[j])
		})
	case taskdomain.RecurrenceEvenDays, taskdomain.RecurrenceOddDays:
		recurrence.IntervalDays = 0
		recurrence.MonthlyDay = 0
		recurrence.SpecificDates = nil
	}

	return recurrence, nil
}

func normalizeToDate(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	return time.Date(t.UTC().Year(), t.UTC().Month(), t.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func buildOccurrences(recurrence *taskdomain.Recurrence, from, to time.Time) []time.Time {
	result := make([]time.Time, 0)

	switch recurrence.Type {
	case taskdomain.RecurrenceDaily:
		step := recurrence.IntervalDays
		for day, index := from, 0; !day.After(to); day, index = day.AddDate(0, 0, 1), index+1 {
			if index%step == 0 {
				result = append(result, day)
			}
		}
	case taskdomain.RecurrenceMonthlyDay:
		for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
			if day.Day() == recurrence.MonthlyDay {
				result = append(result, day)
			}
		}
	case taskdomain.RecurrenceSpecificDates:
		for _, date := range recurrence.SpecificDates {
			if (date.Equal(from) || date.After(from)) && (date.Equal(to) || date.Before(to)) {
				result = append(result, date)
			}
		}
	case taskdomain.RecurrenceEvenDays:
		for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
			if day.Day()%2 == 0 {
				result = append(result, day)
			}
		}
	case taskdomain.RecurrenceOddDays:
		for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
			if day.Day()%2 != 0 {
				result = append(result, day)
			}
		}
	}

	return result
}
