package migration

import (
	common "MoneyHook/MoneyHook-API/common"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

const sampleUserID = common.DevelopmentUserID

type sampleSubCategoryDefinition struct {
	CategoryID uint64
	Name       string
}

var sampleCustomSubCategories = []sampleSubCategoryDefinition{
	{CategoryID: 2, Name: "ラーメン巡り"}, {CategoryID: 2, Name: "寿司巡り"}, {CategoryID: 2, Name: "カフェ巡り"},
	{CategoryID: 7, Name: "Amazon Prime"}, {CategoryID: 7, Name: "Youtube Premium"}, {CategoryID: 9, Name: "自転車用品"}, {CategoryID: 22, Name: "DIY"},
}

var sampleHiddenSubCategories = []sampleSubCategoryDefinition{
	{CategoryID: 5, Name: "スポーツ用品"}, {CategoryID: 8, Name: "テレビ"}, {CategoryID: 11, Name: "ゲームセンター"},
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
	referenceTime = referenceTime.In(location)
	result := sampleSeedResult{Status: "created", BaseMonth: referenceTime.Format("2006-01")}

	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user, err := findExistingSampleUser(tx)
		if err != nil {
			return err
		}
		if user == nil {
			user = &userSchema{UserID: sampleUserID}
			if err := tx.Create(user).Error; err != nil {
				return fmt.Errorf("create sample user: %w", err)
			}
		} else {
			result.Status = "refreshed"
			if err := clearSampleUserData(tx, user.UserNo); err != nil {
				return err
			}
			if err := resetSampleUserSettings(tx, user.UserNo); err != nil {
				return err
			}
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
		if err := createSampleBudgets(tx, user.UserNo, referenceTime); err != nil {
			return err
		}
		if err := createSampleMonthlyTransactions(tx, user.UserNo, masterIDs, customIDs, paymentIDs); err != nil {
			return err
		}
		transactions := buildSampleTransactions(user.UserNo, referenceTime, masterIDs, customIDs, paymentIDs)
		if err := tx.Create(&transactions).Error; err != nil {
			return fmt.Errorf("create sample transactions: %w", err)
		}
		result.TransactionCount = len(transactions)
		return nil
	})
	if err != nil {
		return err
	}

	log.Printf("event=schema_migration_seed status=%s user_no=%d transaction_count=%d base_month=%s", result.Status, result.UserNo, result.TransactionCount, result.BaseMonth)
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

func clearSampleUserData(db *gorm.DB, userNo uint64) error {
	for _, deletion := range []struct {
		name  string
		model any
	}{
		{name: "transactions", model: &transactionSchema{}},
		{name: "monthly transactions", model: &monthlyTransactionSchema{}},
		{name: "hidden sub-categories", model: &hiddenSubCategorySchema{}},
		{name: "payment resources", model: &paymentResourceSchema{}},
		{name: "sub-categories", model: &subCategorySchema{}},
		{name: "budgets", model: &budgetSchema{}},
	} {
		if err := db.Where("user_no = ?", userNo).Delete(deletion.model).Error; err != nil {
			return fmt.Errorf("delete sample %s: %w", deletion.name, err)
		}
	}
	return nil
}

func resetSampleUserSettings(db *gorm.DB, userNo uint64) error {
	if err := db.Model(&userSchema{}).Where("user_no = ?", userNo).Updates(map[string]any{"accent_color": "blue", "theme_mode": "system", "chart_palette": "default"}).Error; err != nil {
		return fmt.Errorf("reset sample user settings: %w", err)
	}
	return nil
}

func sampleSubCategoryKey(categoryID uint64, name string) string {
	return fmt.Sprintf("%d:%s", categoryID, name)
}

func requiredSampleMasterSubCategories() []sampleSubCategoryDefinition {
	return []sampleSubCategoryDefinition{
		{CategoryID: 1, Name: "なし"}, {CategoryID: 2, Name: "レストラン"}, {CategoryID: 3, Name: "なし"}, {CategoryID: 4, Name: "なし"},
		{CategoryID: 5, Name: "スポーツ用品"}, {CategoryID: 6, Name: "なし"}, {CategoryID: 7, Name: "なし"}, {CategoryID: 8, Name: "映画"}, {CategoryID: 8, Name: "テレビ"},
		{CategoryID: 9, Name: "なし"}, {CategoryID: 10, Name: "なし"}, {CategoryID: 11, Name: "ゲームセンター"}, {CategoryID: 13, Name: "なし"},
		{CategoryID: 15, Name: "ジム・フィットネス"}, {CategoryID: 15, Name: "病院"}, {CategoryID: 20, Name: "電気"}, {CategoryID: 20, Name: "水道"},
		{CategoryID: 20, Name: "ガス"}, {CategoryID: 21, Name: "携帯電話"}, {CategoryID: 22, Name: "家賃"}, {CategoryID: 27, Name: "なし"},
		{CategoryID: 27, Name: "ボーナス"}, {CategoryID: 28, Name: "利子所得"}, {CategoryID: 29, Name: "なし"},
	}
}

func resolveMasterSubCategories(db *gorm.DB, definitions []sampleSubCategoryDefinition) (map[string]uint64, error) {
	ids := make(map[string]uint64, len(definitions))
	for _, definition := range definitions {
		var rows []subCategorySchema
		if err := db.Where("user_no = ? AND category_id = ? AND sub_category_name = ?", systemUserNo, definition.CategoryID, definition.Name).Find(&rows).Error; err != nil {
			return nil, err
		}
		if len(rows) != 1 {
			return nil, fmt.Errorf("master sub-category user_no=%d category_id=%d name=%q matched %d rows", systemUserNo, definition.CategoryID, definition.Name, len(rows))
		}
		ids[sampleSubCategoryKey(definition.CategoryID, definition.Name)] = rows[0].SubCategoryID
	}
	return ids, nil
}

func createSampleSubCategories(db *gorm.DB, userNo uint64) (map[string]uint64, error) {
	ids := make(map[string]uint64, len(sampleCustomSubCategories))
	for _, definition := range sampleCustomSubCategories {
		row := subCategorySchema{UserNo: userNo, CategoryID: definition.CategoryID, SubCategoryName: definition.Name}
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
		rows = append(rows, hiddenSubCategorySchema{UserNo: userNo, SubCategoryID: masterIDs[sampleSubCategoryKey(definition.CategoryID, definition.Name)]})
	}
	if err := db.Create(&rows).Error; err != nil {
		return fmt.Errorf("create sample hidden sub-categories: %w", err)
	}
	return nil
}

func createSamplePaymentResources(db *gorm.DB, userNo uint64) (map[string]uint64, error) {
	paymentDate, closingDate := int32(27), int32(31)
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
	if name == "" {
		return nil
	}
	id := ids[name]
	return &id
}

type sampleMonthDefinition struct {
	BudgetAmount int64
	Transactions []sampleTransactionDefinition
}

type sampleTransactionDefinition struct {
	Name              string
	Amount            int64
	Day               int
	Time              string
	CategoryID        uint64
	SubCategoryName   string
	CustomSubCategory bool
	Fixed             bool
	PaymentName       string
}

var sampleMonthDefinitions = []sampleMonthDefinition{
	{BudgetAmount: 240000, Transactions: []sampleTransactionDefinition{
		{Name: "給与", Amount: 248000, Day: 25, Time: "09:00:00", CategoryID: 27, SubCategoryName: "なし", Fixed: true, PaymentName: "現金"},
		{Name: "家賃", Amount: -78550, Day: 27, Time: "08:00:00", CategoryID: 22, SubCategoryName: "家賃", Fixed: true, PaymentName: "楽天カード"},
		{Name: "電気", Amount: -4200, Day: 13, Time: "07:30:00", CategoryID: 20, SubCategoryName: "電気", Fixed: true, PaymentName: "楽天カード"},
		{Name: "通信", Amount: -3200, Day: 13, Time: "07:35:00", CategoryID: 21, SubCategoryName: "携帯電話", Fixed: true, PaymentName: "楽天カード"},
		{Name: "スーパー", Amount: -8420, Day: 5, Time: "18:10:00", CategoryID: 1, SubCategoryName: "なし", PaymentName: "現金"},
		{Name: "ラーメン", Amount: -1180, Day: 6, Time: "19:00:00", CategoryID: 2, SubCategoryName: "ラーメン巡り", CustomSubCategory: true, PaymentName: "PayPay"},
		{Name: "コンビニ", Amount: -630, Day: 8, Time: "12:10:00", CategoryID: 3, SubCategoryName: "なし", PaymentName: "PayPay"},
		{Name: "通勤", Amount: -2460, Day: 15, Time: "08:30:00", CategoryID: 13, SubCategoryName: "なし", PaymentName: "楽天カード"},
		{Name: "映画", Amount: -2000, Day: 22, Time: "16:00:00", CategoryID: 8, SubCategoryName: "映画", PaymentName: "楽天カード"},
		{Name: "投資積立", Amount: -10000, Day: 28, CategoryID: 29, SubCategoryName: "なし"},
		{Name: "配当", Amount: 305, Day: 25, Time: "10:00:00", CategoryID: 28, SubCategoryName: "利子所得", PaymentName: "現金"},
		{Name: "医療費", Amount: -1800, Day: 24, Time: "16:00:00", CategoryID: 15, SubCategoryName: "病院"},
	}},
	{BudgetAmount: 240000, Transactions: []sampleTransactionDefinition{
		{Name: "給与", Amount: 250000, Day: 25, Time: "09:00:00", CategoryID: 27, SubCategoryName: "なし", Fixed: true, PaymentName: "現金"},
		{Name: "家賃", Amount: -78550, Day: 27, Time: "08:00:00", CategoryID: 22, SubCategoryName: "家賃", Fixed: true, PaymentName: "楽天カード"},
		{Name: "水道", Amount: -2100, Day: 12, CategoryID: 20, SubCategoryName: "水道", Fixed: true, PaymentName: "楽天カード"},
		{Name: "通信", Amount: -3200, Day: 13, Time: "07:35:00", CategoryID: 21, SubCategoryName: "携帯電話", Fixed: true, PaymentName: "楽天カード"},
		{Name: "スーパー", Amount: -7600, Day: 4, Time: "18:20:00", CategoryID: 1, SubCategoryName: "なし", PaymentName: "現金"},
		{Name: "カフェ", Amount: -680, Day: 7, Time: "15:00:00", CategoryID: 2, SubCategoryName: "カフェ巡り", CustomSubCategory: true, PaymentName: "PayPay"},
		{Name: "寿司", Amount: -3800, Day: 11, Time: "12:30:00", CategoryID: 2, SubCategoryName: "寿司巡り", CustomSubCategory: true, PaymentName: "楽天カード"},
		{Name: "日用品", Amount: -2350, Day: 14, Time: "17:10:00", CategoryID: 4, SubCategoryName: "なし", PaymentName: "楽天カード"},
		{Name: "自転車用品", Amount: -4800, Day: 18, Time: "11:00:00", CategoryID: 9, SubCategoryName: "自転車用品", CustomSubCategory: true, PaymentName: "楽天カード"},
		{Name: "動画サービス", Amount: -1280, Day: 20, Time: "20:00:00", CategoryID: 7, SubCategoryName: "Youtube Premium", CustomSubCategory: true, Fixed: true, PaymentName: "楽天カード"},
		{Name: "スポーツ用品", Amount: -6500, Day: 21, Time: "11:30:00", CategoryID: 5, SubCategoryName: "スポーツ用品", PaymentName: "楽天カード"},
		{Name: "投資積立", Amount: -10000, Day: 28, CategoryID: 29, SubCategoryName: "なし"},
		{Name: "配当", Amount: 280, Day: 25, Time: "10:00:00", CategoryID: 28, SubCategoryName: "利子所得", PaymentName: "現金"},
		{Name: "医療費", Amount: -2600, Day: 26, Time: "10:30:00", CategoryID: 15, SubCategoryName: "病院"},
	}},
	{BudgetAmount: 245000, Transactions: []sampleTransactionDefinition{
		{Name: "給与", Amount: 250000, Day: 25, Time: "09:00:00", CategoryID: 27, SubCategoryName: "なし", Fixed: true, PaymentName: "現金"},
		{Name: "家賃", Amount: -78550, Day: 27, Time: "08:00:00", CategoryID: 22, SubCategoryName: "家賃", Fixed: true, PaymentName: "楽天カード"},
		{Name: "ガス", Amount: -3900, Day: 10, Time: "07:20:00", CategoryID: 20, SubCategoryName: "ガス", Fixed: true, PaymentName: "楽天カード"},
		{Name: "電気", Amount: -5100, Day: 13, Time: "07:30:00", CategoryID: 20, SubCategoryName: "電気", Fixed: true, PaymentName: "楽天カード"},
		{Name: "スーパー", Amount: -9100, Day: 5, Time: "18:10:00", CategoryID: 1, SubCategoryName: "なし", PaymentName: "現金"},
		{Name: "コンビニ", Amount: -720, Day: 9, Time: "12:20:00", CategoryID: 3, SubCategoryName: "なし", PaymentName: "PayPay"},
		{Name: "外食", Amount: -2600, Day: 12, Time: "19:30:00", CategoryID: 2, SubCategoryName: "レストラン", PaymentName: "PayPay"},
		{Name: "書籍", Amount: -2400, Day: 16, Time: "14:00:00", CategoryID: 8, SubCategoryName: "映画", PaymentName: "楽天カード"},
		{Name: "DIY", Amount: -7200, Day: 19, Time: "13:00:00", CategoryID: 22, SubCategoryName: "DIY", CustomSubCategory: true, PaymentName: "楽天カード"},
		{Name: "テレビ", Amount: -1500, Day: 23, CategoryID: 8, SubCategoryName: "テレビ", Fixed: true, PaymentName: "楽天カード"},
		{Name: "投資積立", Amount: -10000, Day: 28, CategoryID: 29, SubCategoryName: "なし"},
		{Name: "配当", Amount: 410, Day: 25, Time: "10:00:00", CategoryID: 28, SubCategoryName: "利子所得", PaymentName: "現金"},
		{Name: "ゲームセンター", Amount: -1800, Day: 24, Time: "21:00:00", CategoryID: 11, SubCategoryName: "ゲームセンター", PaymentName: "PayPay"},
	}},
	{BudgetAmount: 250000, Transactions: []sampleTransactionDefinition{
		{Name: "給与", Amount: 252000, Day: 25, Time: "09:00:00", CategoryID: 27, SubCategoryName: "なし", Fixed: true, PaymentName: "現金"},
		{Name: "家賃", Amount: -78550, Day: 27, Time: "08:00:00", CategoryID: 22, SubCategoryName: "家賃", Fixed: true, PaymentName: "楽天カード"},
		{Name: "水道", Amount: -1900, Day: 11, CategoryID: 20, SubCategoryName: "水道", Fixed: true, PaymentName: "楽天カード"},
		{Name: "通信", Amount: -3200, Day: 13, Time: "07:35:00", CategoryID: 21, SubCategoryName: "携帯電話", Fixed: true, PaymentName: "楽天カード"},
		{Name: "スーパー", Amount: -8150, Day: 3, Time: "18:10:00", CategoryID: 1, SubCategoryName: "なし", PaymentName: "現金"},
		{Name: "カフェ", Amount: -550, Day: 6, Time: "15:20:00", CategoryID: 2, SubCategoryName: "カフェ巡り", CustomSubCategory: true, PaymentName: "PayPay"},
		{Name: "通勤", Amount: -2850, Day: 8, Time: "08:30:00", CategoryID: 13, SubCategoryName: "なし", PaymentName: "楽天カード"},
		{Name: "洋服", Amount: -9200, Day: 14, Time: "14:30:00", CategoryID: 6, SubCategoryName: "なし", PaymentName: "楽天カード"},
		{Name: "旅行", Amount: -18500, Day: 17, Time: "09:30:00", CategoryID: 10, SubCategoryName: "なし", PaymentName: "楽天カード"},
		{Name: "動画サービス", Amount: -1280, Day: 20, Time: "20:00:00", CategoryID: 7, SubCategoryName: "Youtube Premium", CustomSubCategory: true, Fixed: true, PaymentName: "楽天カード"},
		{Name: "ジム", Amount: -5000, Day: 21, Time: "18:00:00", CategoryID: 15, SubCategoryName: "ジム・フィットネス", Fixed: true, PaymentName: "PayPay"},
		{Name: "投資積立", Amount: -10500, Day: 28, CategoryID: 29, SubCategoryName: "なし"},
		{Name: "配当", Amount: 330, Day: 25, Time: "10:00:00", CategoryID: 28, SubCategoryName: "利子所得", PaymentName: "現金"},
		{Name: "医療費", Amount: -1400, Day: 26, Time: "11:20:00", CategoryID: 15, SubCategoryName: "病院"},
		{Name: "Amazon Prime", Amount: -600, Day: 22, Time: "20:30:00", CategoryID: 7, SubCategoryName: "Amazon Prime", CustomSubCategory: true, Fixed: true, PaymentName: "楽天カード"},
	}},
	{BudgetAmount: 250000, Transactions: []sampleTransactionDefinition{
		{Name: "給与", Amount: 252000, Day: 25, Time: "09:00:00", CategoryID: 27, SubCategoryName: "なし", Fixed: true, PaymentName: "現金"},
		{Name: "家賃", Amount: -78550, Day: 27, Time: "08:00:00", CategoryID: 22, SubCategoryName: "家賃", Fixed: true, PaymentName: "楽天カード"},
		{Name: "電気", Amount: -5800, Day: 13, Time: "07:30:00", CategoryID: 20, SubCategoryName: "電気", Fixed: true, PaymentName: "楽天カード"},
		{Name: "ガス", Amount: -4300, Day: 14, Time: "07:20:00", CategoryID: 20, SubCategoryName: "ガス", Fixed: true, PaymentName: "楽天カード"},
		{Name: "スーパー", Amount: -10200, Day: 4, Time: "18:10:00", CategoryID: 1, SubCategoryName: "なし", PaymentName: "現金"},
		{Name: "ラーメン", Amount: -1350, Day: 7, Time: "19:10:00", CategoryID: 2, SubCategoryName: "ラーメン巡り", CustomSubCategory: true, PaymentName: "PayPay"},
		{Name: "コンビニ", Amount: -980, Day: 9, Time: "12:10:00", CategoryID: 3, SubCategoryName: "なし", PaymentName: "PayPay"},
		{Name: "外食", Amount: -3400, Day: 12, Time: "19:40:00", CategoryID: 2, SubCategoryName: "レストラン", PaymentName: "PayPay"},
		{Name: "趣味", Amount: -5200, Day: 16, Time: "13:20:00", CategoryID: 9, SubCategoryName: "なし", PaymentName: "楽天カード"},
		{Name: "スポーツ用品", Amount: -3200, Day: 18, Time: "11:00:00", CategoryID: 5, SubCategoryName: "スポーツ用品", PaymentName: "楽天カード"},
		{Name: "自転車用品", Amount: -3800, Day: 20, Time: "14:10:00", CategoryID: 9, SubCategoryName: "自転車用品", CustomSubCategory: true, PaymentName: "楽天カード"},
		{Name: "投資積立", Amount: -10500, Day: 28, CategoryID: 29, SubCategoryName: "なし"},
		{Name: "配当", Amount: 510, Day: 25, Time: "10:00:00", CategoryID: 28, SubCategoryName: "利子所得", PaymentName: "現金"},
		{Name: "医療費", Amount: -3200, Day: 24, Time: "09:40:00", CategoryID: 15, SubCategoryName: "病院"},
	}},
	{BudgetAmount: 260000, Transactions: []sampleTransactionDefinition{
		{Name: "給与", Amount: 255000, Day: 25, Time: "09:00:00", CategoryID: 27, SubCategoryName: "なし", Fixed: true, PaymentName: "現金"},
		{Name: "ボーナス", Amount: 50000, Day: 25, Time: "09:05:00", CategoryID: 27, SubCategoryName: "ボーナス", PaymentName: "現金"},
		{Name: "家賃", Amount: -78550, Day: 27, Time: "08:00:00", CategoryID: 22, SubCategoryName: "家賃", Fixed: true, PaymentName: "楽天カード"},
		{Name: "水道", Amount: -2200, Day: 12, CategoryID: 20, SubCategoryName: "水道", Fixed: true, PaymentName: "楽天カード"},
		{Name: "通信", Amount: -3200, Day: 13, Time: "07:35:00", CategoryID: 21, SubCategoryName: "携帯電話", Fixed: true, PaymentName: "楽天カード"},
		{Name: "スーパー", Amount: -8900, Day: 3, Time: "18:10:00", CategoryID: 1, SubCategoryName: "なし", PaymentName: "現金"},
		{Name: "カフェ", Amount: -720, Day: 5, Time: "15:10:00", CategoryID: 2, SubCategoryName: "カフェ巡り", CustomSubCategory: true, PaymentName: "PayPay"},
		{Name: "寿司", Amount: -4500, Day: 8, Time: "12:30:00", CategoryID: 2, SubCategoryName: "寿司巡り", CustomSubCategory: true, PaymentName: "楽天カード"},
		{Name: "日用品", Amount: -1800, Day: 10, Time: "17:20:00", CategoryID: 4, SubCategoryName: "なし", PaymentName: "楽天カード"},
		{Name: "旅行", Amount: -12600, Day: 14, Time: "10:00:00", CategoryID: 10, SubCategoryName: "なし", PaymentName: "楽天カード"},
		{Name: "テレビ", Amount: -1500, Day: 17, CategoryID: 8, SubCategoryName: "テレビ", Fixed: true, PaymentName: "楽天カード"},
		{Name: "動画サービス", Amount: -1280, Day: 20, Time: "20:00:00", CategoryID: 7, SubCategoryName: "Youtube Premium", CustomSubCategory: true, Fixed: true, PaymentName: "楽天カード"},
		{Name: "ジム", Amount: -5000, Day: 21, Time: "18:00:00", CategoryID: 15, SubCategoryName: "ジム・フィットネス", Fixed: true, PaymentName: "PayPay"},
		{Name: "投資積立", Amount: -11000, Day: 28, CategoryID: 29, SubCategoryName: "なし"},
		{Name: "配当", Amount: 620, Day: 25, Time: "10:00:00", CategoryID: 28, SubCategoryName: "利子所得", PaymentName: "現金"},
		{Name: "ゲームセンター", Amount: -2200, Day: 23, Time: "20:40:00", CategoryID: 11, SubCategoryName: "ゲームセンター", PaymentName: "PayPay"},
	}},
}

func createSampleBudgets(db *gorm.DB, userNo uint64, referenceTime time.Time) error {
	rows := buildSampleBudgets(userNo, referenceTime)
	if err := db.Create(&rows).Error; err != nil {
		return fmt.Errorf("create sample budgets: %w", err)
	}
	return nil
}

func buildSampleBudgets(userNo uint64, referenceTime time.Time) []budgetSchema {
	firstMonth := time.Date(referenceTime.Year(), referenceTime.Month(), 1, 0, 0, 0, 0, referenceTime.Location()).AddDate(0, -len(sampleMonthDefinitions)+1, 0)
	rows := make([]budgetSchema, 0, len(sampleMonthDefinitions))
	for monthIndex, definition := range sampleMonthDefinitions {
		rows = append(rows, budgetSchema{UserNo: userNo, EffectiveFrom: firstMonth.AddDate(0, monthIndex, 0).Format("2006-01-02"), MonthlyBudgetAmount: definition.BudgetAmount})
	}
	return rows
}

func createSampleMonthlyTransactions(db *gorm.DB, userNo uint64, masterIDs map[string]uint64, customIDs map[string]uint64, paymentIDs map[string]uint64) error {
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

func buildSampleTransactions(userNo uint64, referenceTime time.Time, masterIDs map[string]uint64, customIDs map[string]uint64, paymentIDs map[string]uint64) []transactionSchema {
	firstMonth := time.Date(referenceTime.Year(), referenceTime.Month(), 1, 0, 0, 0, 0, referenceTime.Location()).AddDate(0, -len(sampleMonthDefinitions)+1, 0)
	rows := make([]transactionSchema, 0, sampleTransactionCount())
	for monthIndex, monthDefinition := range sampleMonthDefinitions {
		month := firstMonth.AddDate(0, monthIndex, 0)
		for _, definition := range monthDefinition.Transactions {
			key := sampleSubCategoryKey(definition.CategoryID, definition.SubCategoryName)
			subCategoryID := masterIDs[key]
			if definition.CustomSubCategory {
				subCategoryID = customIDs[key]
			}
			var transactionTime *string
			if definition.Time != "" {
				value := definition.Time
				transactionTime = &value
			}
			rows = append(rows, transactionSchema{UserNo: userNo, TransactionName: definition.Name, TransactionAmount: definition.Amount, TransactionDate: time.Date(month.Year(), month.Month(), definition.Day, 0, 0, 0, 0, month.Location()).Format("2006-01-02"), TransactionTime: transactionTime, CategoryID: definition.CategoryID, SubCategoryID: subCategoryID, FixedFlg: definition.Fixed, PaymentID: paymentIDPointer(paymentIDs, definition.PaymentName)})
		}
	}
	return rows
}

func sampleTransactionCount() int {
	count := 0
	for _, definition := range sampleMonthDefinitions {
		count += len(definition.Transactions)
	}
	return count
}
