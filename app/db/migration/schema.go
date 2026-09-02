package migration

type userSchema struct {
	UserID       string `gorm:"column:user_id;type:varchar(128);not null;unique"`
	UserNo       uint64 `gorm:"column:user_no;primaryKey;autoIncrement"`
	AccentColor  string `gorm:"column:accent_color;type:varchar(16);not null;default:blue"`
	ThemeMode    string `gorm:"column:theme_mode;type:varchar(16);not null;default:system"`
	ChartPalette string `gorm:"column:chart_palette;type:varchar(16);not null;default:default"`
}

func (userSchema) TableName() string { return "users" }

type budgetSchema struct {
	UserNo              uint64 `gorm:"column:user_no;primaryKey"`
	EffectiveFrom       string `gorm:"column:effective_from;type:date;primaryKey"`
	MonthlyBudgetAmount int64  `gorm:"column:monthly_budget_amount;not null"`
}

func (budgetSchema) TableName() string { return "budget" }

type categorySchema struct {
	CategoryID   uint64 `gorm:"column:category_id;primaryKey;autoIncrement"`
	CategoryName string `gorm:"column:category_name;type:varchar(16);not null"`
	OrderNum     int32  `gorm:"column:order_num;not null"`
}

func (categorySchema) TableName() string { return "category" }

type subCategorySchema struct {
	SubCategoryID   uint64 `gorm:"column:sub_category_id;primaryKey;autoIncrement"`
	UserNo          uint64 `gorm:"column:user_no;not null"`
	CategoryID      uint64 `gorm:"column:category_id;not null"`
	SubCategoryName string `gorm:"column:sub_category_name;type:varchar(16);not null"`
}

func (subCategorySchema) TableName() string { return "sub_category" }

type hiddenSubCategorySchema struct {
	UserNo        uint64 `gorm:"column:user_no;not null"`
	SubCategoryID uint64 `gorm:"column:sub_category_id;not null"`
}

func (hiddenSubCategorySchema) TableName() string { return "hidden_sub_category" }

type paymentTypeSchema struct {
	PaymentTypeID     uint64 `gorm:"column:payment_type_id;primaryKey;autoIncrement"`
	PaymentTypeName   string `gorm:"column:payment_type_name;type:varchar(32);not null;unique"`
	IsPaymentDueLater bool   `gorm:"column:is_payment_due_later;not null"`
	OrderNum          int32  `gorm:"column:order_num;not null"`
}

func (paymentTypeSchema) TableName() string { return "payment_type" }

type paymentResourceSchema struct {
	PaymentID     uint64 `gorm:"column:payment_id;primaryKey;autoIncrement"`
	PaymentTypeID uint64 `gorm:"column:payment_type_id;not null;default:1"`
	UserNo        uint64 `gorm:"column:user_no;not null"`
	PaymentName   string `gorm:"column:payment_name;type:varchar(32);not null"`
	PaymentDate   *int32 `gorm:"column:payment_date"`
	ClosingDate   *int32 `gorm:"column:closing_date"`
}

func (paymentResourceSchema) TableName() string { return "payment_resource" }

type transactionSchema struct {
	TransactionID     uint64  `gorm:"column:transaction_id;primaryKey;autoIncrement"`
	UserNo            uint64  `gorm:"column:user_no;not null"`
	TransactionName   string  `gorm:"column:transaction_name;type:varchar(32);not null"`
	TransactionAmount int64   `gorm:"column:transaction_amount;not null"`
	TransactionDate   string  `gorm:"column:transaction_date;type:date;not null"`
	TransactionTime   *string `gorm:"column:transaction_time;type:time(0)"`
	CategoryID        uint64  `gorm:"column:category_id;not null"`
	SubCategoryID     uint64  `gorm:"column:sub_category_id;not null"`
	FixedFlg          bool    `gorm:"column:fixed_flg;not null"`
	PaymentID         *uint64 `gorm:"column:payment_id"`
}

func (transactionSchema) TableName() string { return "transaction" }

type monthlyTransactionSchema struct {
	MonthlyTransactionID     uint64  `gorm:"column:monthly_transaction_id;primaryKey;autoIncrement"`
	UserNo                   uint64  `gorm:"column:user_no;not null"`
	MonthlyTransactionName   string  `gorm:"column:monthly_transaction_name;type:varchar(32);not null"`
	MonthlyTransactionAmount int64   `gorm:"column:monthly_transaction_amount;not null"`
	MonthlyTransactionDate   int32   `gorm:"column:monthly_transaction_date;not null"`
	CategoryID               uint64  `gorm:"column:category_id;not null"`
	SubCategoryID            uint64  `gorm:"column:sub_category_id;not null"`
	IncludeFlg               bool    `gorm:"column:include_flg;not null"`
	PaymentID                *uint64 `gorm:"column:payment_id"`
}

func (monthlyTransactionSchema) TableName() string { return "monthly_transaction" }

var schemaModels = []any{
	&userSchema{},
	&budgetSchema{},
	&categorySchema{},
	&subCategorySchema{},
	&hiddenSubCategorySchema{},
	&paymentTypeSchema{},
	&paymentResourceSchema{},
	&transactionSchema{},
	&monthlyTransactionSchema{},
}

type uniqueRequirement struct {
	Name    string
	Table   string
	Columns []string
}

var uniqueRequirements = []uniqueRequirement{
	{Name: "uq_users_user_id", Table: "users", Columns: []string{"user_id"}},
	{Name: "uq_sub_category_user_category_name", Table: "sub_category", Columns: []string{"user_no", "category_id", "sub_category_name"}},
	{Name: "uq_hidden_sub_category_user_sub_category", Table: "hidden_sub_category", Columns: []string{"user_no", "sub_category_id"}},
	{Name: "uq_payment_type_name", Table: "payment_type", Columns: []string{"payment_type_name"}},
	{Name: "uq_payment_resource_user_name", Table: "payment_resource", Columns: []string{"user_no", "payment_name"}},
}

type foreignKeyRequirement struct {
	Name             string
	Table            string
	Column           string
	ReferencedTable  string
	ReferencedColumn string
}

var foreignKeyRequirements = []foreignKeyRequirement{
	{Name: "fk_budget_user", Table: "budget", Column: "user_no", ReferencedTable: "users", ReferencedColumn: "user_no"},
	{Name: "fk_sub_category_user", Table: "sub_category", Column: "user_no", ReferencedTable: "users", ReferencedColumn: "user_no"},
	{Name: "fk_sub_category_category", Table: "sub_category", Column: "category_id", ReferencedTable: "category", ReferencedColumn: "category_id"},
	{Name: "fk_hidden_sub_category_user", Table: "hidden_sub_category", Column: "user_no", ReferencedTable: "users", ReferencedColumn: "user_no"},
	{Name: "fk_hidden_sub_category_sub_category", Table: "hidden_sub_category", Column: "sub_category_id", ReferencedTable: "sub_category", ReferencedColumn: "sub_category_id"},
	{Name: "fk_payment_resource_payment_type", Table: "payment_resource", Column: "payment_type_id", ReferencedTable: "payment_type", ReferencedColumn: "payment_type_id"},
	{Name: "fk_payment_resource_user", Table: "payment_resource", Column: "user_no", ReferencedTable: "users", ReferencedColumn: "user_no"},
	{Name: "fk_transaction_user", Table: "transaction", Column: "user_no", ReferencedTable: "users", ReferencedColumn: "user_no"},
	{Name: "fk_transaction_category", Table: "transaction", Column: "category_id", ReferencedTable: "category", ReferencedColumn: "category_id"},
	{Name: "fk_transaction_sub_category", Table: "transaction", Column: "sub_category_id", ReferencedTable: "sub_category", ReferencedColumn: "sub_category_id"},
	{Name: "fk_transaction_payment_resource", Table: "transaction", Column: "payment_id", ReferencedTable: "payment_resource", ReferencedColumn: "payment_id"},
	{Name: "fk_monthly_transaction_user", Table: "monthly_transaction", Column: "user_no", ReferencedTable: "users", ReferencedColumn: "user_no"},
	{Name: "fk_monthly_transaction_category", Table: "monthly_transaction", Column: "category_id", ReferencedTable: "category", ReferencedColumn: "category_id"},
	{Name: "fk_monthly_transaction_sub_category", Table: "monthly_transaction", Column: "sub_category_id", ReferencedTable: "sub_category", ReferencedColumn: "sub_category_id"},
	{Name: "fk_monthly_transaction_payment_resource", Table: "monthly_transaction", Column: "payment_id", ReferencedTable: "payment_resource", ReferencedColumn: "payment_id"},
}

type sequenceRequirement struct {
	Table  string
	Column string
}

var sequenceRequirements = []sequenceRequirement{
	{Table: "users", Column: "user_no"},
	{Table: "category", Column: "category_id"},
	{Table: "sub_category", Column: "sub_category_id"},
	{Table: "payment_type", Column: "payment_type_id"},
	{Table: "payment_resource", Column: "payment_id"},
	{Table: "transaction", Column: "transaction_id"},
	{Table: "monthly_transaction", Column: "monthly_transaction_id"},
}
