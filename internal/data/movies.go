package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/3WDeveloper-GM/json-app/internal/validator"
	"github.com/lib/pq"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

type MovieModel struct {
	DB *sql.DB
}

func ValidateMovie(v *validator.Validator, movie *Movie) {

	var messageNotProvided = "must be provided"

	v.Check(len(movie.Genres) > 0, "genres", messageNotProvided)
	v.Check(movie.Title != "", "title", messageNotProvided)
	v.Check(movie.Year != 0, "year", messageNotProvided)
	v.Check(movie.Runtime != 0, "runtime", messageNotProvided)

	var messageMore = "must be more than"
	var messageLess = "must be less than"

	var minimumGenreAmount = 1
	var maxGenreAmount = 5

	v.Check(len(movie.Genres) >= minimumGenreAmount, "genres", fmt.Sprintf(messageMore+" %v genre, add a genre", minimumGenreAmount))
	v.Check(len(movie.Genres) <= maxGenreAmount, "genres", fmt.Sprintf(messageLess+"%v genres, remove a genre", maxGenreAmount))

	var minimimMovieRuntime = 0

	v.Check(movie.Runtime > 0, "runtime", fmt.Sprintf(messageMore+" %d minutes, input a positive integer", minimimMovieRuntime))

	var maximumCharacterAmount = 500

	v.Check(len(movie.Title) < maximumCharacterAmount, "title", fmt.Sprintf(messageLess+" it must be less than %d bytes in length", maximumCharacterAmount))

	var presentYear = int32(time.Now().Year())

	v.Check(movie.Year <= presentYear, "creation_date", fmt.Sprintf(messageLess+" it must have a creation date not set in the future, the present year is %d", presentYear))

	var firstFilmYear = int32(1_888)

	v.Check(movie.Year >= firstFilmYear, "creation_date", fmt.Sprintf(messageMore+" %d, the first film dates from %d", firstFilmYear, firstFilmYear))

	var messageUniqueValues = "must have non-repeating genre tags"
	v.Check(validator.Unique(movie.Genres), "genres", messageUniqueValues)
}

func (m MovieModel) Insert(movie *Movie) error {

	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
		`

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id,created_at,title,year,runtime,genres,version
		FROM movies
		WHERE id = $1
		`

	var movie Movie

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {

	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		ORDER BY id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	movies := []*Movie{}

	for rows.Next() {

		var movie Movie

		err = rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}
		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres =$4, version = version+1
		WHERE id = $5 AND version = $6
		RETURNING version
	`

	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	query := `
		DELETE FROM movies
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	nRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if nRows == 0 {
		return ErrRecordNotFound
	}

	return nil

}
