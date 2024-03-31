package model

type SubCategory struct {
	SubCategoryId   int
	SubCategoryName string
}

type SubCategoryModel struct {
	SubCategoryId   int `gorm:"primaryKey"`
	UserNo          int
	CategoryId      int
	SubCategoryName string
}
