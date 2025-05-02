package model

type SubCategory struct {
	SubCategoryId   string
	SubCategoryName string
}

type SubCategoryModel struct {
	SubCategoryId   int64 `gorm:"primaryKey"`
	UserNo          string
	CategoryId      string
	SubCategoryName string
}

type EditSubCategoryModel struct {
	UserId        string
	SubCategoryId string
	IsEnable      bool
}
