CREATE TYPE "DrinkType" AS ENUM (
  'LONG_DRINK'
);

CREATE TABLE IF NOT EXISTS "bars" (
  "id" bigserial PRIMARY KEY,
  "title" varchar NOT NULL,
  "address" varchar NOT NULL,
  "description" varchar NOT NULL,
  "drinks_id" bigserial[],
  "created at" timestamptz DEFAULT 'now()',
  "updated at" timestamptz DEFAULT 'now()',
  ("title", "address") UNIQUE
);

CREATE TABLE IF NOT EXISTS "drinks" (
  "id" bigserial PRIMARY KEY,
  "title" varchar NOT NULL,
  "price" int NOT NULL,
  "type" DrinkType NOT NULL,
  "description" verchar NOT NULL,
  "bar_id" bigserial,
  "ingredients_id" bigserial[],
  "created at" timestamptz DEFAULT 'now()',
  "updated at" timestamptz DEFAULT 'now()',
  ("bar_id", "title") UNIQUE
);

CREATE TABLE IF NOT EXISTS "ingredients" (
  "id" bigserial PRIMARY KEY,
  "title" varchar NOT NULL,
  "amount" int NOT NULL,
  "drink_id" bigserial,
  "created at" timestamptz DEFAULT 'now()',
  "updated at" timestamptz DEFAULT 'now()',
  ("drink_id", "title") UNIQUE
);

CREATE INDEX ON "bars" ("title");

CREATE INDEX ON "bars" ("address");

CREATE INDEX ON "drinks" ("bar_id");

CREATE INDEX ON "drinks" ("title");

CREATE INDEX ON "drinks" ("bar_id", "title");

CREATE INDEX ON "ingredients" ("drink_id");

CREATE INDEX ON "ingredients" ("drink_id", "title");

ALTER TABLE "drinks" ADD FOREIGN KEY ("bar_id") REFERENCES "bars" ("id");

ALTER TABLE "ingredients" ADD FOREIGN KEY ("drink_id") REFERENCES "drinks" ("id");
