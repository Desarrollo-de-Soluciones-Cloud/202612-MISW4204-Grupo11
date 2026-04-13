package spaces_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/spaces"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/domain"
)

type periodRepoStub struct {
	created []*domain.AcademicPeriod
	period  *domain.AcademicPeriod
	err     error
	periods []domain.AcademicPeriod
	listErr error
}

func (p *periodRepoStub) FindByID(_ context.Context, id int64) (*domain.AcademicPeriod, error) {
	if p.period != nil && p.period.ID == id {
		return p.period, nil
	}
	if p.period != nil {
		return p.period, nil
	}
	return nil, domain.ErrPeriodoNoEncontrado
}

func (p *periodRepoStub) Create(_ context.Context, period *domain.AcademicPeriod) error {
	if p.err != nil {
		return p.err
	}
	period.ID = 1
	p.created = append(p.created, period)
	return nil
}

func (p *periodRepoStub) List(_ context.Context) ([]domain.AcademicPeriod, error) {
	if p.listErr != nil {
		return nil, p.listErr
	}
	if p.periods != nil {
		return p.periods, nil
	}
	return nil, nil
}

func (p *periodRepoStub) UpdateStatus(_ context.Context, _ int64, _ string) error {
	if p.period != nil {
		p.period.Status = "closed"
	}
	return nil
}

func TestAcademicPeriodService_CreatePeriod_OK(t *testing.T) {
	repo := &periodRepoStub{}
	svc := spaces.NewAcademicPeriodService(repo)
	start := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	p, err := svc.CreatePeriod(context.Background(), spaces.CreatePeriodInput{
		Code: "2026-10", StartDate: start, EndDate: end,
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Code != "2026-10" || p.Status != "active" {
		t.Fatalf("unexpected %+v", p)
	}
}

func TestAcademicPeriodService_CreatePeriod_InvalidDates(t *testing.T) {
	svc := spaces.NewAcademicPeriodService(&periodRepoStub{})
	start := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	_, err := svc.CreatePeriod(context.Background(), spaces.CreatePeriodInput{
		Code: "bad", StartDate: start, EndDate: end,
	})
	if !errors.Is(err, domain.ErrFechasCierreInvalidas) {
		t.Fatalf("want ErrFechasCierreInvalidas, got %v", err)
	}
}

func TestAcademicPeriodService_ClosePeriod_OK(t *testing.T) {
	repo := &periodRepoStub{
		period: &domain.AcademicPeriod{
			ID: 1, Code: "2026-1", Status: "active",
			StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	svc := spaces.NewAcademicPeriodService(repo)
	if err := svc.ClosePeriod(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if repo.period.Status != "closed" {
		t.Fatalf("status %q", repo.period.Status)
	}
}

func TestAcademicPeriodService_ClosePeriod_AlreadyClosed(t *testing.T) {
	repo := &periodRepoStub{
		period: &domain.AcademicPeriod{ID: 2, Status: "closed"},
	}
	svc := spaces.NewAcademicPeriodService(repo)
	err := svc.ClosePeriod(context.Background(), 2)
	if !errors.Is(err, domain.ErrPeriodoCerrado) {
		t.Fatalf("want ErrPeriodoCerrado, got %v", err)
	}
}

func TestAcademicPeriodService_ListPeriods_OK(t *testing.T) {
	repo := &periodRepoStub{periods: []domain.AcademicPeriod{{ID: 1, Code: "2026-1", Status: "active"}}}
	svc := spaces.NewAcademicPeriodService(repo)
	out, err := svc.ListPeriods(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Code != "2026-1" {
		t.Fatalf("unexpected %+v", out)
	}
}

func TestAcademicPeriodService_ListPeriods_RepositoryError(t *testing.T) {
	want := errors.New("list fail")
	repo := &periodRepoStub{listErr: want}
	svc := spaces.NewAcademicPeriodService(repo)
	_, err := svc.ListPeriods(context.Background())
	if !errors.Is(err, want) {
		t.Fatalf("want %v, got %v", want, err)
	}
}
