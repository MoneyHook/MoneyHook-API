package model

type Category struct {
	CategoryId   int    `gorm:"category_id"`
	CategoryName string `gorm:"category_name"`
}

type CategoryList struct {
	CategoryList []Category
}

type CategoryWithSubCategoryList struct {
	CategoryWithSubCategoryList []CategoryWithSubCategory
}

type CategoryWithSubCategory struct {
	CategoryId      int
	CategoryName    string
	SubCategoryList []SubCategory `gorm:"foreignKey:SubCategoryId;references:CategoryId"`
}

type Tabler interface {
	TableName() string
}

func (SubCategory) TableName() string {
	return "sub_category"
}

type SubCategory struct {
	SubCategoryId   int
	SubCategoryName string
	// Enable          bool
}
