-- ユーザー
INSERT INTO
    users (user_id, email, PASSWORD)
VALUES
    (
        '4f4da417-7693-4fa1-b153-a3511ed1a57a',
        'MasterUser',
        '5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8'
    ),
    (
        'a77a6e94-6aa2-47ea-87dd-129f580fb669',
        'sample@sample.com',
        '5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8'
    );

-- トークン
INSERT INTO
    user_token (user_no, token)
VALUES
    (2, 'sample_token');

-- カテゴリ
INSERT INTO
    category (category_name, order_num)
VALUES
    ('食費', 1),
    ('外食', 2),
    ('コンビニ', 3),
    ('日用品', 4),
    ('ショッピング', 5),
    ('ファッション', 6),
    ('WEBサービス', 7),
    ('エンタメ', 8),
    ('趣味', 9),
    ('旅行・レジャー', 10),
    ('交際費', 11),
    ('ギフト', 12),
    ('交通費', 13),
    ('美容・コスメ', 14),
    ('医療・健康', 15),
    ('車', 16),
    ('教育', 17),
    ('子供', 18),
    ('手数料', 19),
    ('水道光熱費', 20),
    ('通信費', 21),
    ('住宅', 22),
    ('税金', 23),
    ('保険', 24),
    ('返済', 26),
    ('ビジネス', 27),
    ('給与', 28),
    ('その他収入', 30),
    ('投資', 25);

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

INSERT INTO
    sub_category (user_no, category_id, sub_category_name)
VALUES
    (2, 2, 'ラーメン巡り'),
    (2, 2, '寿司巡り'),
    (2, 2, 'カフェ巡り'),
    (2, 7, 'Amazon Prime'),
    (2, 7, 'Youtube Premium'),
    (2, 9, '自転車用品'),
    (2, 22, 'DIY');

-- 非表示サブカテゴリ
INSERT INTO
    hidden_sub_category (user_no, sub_category_id)
VALUES
    (2, 3),
    (2, 10),
    (2, 19);

-- 取引
INSERT INTO
    transaction (
        user_no,
        transaction_name,
        transaction_amount,
        transaction_date,
        category_id,
        sub_category_id,
        fixed_flg
    )
VALUES
    (2, 'サーキュレーター', -646, '2024-06-30', 5, 8, FALSE),
    (2, 'AppleOne', -130, '2024-06-30', 7, 13, TRUE),
    (2, '水道代', -1927, '2024-06-28', 20, 59, TRUE),
    (2, 'カーシェア', -9900, '2024-06-27', 10, 21, FALSE),
    (2, 'NIKEウェア', -1570, '2024-06-27', 5, 8, FALSE),
    (2, '家賃', -78000, '2024-06-27', 22, 68, TRUE),
    (2, '奨学金', -14222, '2024-06-27', 25, 76, TRUE),
    (2, 'スーパーアルプス', -1999, '2024-06-17', 1, 1, TRUE),
    (2, 'サプリメント', -544, '2024-06-24', 5, 8, FALSE),
    (2, 'Youtube', -1180, '2024-06-24', 7, 13, TRUE),
    (2, '給与', 185305, '2024-06-25', 27, 85, TRUE),
    (2, '配当', 305, '2024-06-25', 28, 87, FALSE),
    (2, 'タバコ', -550, '2024-06-23', 3, 6, FALSE),
    (2, 'スーパーアルプス', -431, '2024-06-23', 1, 1, TRUE),
    (2, 'ヘアカット', -3300, '2024-06-22', 14, 32, FALSE),
    (2, 'Webを支える技術', -2827, '2024-06-21', 5, 8, FALSE),
    (2, 'スーパーアルプス', -3182, '2024-06-21', 1, 1, TRUE),
    (2, 'アイスノン', -625, '2024-06-20', 5, 8, FALSE),
    (2, 'ヨーグリーナ', -117, '2024-06-19', 3, 6, FALSE),
    (2, '死刑にいたる病', -1000, '2024-06-19', 8, 18, FALSE),
    (2, 'スーパーアルプス', -1561, '2024-06-18', 1, 1, TRUE),
    (2, '100均', -330, '2024-06-18', 5, 8, FALSE),
    (2, 'スーパーアルプス', -1705, '2024-06-17', 1, 1, TRUE),
    (2, 'ガス代', -4376, '2024-06-15', 20, 59, TRUE),
    (2, 'ウェルシア', -600, '2024-06-14', 5, 8, FALSE),
    (2, 'タバコ', -600, '2024-06-14', 3, 6, FALSE),
    (2, '電気代', -3225, '2024-06-13', 20, 60, TRUE),
    (2, '電話料金', -1781, '2024-06-13', 21, 64, TRUE),
    (2, 'スーパーアルプス', -2757, '2024-06-12', 1, 1, TRUE),
    (2, 'はなまるうどん', -610, '2024-06-11', 2, 4, FALSE),
    (2, '台湾まぜそば', -850, '2024-06-11', 2, 2, FALSE),
    (2, 'DisneyPlys', -990, '2024-06-08', 7, 13, TRUE),
    (2, 'ふるさと納税', -10000, '2024-06-05', 5, 8, FALSE),
    (2, 'タバコ', -540, '2024-06-04', 3, 6, FALSE),
    (2, 'スーパーアルプス', -4153, '2024-06-04', 1, 1, TRUE),
    (2, '洗車グッズ', -1900, '2024-06-03', 5, 8, FALSE),
    (2, 'ほうじ茶ラテ', -160, '2024-06-03', 3, 6, FALSE),
    (2, 'ふるさと納税', -5000, '2024-06-02', 5, 8, FALSE),
    (2, '洗車スポンジ', -190, '2024-06-02', 5, 8, FALSE),
    (2, '牛乳', -192, '2024-06-02', 1, 1, FALSE),
    (2, 'きさらぎ駅', -1000, '2024-06-02', 8, 18, FALSE),
    (2, 'ケーブルトレー', -1580, '2024-06-01', 5, 8, FALSE),
    (2, 'ワイプオール', -1330, '2024-06-01', 5, 8, FALSE),
    (2, 'サーキュレーター', -646, '2024-07-30', 5, 8, FALSE),
    (2, 'AppleOne', -130, '2024-07-30', 7, 13, TRUE),
    (2, '水道代', -1927, '2024-07-28', 20, 59, TRUE),
    (2, 'カーシェア', -9900, '2024-07-27', 10, 21, FALSE),
    (2, 'NIKEウェア', -1570, '2024-07-27', 5, 8, FALSE),
    (2, '家賃', -78000, '2024-07-27', 22, 68, TRUE),
    (2, '奨学金', -14222, '2024-07-27', 25, 76, TRUE),
    (2, 'スーパーアルプス', -1999, '2024-07-17', 1, 1, TRUE),
    (2, 'サプリメント', -544, '2024-07-24', 5, 8, FALSE),
    (2, 'Youtube', -1180, '2024-07-24', 7, 13, TRUE),
    (2, '給与', 185305, '2024-07-25', 27, 85, TRUE),
    (2, '配当', 305, '2024-07-25', 28, 87, FALSE),
    (2, 'タバコ', -550, '2024-07-23', 3, 6, FALSE),
    (2, 'スーパーアルプス', -431, '2024-07-23', 1, 1, TRUE),
    (2, 'ヘアカット', -3300, '2024-07-22', 14, 32, FALSE),
    (2, 'Webを支える技術', -2827, '2024-07-21', 5, 8, FALSE),
    (2, 'スーパーアルプス', -3182, '2024-07-21', 1, 1, TRUE),
    (2, 'アイスノン', -625, '2024-07-20', 5, 8, FALSE),
    (2, 'ヨーグリーナ', -117, '2024-07-19', 3, 6, FALSE),
    (2, '死刑にいたる病', -1000, '2024-07-19', 8, 18, FALSE),
    (2, 'スーパーアルプス', -1561, '2024-07-18', 1, 1, TRUE),
    (2, '100均', -330, '2024-07-18', 5, 8, FALSE),
    (2, 'スーパーアルプス', -1705, '2024-07-17', 1, 1, TRUE),
    (2, 'ガス代', -4376, '2024-07-15', 20, 59, TRUE),
    (2, 'ウェルシア', -600, '2024-07-14', 5, 8, FALSE),
    (2, 'タバコ', -600, '2024-07-14', 3, 6, FALSE),
    (2, '電気代', -3225, '2024-07-13', 20, 60, TRUE),
    (2, '電話料金', -1781, '2024-07-13', 21, 64, TRUE),
    (2, 'スーパーアルプス', -2757, '2024-07-12', 1, 1, TRUE),
    (2, 'はなまるうどん', -610, '2024-07-11', 2, 4, FALSE),
    (2, '台湾まぜそば', -850, '2024-07-11', 2, 2, FALSE),
    (2, 'DisneyPlys', -990, '2024-07-08', 7, 13, TRUE),
    (2, 'ふるさと納税', -10000, '2024-07-05', 5, 8, FALSE),
    (2, 'タバコ', -540, '2024-07-04', 3, 6, FALSE),
    (2, 'スーパーアルプス', -4153, '2024-07-04', 1, 1, TRUE),
    (2, '洗車グッズ', -1900, '2024-07-03', 5, 8, FALSE),
    (2, 'ほうじ茶ラテ', -160, '2024-07-03', 3, 6, FALSE),
    (2, 'ふるさと納税', -5000, '2024-07-02', 5, 8, FALSE),
    (2, '洗車スポンジ', -190, '2024-07-02', 5, 8, FALSE),
    (2, '牛乳', -192, '2024-07-02', 1, 1, FALSE),
    (2, 'きさらぎ駅', -1000, '2024-07-02', 8, 18, FALSE),
    (2, 'ケーブルトレー', -1580, '2024-07-01', 5, 8, FALSE),
    (2, 'ワイプオール', -1330, '2024-07-01', 5, 8, FALSE);

-- 月次取引
INSERT INTO
    monthly_transaction (
        user_no,
        monthly_transaction_name,
        monthly_transaction_amount,
        monthly_transaction_date,
        category_id,
        sub_category_id,
        include_flg
    )
VALUES
    (2, '家賃', -78550, '31', 22, 67, TRUE),
    (2, '見放題chライト', -550, '27', 7, 13, TRUE),
    (2, 'Youtube Premium', -1150, '25', 7, 13, TRUE),
    (2, 'DisneyPlus', -980, '25', 7, 13, TRUE),
    (2, '給与', 200000, '25', 27, 85, TRUE),
    (2, 'Amazon', -550, '25', 7, 13, FALSE);

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

-- 支払い方法
INSERT INTO
    payment_resource (
        payment_id,
        payment_type_id,
        user_no,
        payment_name,
        payment_date,
        closing_date
    )
VALUES
    (1, 2, 2, '楽天カード', 27, 31);

INSERT INTO
    payment_resource (
        payment_id,
        payment_type_id,
        user_no,
        payment_name
    )
VALUES
    (2, DEFAULT, 2, '現金'),
    (3, 3, 2, 'PayPay');