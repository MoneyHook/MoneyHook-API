package model

type Category struct {
	CategoryId   int
	CategoryName string
}

type CategoryWithSubCategory struct {
	CategoryId      int
	CategoryName    string
	SubCategoryList []SubCategoryWithEnable `gorm:"foreignKey:SubCategoryId;references:CategoryId"`
}

type Tabler interface {
	TableName() string
}

func (SubCategoryWithEnable) TableName() string {
	return "sub_category"
}

type SubCategoryWithEnable struct {
	SubCategoryId   int
	SubCategoryName string
	Enable          bool
}
