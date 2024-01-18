CREATE TABLE IF NOT EXISTS Wallets
(
	wallet_id bigint not null,
	currency char(20) not null,
	value float default 0 not null
);
		
CREATE INDEX ix_wallets_person_id ON Wallets (wallet_id);

CREATE TABLE IF NOT EXISTS Transactions
(
	id serial not null unique,
	wallet_id bigint not null,
	currency char(100) not null,
	sum float,
	status char(100) default 'Created'
);