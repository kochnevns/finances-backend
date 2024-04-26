-- migrate:up
UPDATE Expenses SET category_id = 9 WHERE category_id is NULL;
UPDATE Expenses SET category_id = 20 WHERE description = 'Бифри шмотки';
-- migrate:down

