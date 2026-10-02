package migration

import "gorm.io/gorm"

// Preserve the effective snapshot correction before removing the obsolete log.
// This is the only production code that still references the legacy table.
func removeHouseholdEntryHistory(db *gorm.DB) error {
	if !db.Migrator().HasTable("household_entry_revision") {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`UPDATE household_entry_data AS d
			SET corrected_payload = latest.after_data
			FROM (
				SELECT DISTINCT ON (r.entry_id) r.entry_id, r.after_data
				FROM household_entry_revision r
				JOIN household_entry e ON e.entry_id = r.entry_id
				WHERE r.action = 'corrected' AND e.kind = 'snapshot'
				ORDER BY r.entry_id, r.entry_version DESC
			) AS latest
			WHERE d.entry_id = latest.entry_id AND d.corrected_payload IS NULL`).Error; err != nil {
			return err
		}
		return tx.Exec("DROP TABLE household_entry_revision").Error
	})
}
