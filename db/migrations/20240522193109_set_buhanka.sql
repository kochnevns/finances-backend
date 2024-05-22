-- migrate:up

update Expenses set category_id = (select id from Categories where name = 'Буханка') where description like '%Буханка%';

-- migrate:down

