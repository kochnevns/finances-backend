-- migrate:up

DELETE FROM Expenses WHERE description like '%test%';
-- migrate:down

