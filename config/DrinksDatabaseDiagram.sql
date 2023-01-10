
CREATE TABLE IF NOT EXISTS bars (
  id bigserial PRIMARY KEY,
  title varchar NOT NULL,
  address varchar NOT NULL,
  description varchar NOT NULL,
  drinks_id bigint[],
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now(),
  UNIQUE (title, address) 
);

CREATE TABLE IF NOT EXISTS drinks (
  id bigserial PRIMARY KEY,
  title varchar NOT NULL,
  price int NOT NULL,
  type int NOT NULL,
  description varchar NOT NULL,
  bar_id bigserial,
  ingredients_id bigint[],
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now(),
  UNIQUE (bar_id, title) 
);

CREATE TABLE IF NOT EXISTS ingredients (
  id bigserial PRIMARY KEY,
  title varchar NOT NULL,
  amount int NOT NULL,
  drink_id bigint,
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now(),
  UNIQUE (drink_id, title) 
);

CREATE INDEX ON bars (title);

CREATE INDEX ON bars (address);

CREATE INDEX ON drinks (bar_id);

CREATE INDEX ON drinks (title);

CREATE INDEX ON drinks (bar_id, title);

CREATE INDEX ON ingredients (drink_id);

CREATE INDEX ON ingredients (drink_id, title);

ALTER TABLE drinks ADD FOREIGN KEY (bar_id) REFERENCES bars (id) ON DELETE CASCADE;

ALTER TABLE ingredients ADD FOREIGN KEY (drink_id) REFERENCES drinks (id) ON DELETE CASCADE;