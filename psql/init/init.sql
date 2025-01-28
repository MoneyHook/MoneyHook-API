DROP TABLE IF EXISTS users,
user_token,
category,
sub_category,
hidden_sub_category,
transaction,
monthly_transaction,
payment_type,
payment_resource CASCADE;

CREATE TABLE users (
    user_id VARCHAR(64) NOT NULL,
    user_no BIGSERIAL PRIMARY KEY,
    email VARCHAR(128) UNIQUE,
    PASSWORD TEXT
);

CREATE TABLE category (
    category_id BIGSERIAL PRIMARY KEY,
    category_name VARCHAR(16) NOT NULL,
    order_num INT NOT NULL
);

CREATE TABLE user_token (
    user_no BIGINT NOT NULL,
    token VARCHAR(64) NOT NULL,
    CONSTRAINT fk_user_no FOREIGN KEY (user_no) REFERENCES users (user_no)
);

CREATE TABLE sub_category (
    sub_category_id BIGSERIAL PRIMARY KEY,
    user_no BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    sub_category_name VARCHAR(16) NOT NULL,
    CONSTRAINT fk_user_no_sub FOREIGN KEY (user_no) REFERENCES users (user_no),
    CONSTRAINT fk_category_id FOREIGN KEY (category_id) REFERENCES category (category_id),
    UNIQUE (user_no, category_id, sub_category_name)
);

CREATE TABLE hidden_sub_category (
    user_no BIGINT NOT NULL,
    sub_category_id BIGINT NOT NULL,
    CONSTRAINT fk_user_no_hidden FOREIGN KEY (user_no) REFERENCES users (user_no),
    CONSTRAINT fk_sub_category_id FOREIGN KEY (sub_category_id) REFERENCES sub_category (sub_category_id),
    UNIQUE (user_no, sub_category_id)
);

CREATE TABLE payment_type (
    payment_type_id BIGSERIAL PRIMARY KEY,
    payment_type_name VARCHAR(32) NOT NULL,
    is_payment_due_later BOOLEAN NOT NULL,
    order_num INT NOT NULL,
    UNIQUE (payment_type_name)
);

CREATE TABLE payment_resource (
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

CREATE TABLE transaction (
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

CREATE TABLE monthly_transaction (
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