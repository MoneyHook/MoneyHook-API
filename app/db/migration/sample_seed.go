package migration

import (
	common "MoneyHook/MoneyHook-API/common"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

const (
	sampleUserID = common.DevelopmentUserID
)

type sampleSubCategoryDefinition struct {
	CategoryID uint64
	Name       string
}

var sampleCustomSubCategories = []sampleSubCategoryDefinition{
	{CategoryID: 2, Name: "ラーメン巡り"},
	{CategoryID: 2, Name: "寿司巡り"},
	{CategoryID: 2, Name: "カフェ巡り"},
	{CategoryID: 7, Name: "Amazon Prime"},
	{CategoryID: 7, Name: "Youtube Premium"},
	{CategoryID: 9, Name: "自転車用品"},
	{CategoryID: 22, Name: "DIY"},
}

var sampleHiddenSubCategories = []sampleSubCategoryDefinition{
	{CategoryID: 5, Name: "スポーツ用品"},
	{CategoryID: 8, Name: "テレビ"},
	{CategoryID: 11, Name: "ゲームセンター"},
}

type sampleSeedResult struct {
	Status           string
	UserNo           uint64
	TransactionCount int
	BaseMonth        string
}

func seedSampleData(ctx context.Context, db *gorm.DB, options Options) error {
	if !options.EnableSeedData {
		log.Printf("event=schema_migration_seed status=disabled")
		return nil
	}

	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return fmt.Errorf("load Asia/Tokyo timezone: %w", err)
	}
	referenceTime := options.SeedReferenceTime
	if referenceTime.IsZero() {
		referenceTime = time.Now()
	}
	baseMonth := referenceTime.In(location).Format("2006-01")
	result := sampleSeedResult{Status: "inserted", BaseMonth: baseMonth}

	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing, err := findExistingSampleUser(tx)
		if err != nil {
			return err
		}
		if existing != nil {
			result.Status = "already_exists"
			result.UserNo = existing.UserNo
			var transactionCount int64
			if err := tx.Model(&transactionSchema{}).Where("user_no = ?", existing.UserNo).Count(&transactionCount).Error; err != nil {
				return fmt.Errorf("count existing sample transactions: %w", err)
			}
			result.TransactionCount = int(transactionCount)
			return nil
		}

		user := userSchema{UserID: sampleUserID}
		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("create sample user: %w", err)
		}
		result.UserNo = user.UserNo
		masterIDs, err := resolveMasterSubCategories(tx, requiredSampleMasterSubCategories())
		if err != nil {
			return err
		}
		customIDs, err := createSampleSubCategories(tx, user.UserNo)
		if err != nil {
			return err
		}
		if err := createSampleHiddenSubCategories(tx, user.UserNo, masterIDs); err != nil {
			return err
		}
		paymentIDs, err := createSamplePaymentResources(tx, user.UserNo)
		if err != nil {
			return err
		}
		if err := createSampleMonthlyTransactions(tx, user.UserNo, masterIDs, customIDs, paymentIDs); err != nil {
			return err
		}
		transactions := buildSampleTransactions(user.UserNo, referenceTime.In(location), masterIDs, customIDs, paymentIDs)
		if err := tx.Create(&transactions).Error; err != nil {
			return fmt.Errorf("create sample transactions: %w", err)
		}
		result.TransactionCount = len(transactions)
		return nil
	})
	if err != nil {
		return err
	}

	log.Printf(
		"event=schema_migration_seed status=%s user_no=%d transaction_count=%d base_month=%s",
		result.Status,
		result.UserNo,
		result.TransactionCount,
		result.BaseMonth,
	)
	return nil
}

func findExistingSampleUser(db *gorm.DB) (*userSchema, error) {
	var users []userSchema
	if err := db.Where("user_id = ?", sampleUserID).Find(&users).Error; err != nil {
		return nil, err
	}
	if len(users) > 1 {
		return nil, fmt.Errorf("sample user_id=%q exists more than once", sampleUserID)
	}
	if len(users) == 1 {
		return &users[0], nil
	}
	return nil, nil
}

func sampleSubCategoryKey(categoryID uint64, name string) string {
	return fmt.Sprintf("%d:%s", categoryID, name)
}

func requiredSampleMasterSubCategories() []sampleSubCategoryDefinition {
	return []sampleSubCategoryDefinition{
		{CategoryID: 1, Name: "なし"},
		{CategoryID: 2, Name: "レストラン"},
		{CategoryID: 3, Name: "なし"},
		{CategoryID: 5, Name: "なし"},
		{CategoryID: 5, Name: "スポーツ用品"},
		{CategoryID: 7, Name: "なし"},
		{CategoryID: 8, Name: "映画"},
		{CategoryID: 8, Name: "テレビ"},
		{CategoryID: 11, Name: "ゲームセンター"},
		{CategoryID: 13, Name: "なし"},
		{CategoryID: 15, Name: "ジム・フィットネス"},
		{CategoryID: 20, Name: "電気"},
		{CategoryID: 21, Name: "携帯電話"},
		{CategoryID: 22, Name: "家賃"},
		{CategoryID: 27, Name: "なし"},
	}
}

func resolveMasterSubCategories(db *gorm.DB, definitions []sampleSubCategoryDefinition) (map[string]uint64, error) {
	ids := make(map[string]uint64, len(definitions))
	for _, definition := range definitions {
		var rows []subCategorySchema
		if err := db.Where(
			"user_no = ? AND category_id = ? AND sub_category_name = ?",
			systemUserNo,
			definition.CategoryID,
			definition.Name,
		).Find(&rows).Error; err != nil {
			return nil, err
		}
		if len(rows) != 1 {
			return nil, fmt.Errorf(
				"master sub-category user_no=%d category_id=%d name=%q matched %d rows",
				systemUserNo,
				definition.CategoryID,
				definition.Name,
				len(rows),
			)
		}
		ids[sampleSubCategoryKey(definition.CategoryID, definition.Name)] = rows[0].SubCategoryID
	}
	return ids, nil
}

func createSampleSubCategories(db *gorm.DB, userNo uint64) (map[string]uint64, error) {
	ids := make(map[string]uint64, len(sampleCustomSubCategories))
	for _, definition := range sampleCustomSubCategories {
		row := subCategorySchema{
			UserNo:          userNo,
			CategoryID:      definition.CategoryID,
			SubCategoryName: definition.Name,
		}
		if err := db.Create(&row).Error; err != nil {
			return nil, fmt.Errorf("create sample sub-category %q: %w", definition.Name, err)
		}
		ids[sampleSubCategoryKey(definition.CategoryID, definition.Name)] = row.SubCategoryID
	}
	return ids, nil
}

func createSampleHiddenSubCategories(db *gorm.DB, userNo uint64, masterIDs map[string]uint64) error {
	rows := make([]hiddenSubCategorySchema, 0, len(sampleHiddenSubCategories))
	for _, definition := range sampleHiddenSubCategories {
		rows = append(rows, hiddenSubCategorySchema{
			UserNo:        userNo,
			SubCategoryID: masterIDs[sampleSubCategoryKey(definition.CategoryID, definition.Name)],
		})
	}
	if err := db.Create(&rows).Error; err != nil {
		return fmt.Errorf("create sample hidden sub-categories: %w", err)
	}
	return nil
}

func createSamplePaymentResources(db *gorm.DB, userNo uint64) (map[string]uint64, error) {
	paymentDate := int32(27)
	closingDate := int32(31)
	rows := []paymentResourceSchema{
		{PaymentTypeID: 2, UserNo: userNo, PaymentName: "楽天カード", PaymentDate: &paymentDate, ClosingDate: &closingDate},
		{PaymentTypeID: 1, UserNo: userNo, PaymentName: "現金"},
		{PaymentTypeID: 3, UserNo: userNo, PaymentName: "PayPay"},
	}
	ids := make(map[string]uint64, len(rows))
	for index := range rows {
		if err := db.Create(&rows[index]).Error; err != nil {
			return nil, fmt.Errorf("create sample payment resource %q: %w", rows[index].PaymentName, err)
		}
		ids[rows[index].PaymentName] = rows[index].PaymentID
	}
	return ids, nil
}

func paymentIDPointer(ids map[string]uint64, name string) *uint64 {
	id := ids[name]
	return &id
}

func createSampleMonthlyTransactions(
	db *gorm.DB,
	userNo uint64,
	masterIDs map[string]uint64,
	customIDs map[string]uint64,
	paymentIDs map[string]uint64,
) error {
	rows := []monthlyTransactionSchema{
		{UserNo: userNo, MonthlyTransactionName: "家賃", MonthlyTransactionAmount: -78550, MonthlyTransactionDate: 27, CategoryID: 22, SubCategoryID: masterIDs[sampleSubCategoryKey(22, "家賃")], IncludeFlg: true, PaymentID: paymentIDPointer(paymentIDs, "楽天カード")},
		{UserNo: userNo, MonthlyTransactionName: "Youtube Premium", MonthlyTransactionAmount: -1280, MonthlyTransactionDate: 25, CategoryID: 7, SubCategoryID: customIDs[sampleSubCategoryKey(7, "Youtube Premium")], IncludeFlg: true, PaymentID: paymentIDPointer(paymentIDs, "楽天カード")},
		{UserNo: userNo, MonthlyTransactionName: "DisneyPlus", MonthlyTransactionAmount: -990, MonthlyTransactionDate: 25, CategoryID: 7, SubCategoryID: masterIDs[sampleSubCategoryKey(7, "なし")], IncludeFlg: true, PaymentID: paymentIDPointer(paymentIDs, "楽天カード")},
		{UserNo: userNo, MonthlyTransactionName: "給与", MonthlyTransactionAmount: 250000, MonthlyTransactionDate: 25, CategoryID: 27, SubCategoryID: masterIDs[sampleSubCategoryKey(27, "なし")], IncludeFlg: true, PaymentID: paymentIDPointer(paymentIDs, "現金")},
		{UserNo: userNo, MonthlyTransactionName: "Amazon Prime", MonthlyTransactionAmount: -600, MonthlyTransactionDate: 20, CategoryID: 7, SubCategoryID: customIDs[sampleSubCategoryKey(7, "Amazon Prime")], IncludeFlg: true, PaymentID: paymentIDPointer(paymentIDs, "楽天カード")},
		{UserNo: userNo, MonthlyTransactionName: "ジム", MonthlyTransactionAmount: -5000, MonthlyTransactionDate: 10, CategoryID: 15, SubCategoryID: masterIDs[sampleSubCategoryKey(15, "ジム・フィットネス")], IncludeFlg: false, PaymentID: paymentIDPointer(paymentIDs, "PayPay")},
	}
	if err := db.Create(&rows).Error; err != nil {
		return fmt.Errorf("create sample monthly transactions: %w", err)
	}
	return nil
}

type sampleTransactionDefinition struct {
	Name              string
	BaseAmount        int64
	MonthlyDelta      int64
	Day               int
	Time              string
	CategoryID        uint64
	SubCategoryName   string
	CustomSubCategory bool
	Fixed             bool
	PaymentName       string
}

var sampleTransactionDefinitions = []sampleTransactionDefinition{
	{Name: "給与", BaseAmount: 250000, MonthlyDelta: 1000, Day: 25, Time: "09:00:00", CategoryID: 27, SubCategoryName: "なし", Fixed: true, PaymentName: "現金"},
	{Name: "家賃", BaseAmount: -78550, Day: 27, Time: "08:00:00", CategoryID: 22, SubCategoryName: "家賃", Fixed: true, PaymentName: "楽天カード"},
	{Name: "動画サービス", BaseAmount: -1280, Day: 25, Time: "20:00:00", CategoryID: 7, SubCategoryName: "Youtube Premium", CustomSubCategory: true, Fixed: true, PaymentName: "楽天カード"},
	{Name: "電気", BaseAmount: -4500, MonthlyDelta: -200, Day: 13, Time: "07:30:00", CategoryID: 20, SubCategoryName: "電気", Fixed: true, PaymentName: "楽天カード"},
	{Name: "通信", BaseAmount: -3200, Day: 13, Time: "07:35:00", CategoryID: 21, SubCategoryName: "携帯電話", Fixed: true, PaymentName: "楽天カード"},
	{Name: "食費", BaseAmount: -8000, MonthlyDelta: -300, Day: 5, Time: "18:10:00", CategoryID: 1, SubCategoryName: "なし", PaymentName: "現金"},
	{Name: "コンビニ", BaseAmount: -650, MonthlyDelta: -10, Day: 8, Time: "12:10:00", CategoryID: 3, SubCategoryName: "なし", PaymentName: "PayPay"},
	{Name: "カフェ", BaseAmount: -550, MonthlyDelta: -20, Day: 10, Time: "15:00:00", CategoryID: 2, SubCategoryName: "カフェ巡り", CustomSubCategory: true, PaymentName: "PayPay"},
	{Name: "外食", BaseAmount: -1800, MonthlyDelta: -50, Day: 12, Time: "19:30:00", CategoryID: 2, SubCategoryName: "レストラン", PaymentName: "PayPay"},
	{Name: "交通", BaseAmount: -2500, MonthlyDelta: -100, Day: 15, Time: "08:30:00", CategoryID: 13, SubCategoryName: "なし", PaymentName: "楽天カード"},
	{Name: "買い物", BaseAmount: -4000, MonthlyDelta: -250, Day: 20, Time: "14:00:00", CategoryID: 5, SubCategoryName: "なし", PaymentName: "楽天カード"},
	{Name: "娯楽", BaseAmount: -2000, MonthlyDelta: -100, Day: 22, Time: "16:00:00", CategoryID: 8, SubCategoryName: "映画", PaymentName: "楽天カード"},
}

func buildSampleTransactions(
	userNo uint64,
	referenceTime time.Time,
	masterIDs map[string]uint64,
	customIDs map[string]uint64,
	paymentIDs map[string]uint64,
) []transactionSchema {
	firstMonth := time.Date(referenceTime.Year(), referenceTime.Month(), 1, 0, 0, 0, 0, referenceTime.Location()).AddDate(0, -5, 0)
	rows := make([]transactionSchema, 0, 6*len(sampleTransactionDefinitions))
	for monthIndex := 0; monthIndex < 6; monthIndex++ {
		month := firstMonth.AddDate(0, monthIndex, 0)
		for _, definition := range sampleTransactionDefinitions {
			key := sampleSubCategoryKey(definition.CategoryID, definition.SubCategoryName)
			subCategoryID := masterIDs[key]
			if definition.CustomSubCategory {
				subCategoryID = customIDs[key]
			}
			transactionTime := definition.Time
			rows = append(rows, transactionSchema{
				UserNo:            userNo,
				TransactionName:   definition.Name,
				TransactionAmount: definition.BaseAmount + int64(monthIndex)*definition.MonthlyDelta,
				TransactionDate:   time.Date(month.Year(), month.Month(), definition.Day, 0, 0, 0, 0, month.Location()).Format("2006-01-02"),
				TransactionTime:   &transactionTime,
				CategoryID:        definition.CategoryID,
				SubCategoryID:     subCategoryID,
				FixedFlg:          definition.Fixed,
				PaymentID:         paymentIDPointer(paymentIDs, definition.PaymentName),
			})
		}
	}
	return rows
}
