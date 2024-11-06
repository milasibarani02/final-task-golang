CREATE TABLE public.auths (
	auth_id int8 GENERATED ALWAYS AS IDENTITY NOT NULL,
	username varchar NOT NULL,
	"password" varchar NOT NULL,
	CONSTRAINT auths_pk null,
	CONSTRAINT auths_unique UNIQUE (username)
);

CREATE TABLE public.accounts (
	account_id int8 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 9223372036854775807 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"name" varchar NOT NULL,
	balance int8 DEFAULT 0 NOT NULL,
	referral_account_id int8 NULL,
	CONSTRAINT account_pk PRIMARY KEY (account_id),
	CONSTRAINT fk_account FOREIGN KEY (referral_account_id) REFERENCES public.accounts(account_id)
);

CREATE TABLE public.transaction_categories (
	transaction_category_id int4 GENERATED ALWAYS AS IDENTITY NOT NULL,
	"name" varchar NULL,
	CONSTRAINT transaction_categories_pk PRIMARY KEY (transaction_category_id)
);


CREATE TABLE public."transaction" (
	transaction_id int8 GENERATED ALWAYS AS IDENTITY NOT NULL,
	transaction_category_id int8 NULL,
	account_id int8 NOT NULL,
	from_account_id int8 NULL,
	to_account_id int8 NULL,
	amount int8 NULL,
	transaction_date timestamp NULL,
	CONSTRAINT transaction_pk PRIMARY KEY (transaction_id)
);

