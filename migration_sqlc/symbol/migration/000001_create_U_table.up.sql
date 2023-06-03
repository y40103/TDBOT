CREATE TABLE "2023" (
                        "tradingtime" timestamp NOT NULL,
                        "price" numeric NOT NULL,
                        "volume" bigint NOT NULL,
                        "eventnum" integer NOT NULL,
                        "insrate" integer NOT NULL,
                        "nextrate" integer NOT NULL,
                        "formt" boolean NOT NULL,
                        "created_at" timestamp DEFAULT (now())
);

CREATE INDEX ON "2023" ("tradingtime");