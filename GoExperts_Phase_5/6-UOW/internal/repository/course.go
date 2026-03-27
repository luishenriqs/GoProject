package repository

import (
	"context"
	"database/sql"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_5/6-UOW/internal/db"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_5/6-UOW/internal/entity"
)

type CourseRepositoryInterface interface {
	Insert(ctx context.Context, course entity.Course) error
}

type CourseRepository struct {
	DB      *sql.DB
	Queries *db.Queries
}

func NewCourseRepository(dtb *sql.DB) *CourseRepository {
	return &CourseRepository{
		DB:      dtb,
		Queries: db.New(dtb),
	}
}

func (r *CourseRepository) Insert(ctx context.Context, course entity.Course) error {
	return r.Queries.CreateCourse(ctx, db.CreateCourseParams{
		Name:       course.Name,
		CategoryID: int32(course.CategoryID),
	})
}
