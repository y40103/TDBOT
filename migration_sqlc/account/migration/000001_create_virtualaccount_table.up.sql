CREATE TABLE "accountinfo" (
                               "account_id" varchar(10) PRIMARY KEY NOT NULL,
                               "balance" decimal NOT NULL,
                               "created_at" timestamptz DEFAULT (now()),
                               "version" int NOT NULL DEFAULT 0
);

CREATE TABLE "hold_symbol" (
                               "account_id" varchar(10) NOT NULL,
                               "symbol" varchar(10) PRIMARY KEY,
                               "quantity" int NOT NULL,
                               "created_at" timestamptz DEFAULT (now())
);

CREATE TABLE "processing_task" (
                                   "account_id" varchar(10) NOT NULL,
                                   "task_id" varchar(10) PRIMARY KEY,
                                   "symbol" varchar(10) NOT NULL,
                                   "created_at" timestamptz DEFAULT (now())
);

CREATE TABLE "his_buy" (
                           "task_id" varchar(10) PRIMARY KEY,
                           "symbol" varchar(10) NOT NULL,
                           "account_id" varchar(10) NOT NULL,
                           "buy_price" decimal NOT NULL,
                           "quantity" int NOT NULL,
                           "trigger" varchar(30) NOT NULL,
                           "created_at" timestamptz DEFAULT (now())
);

CREATE TABLE "his_sell" (
                            "task_id" varchar(10) NOT NULL,
                            "symbol" varchar(10) NOT NULL,
                            "account_id" varchar(10) NOT NULL,
                            "sell_price" decimal NOT NULL,
                            "quantity" int NOT NULL,
                            "income" decimal NOT NULL,
                            "trigger" varchar(30) NOT NULL,
                            "created_at" timestamptz DEFAULT (now())
);

CREATE INDEX ON "accountinfo" ("account_id");

CREATE INDEX ON "hold_symbol" ("created_at");

CREATE INDEX ON "processing_task" ("account_id");

CREATE INDEX ON "processing_task" ("symbol");

CREATE INDEX ON "processing_task" ("created_at");

CREATE INDEX ON "his_buy" ("created_at");

CREATE INDEX ON "his_buy" ("trigger");

CREATE INDEX ON "his_sell" ("task_id");

CREATE INDEX ON "his_sell" ("created_at");

CREATE INDEX ON "his_sell" ("trigger");

COMMENT ON COLUMN "accountinfo"."balance" IS 'not can be negtive';