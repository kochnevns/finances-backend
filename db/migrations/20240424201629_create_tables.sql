-- migrate:up

CREATE TABLE Categories (
	id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	icon TEXT
);

CREATE TABLE "Expenses"
(
    id          INTEGER not null
        primary key autoincrement,
    date        TEXT,
    description TEXT,
    amount      INTEGER,
    category_id INTEGER
);

-- migrate:down

