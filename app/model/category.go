package model

type Category struct {
	CategoryId   string
	CategoryName string
}

type CategoryWithSubCategory struct {
	CategoryId      string
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
	SubCategoryId   string
	SubCategoryName string
	Enable          bool
}
