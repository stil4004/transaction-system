CREATE TABLE IF NOT EXISTS Wallets
(
	wallet_id bigint not null,
	currency char(10) not null,
	value float default 0 not null
);
		
CREATE INDEX ix_wallets_person_id ON Wallets (wallet_id);

CREATE TABLE IF NOT EXISTS Transactions
(
	id serial not null unique,
	wallet_id bigint not null,
	currency char(10) not null,
	typeOF char(20) not null,
	sum float,
	status char(20) default 'Created'
);

CREATE TABLE IF NOT EXISTS Transfers
(
		id serial not null unique,
		wallet_id_from bigint not null,
		wallet_id_to bigint not null,
		currency char(10) not null,
		sum float
);