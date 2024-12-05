package dal

import "go-api-testing/internal/dal/model"

func GetQueryModels() []interface{} {
	return []interface{}{
		&model.User{},
	}
}
