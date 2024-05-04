CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE Categories (
	id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	icon TEXT
, color TEXT);
CREATE TABLE IF NOT EXISTS "Expenses"
(
    id          INTEGER not null
        primary key autoincrement,
    date        TEXT,
    description TEXT,
    amount      INTEGER,
    category_id INTEGER
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20240424201629'),
  ('20240424201734'),
  ('20240426164111'),
  ('20240426164205'),
  ('20240426165157'),
  ('20240426165802'),
  ('20240426171925'),
  ('20240426172847'),
  ('20240430184719'),
  ('20240504182624'),
  ('20240504182746');
