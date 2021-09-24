package util

import (
	"database/sql"
	"main/model"

	"github.com/lib/pq"
)

func GetTsundokus(DB *sql.DB, userID int) ([]model.Tsundoku, error) {
	var results []model.Tsundoku
	rows, err := DB.Query("select * from tsundokus where user_id = $1;", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var result model.Tsundoku
		nullAuthor := new(sql.NullString)
		nullURL := new(sql.NullString)
		nullDeadLine := new(pq.NullTime)
		nullRequiredTime := new(sql.NullString)
		nullCreatedAt := new(pq.NullTime)
		err := rows.Scan(&result.ID, &result.UserID, &result.Category, &result.Title, nullAuthor, nullURL, nullDeadLine, nullRequiredTime, nullCreatedAt)
		if err != nil {
			return nil, err
		}
		if nullAuthor.Valid {
			result.Author = nullAuthor.String
		}
		if nullURL.Valid {
			result.URL = nullURL.String
		}
		if nullDeadLine.Valid {
			result.Deadline = nullDeadLine.Time
		}
		if nullRequiredTime.Valid {
			result.RequiredTime = nullRequiredTime.String
		}
		if nullCreatedAt.Valid {
			result.CreatedAt = nullCreatedAt.Time
		}
		results = append(results, result)

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
