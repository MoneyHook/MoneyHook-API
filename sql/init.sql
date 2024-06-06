DROP TABLE IF EXISTS users,
user_token,
category,
sub_category,
hidden_sub_category,
transaction,
monthly_transaction,
payment_resource;

CREATE TABLE users (
        user_id VARCHAR(64) NOT NULL,
        user_no BIGINT UNSIGNED AUTO_INCREMENT,
        email VARCHAR(128) UNIQUE,
        PASSWORD text,
        PRIMARY KEY (user_no)
    );

CREATE TABLE category (
        category_id BIGINT UNSIGNED AUTO_INCREMENT,
        category_name VARCHAR(16) NOT NULL,
        PRIMARY KEY (category_id)
    );

CREATE TABLE user_token (
        user_no BIGINT UNSIGNED NOT NULL,
        token VARCHAR(64) NOT NULL,
        FOREIGN KEY user_no (user_no) REFERENCES users (user_no)
    );

CREATE TABLE sub_category (
        sub_category_id BIGINT UNSIGNED AUTO_INCREMENT,
        user_no BIGINT UNSIGNED NOT NULL,
        category_id BIGINT UNSIGNED NOT NULL,
        sub_category_name VARCHAR(16) NOT NULL,
        PRIMARY KEY (sub_category_id),
        FOREIGN KEY user_no (user_no) REFERENCES users (user_no),
        FOREIGN KEY category_id (category_id) REFERENCES category (category_id),
        UNIQUE (user_no, category_id, sub_category_name)
    );

CREATE TABLE hidden_sub_category (
        user_no BIGINT UNSIGNED NOT NULL,
        sub_category_id BIGINT UNSIGNED NOT NULL,
        FOREIGN KEY user_no (user_no) REFERENCES users (user_no),
        FOREIGN KEY sub_category_id (sub_category_id) REFERENCES sub_category (sub_category_id),
        UNIQUE (user_no, sub_category_id)
    );

CREATE TABLE payment_resource (
        payment_id BIGINT UNSIGNED AUTO_INCREMENT,
        user_no BIGINT UNSIGNED NOT NULL,
        payment_name VARCHAR(32) NOT NULL,
        PRIMARY KEY (payment_id),
        FOREIGN KEY user_no (user_no) REFERENCES users (user_no),
        UNIQUE (user_no, payment_name)
    );

CREATE TABLE transaction (
        transaction_id BIGINT UNSIGNED AUTO_INCREMENT,
        user_no BIGINT UNSIGNED NOT NULL,
        transaction_name VARCHAR(32) NOT NULL,
        transaction_amount BIGINT NOT NULL,
        transaction_date DATE NOT NULL,
        category_id BIGINT UNSIGNED NOT NULL,
        sub_category_id BIGINT UNSIGNED NOT NULL,
        fixed_flg BOOLEAN NOT NULL,
        payment_id BIGINT UNSIGNED,
        PRIMARY KEY (transaction_id),
        FOREIGN KEY user_no (user_no) REFERENCES users (user_no),
        FOREIGN KEY category_id (category_id) REFERENCES category (category_id),
        FOREIGN KEY sub_category_id (sub_category_id) REFERENCES sub_category (sub_category_id),
        FOREIGN KEY payment_id (payment_id) REFERENCES payment_resource (payment_id)
    );

CREATE TABLE monthly_transaction (
        monthly_transaction_id BIGINT UNSIGNED AUTO_INCREMENT,
        user_no BIGINT UNSIGNED NOT NULL,
        monthly_transaction_name VARCHAR(32) NOT NULL,
        monthly_transaction_amount BIGINT NOT NULL,
        monthly_transaction_date INT NOT NULL,
        category_id BIGINT UNSIGNED NOT NULL,
        sub_category_id BIGINT UNSIGNED NOT NULL,
        include_flg BOOLEAN NOT NULL,
        payment_id BIGINT UNSIGNED,
        PRIMARY KEY (monthly_transaction_id),
        FOREIGN KEY user_no (user_no) REFERENCES users (user_no),
        FOREIGN KEY category_id (category_id) REFERENCES category (category_id),
        FOREIGN KEY sub_category_id (sub_category_id) REFERENCES sub_category (sub_category_id),
        FOREIGN KEY payment_id (payment_id) REFERENCES payment_resource (payment_id)
    );