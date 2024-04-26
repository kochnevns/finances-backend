-- migrate:up
UPDATE Categories SET name = 'Здоровье' WHERE name = 'Здооровье';
INSERT INTO Categories (name) VALUES ('Одежда');

-- migrate:down

