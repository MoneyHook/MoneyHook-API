package request

import (
	"MoneyHook/MoneyHook-API/model"

	"github.com/labstack/echo/v4"
)

type EditSubCategoryRequest struct {
	SubCategoryId int  `json:"sub_category_id" validate:"required"`
	IsEnable      bool `json:"is_enable"  validate:"required"`
}

func (r *EditSubCategoryRequest) Bind(c echo.Context, u *model.EditSubCategoryModel) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	// TODO バリデーション
	// if err := c.Validate(r); err != nil {
	// 	return err
	// }

	u.SubCategoryId = r.SubCategoryId
	u.IsEnable = r.IsEnable

	return nil
}
