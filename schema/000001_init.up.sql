CREATE TABLE Wallet
(
    wallet_id bigint not null unique,
    USDT float default 0,
    RUB float default 0,
    EUR float default 0
);

CREATE TABLE Transaction
(
    id serial not null unique,
    wallet_id bigint not null,
    currency char(100) not null,
    sum float,
    status char(100) default 'created'
);