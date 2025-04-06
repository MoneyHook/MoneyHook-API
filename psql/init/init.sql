/* ============================================================
CREATE TABLES
============================================================ */
DROP TABLE IF EXISTS users,
user_token,
category,
sub_category,
hidden_sub_category,
transaction,
monthly_transaction,
payment_type,
payment_resource CASCADE;

CREATE TABLE
    users (
        user_id VARCHAR(64) NOT NULL,
        user_no BIGSERIAL PRIMARY KEY,
        email VARCHAR(128) UNIQUE,
        PASSWORD TEXT
    );

CREATE TABLE
    category (
        category_id BIGSERIAL PRIMARY KEY,
        category_name VARCHAR(16) NOT NULL,
        order_num INT NOT NULL
    );

CREATE TABLE
    user_token (
        user_no BIGINT NOT NULL,
        token VARCHAR(64) NOT NULL,
        CONSTRAINT fk_user_no FOREIGN KEY (user_no) REFERENCES users (user_no)
    );

CREATE TABLE
    sub_category (
        sub_category_id BIGSERIAL PRIMARY KEY,
        user_no BIGINT NOT NULL,
        category_id BIGINT NOT NULL,
        sub_category_name VARCHAR(16) NOT NULL,
        CONSTRAINT fk_user_no_sub FOREIGN KEY (user_no) REFERENCES users (user_no),
        CONSTRAINT fk_category_id FOREIGN KEY (category_id) REFERENCES category (category_id),
        UNIQUE (user_no, category_id, sub_category_name)
    );

CREATE TABLE
    hidden_sub_category (
        user_no BIGINT NOT NULL,
        sub_category_id BIGINT NOT NULL,
        CONSTRAINT fk_user_no_hidden FOREIGN KEY (user_no) REFERENCES users (user_no),
        CONSTRAINT fk_sub_category_id FOREIGN KEY (sub_category_id) REFERENCES sub_category (sub_category_id),
        UNIQUE (user_no, sub_category_id)
    );

CREATE TABLE
    payment_type (
        payment_type_id BIGSERIAL PRIMARY KEY,
        payment_type_name VARCHAR(32) NOT NULL,
        is_payment_due_later BOOLEAN NOT NULL,
        order_num INT NOT NULL,
        UNIQUE (payment_type_name)
    );

CREATE TABLE
    payment_resource (
        payment_id BIGSERIAL PRIMARY KEY,
        payment_type_id BIGINT NOT NULL DEFAULT 1,
        user_no BIGINT NOT NULL,
        payment_name VARCHAR(32) NOT NULL,
        payment_date INT,
        closing_date INT,
        CONSTRAINT fk_user_no_payment FOREIGN KEY (user_no) REFERENCES users (user_no),
        CONSTRAINT fk_payment_type_id FOREIGN KEY (payment_type_id) REFERENCES payment_type (payment_type_id),
        UNIQUE (user_no, payment_name)
    );

CREATE TABLE
    transaction (
        transaction_id BIGSERIAL PRIMARY KEY,
        user_no BIGINT NOT NULL,
        transaction_name VARCHAR(32) NOT NULL,
        transaction_amount BIGINT NOT NULL,
        transaction_date DATE NOT NULL,
        category_id BIGINT NOT NULL,
        sub_category_id BIGINT NOT NULL,
        fixed_flg BOOLEAN NOT NULL,
        payment_id BIGINT,
        CONSTRAINT fk_user_no_transaction FOREIGN KEY (user_no) REFERENCES users (user_no),
        CONSTRAINT fk_category_id_transaction FOREIGN KEY (category_id) REFERENCES category (category_id),
        CONSTRAINT fk_sub_category_id_transaction FOREIGN KEY (sub_category_id) REFERENCES sub_category (sub_category_id),
        CONSTRAINT fk_payment_id_transaction FOREIGN KEY (payment_id) REFERENCES payment_resource (payment_id)
    );

CREATE TABLE
    monthly_transaction (
        monthly_transaction_id BIGSERIAL PRIMARY KEY,
        user_no BIGINT NOT NULL,
        monthly_transaction_name VARCHAR(32) NOT NULL,
        monthly_transaction_amount BIGINT NOT NULL,
        monthly_transaction_date INT NOT NULL,
        category_id BIGINT NOT NULL,
        sub_category_id BIGINT NOT NULL,
        include_flg BOOLEAN NOT NULL,
        payment_id BIGINT,
        CONSTRAINT fk_user_no_monthly_transaction FOREIGN KEY (user_no) REFERENCES users (user_no),
        CONSTRAINT fk_category_id_monthly_transaction FOREIGN KEY (category_id) REFERENCES category (category_id),
        CONSTRAINT fk_sub_category_id_monthly_transaction FOREIGN KEY (sub_category_id) REFERENCES sub_category (sub_category_id),
        CONSTRAINT fk_payment_id_monthly_transaction FOREIGN KEY (payment_id) REFERENCES payment_resource (payment_id)
    );

/* ============================================================
INSERT DATA
============================================================ */
-- ユーザー
INSERT INTO
    users (user_id, user_no, email, PASSWORD)
VALUES
    (
        '4f4da417-7693-4fa1-b153-a3511ed1a57a',
        1,
        'MasterUser',
        '5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8'
    );

-- カテゴリ
INSERT INTO
    category (category_id, category_name, order_num)
VALUES
    (1, '食費', 1),
    (2, '外食', 2),
    (3, 'コンビニ', 3),
    (4, '日用品', 4),
    (5, 'ショッピング', 5),
    (6, 'ファッション', 6),
    (7, 'WEBサービス', 7),
    (8, 'エンタメ', 8),
    (9, '趣味', 9),
    (10, '旅行・レジャー', 10),
    (11, '交際費', 11),
    (12, 'ギフト', 12),
    (13, '交通費', 13),
    (14, '美容・コスメ', 14),
    (15, '医療・健康', 15),
    (16, '車', 16),
    (17, '教育', 17),
    (18, '子供', 18),
    (19, '手数料', 19),
    (20, '水道光熱費', 20),
    (21, '通信費', 21),
    (22, '住宅', 22),
    (23, '税金', 23),
    (24, '保険', 24),
    (25, '返済', 26),
    (26, 'ビジネス', 27),
    (27, '給与', 28),
    (28, 'その他収入', 30),
    (29, '投資', 25);

-- サブカテゴリ
INSERT INTO
    sub_category (user_no, category_id, sub_category_name)
VALUES
    (1, 1, 'なし'),
    (1, 2, 'なし'),
    (1, 2, 'カフェ'),
    (1, 2, 'お昼'),
    (1, 2, 'レストラン'),
    (1, 3, 'なし'),
    (1, 4, 'なし'),
    (1, 5, 'なし'),
    (1, 5, 'スポーツ用品'),
    (1, 5, 'ペット'),
    (1, 5, '電化製品'),
    (1, 6, 'なし'),
    (1, 7, 'なし'),
    (1, 8, 'なし'),
    (1, 8, '本・雑誌'),
    (1, 8, 'ゲーム'),
    (1, 8, '音楽'),
    (1, 8, '映画'),
    (1, 8, 'テレビ'),
    (1, 9, 'なし'),
    (1, 10, 'なし'),
    (1, 10, '博物館・劇場'),
    (1, 10, '遊園地'),
    (1, 10, 'ホテル'),
    (1, 10, '温泉'),
    (1, 11, 'なし'),
    (1, 11, 'ゲームセンター'),
    (1, 11, 'カラオケ'),
    (1, 11, '居酒屋・バー'),
    (1, 12, 'なし'),
    (1, 13, 'なし'),
    (1, 14, 'なし'),
    (1, 14, '美容'),
    (1, 14, 'コスメ'),
    (1, 15, 'なし'),
    (1, 15, '眼科'),
    (1, 15, '薬'),
    (1, 15, '病院'),
    (1, 15, 'ジム・フィットネス'),
    (1, 16, 'なし'),
    (1, 16, 'ガソリン'),
    (1, 16, '駐車場'),
    (1, 16, '自動車保険'),
    (1, 16, '有料道路'),
    (1, 16, '自動車税'),
    (1, 16, '維持費'),
    (1, 17, 'なし'),
    (1, 17, '学費'),
    (1, 17, '学生ローン返済'),
    (1, 17, '教科書等'),
    (1, 18, 'なし'),
    (1, 18, '小児科'),
    (1, 18, '保育園'),
    (1, 19, 'なし'),
    (1, 19, '利息'),
    (1, 19, '振込手数料'),
    (1, 19, '銀行手数料'),
    (1, 19, 'ATM手数料'),
    (1, 20, 'なし'),
    (1, 20, '電気'),
    (1, 20, '水道'),
    (1, 20, 'ガス'),
    (1, 21, 'なし'),
    (1, 21, '携帯電話'),
    (1, 21, '固定電話'),
    (1, 21, 'インターネット'),
    (1, 22, 'なし'),
    (1, 22, '家賃'),
    (1, 22, '住宅設備'),
    (1, 22, '保険'),
    (1, 23, 'なし'),
    (1, 23, '住民税'),
    (1, 24, 'なし'),
    (1, 24, 'その他の保険'),
    (1, 24, '生命保険'),
    (1, 25, 'なし'),
    (1, 25, 'カード引落し'),
    (1, 26, 'なし'),
    (1, 26, 'ビジネス'),
    (1, 26, 'オフィス用品'),
    (1, 26, '法務会計'),
    (1, 26, '送料'),
    (1, 26, 'オフィス設備'),
    (1, 26, '印刷'),
    (1, 27, 'なし'),
    (1, 27, 'ボーナス'),
    (1, 28, 'なし'),
    (1, 28, '家賃所得'),
    (1, 28, '利子所得'),
    (1, 28, 'お小遣い'),
    (1, 29, 'なし');

-- 支払いの種類
INSERT INTO
    payment_type (
        payment_type_id,
        payment_type_name,
        is_payment_due_later,
        order_num
    )
VALUES
    (1, '現金', FALSE, 1),
    (2, 'カード', TRUE, 2),
    (3, 'QRペイ', FALSE, 3);