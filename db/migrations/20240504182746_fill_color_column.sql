-- migrate:up
UPDATE Categories
SET name='Доставки', icon=NULL, color='#ff4747'
WHERE id=1;
UPDATE Categories
SET name='Гаджеты', icon=NULL, color='#00857a'
WHERE id=2;
UPDATE Categories
SET name='Готовая еда', icon=NULL, color='#fa6000'
WHERE id=3;
UPDATE Categories
SET name='Дуделки', icon=NULL, color='#ff1fad'
WHERE id=4;
UPDATE Categories
SET name='Моти', icon=NULL, color='#7857ff'
WHERE id=5;
UPDATE Categories
SET name='Путешествия', icon=NULL, color='#268500'
WHERE id=6;
UPDATE Categories
SET name='Рестораны и кафе', icon=NULL, color='#FFC800'
WHERE id=7;
UPDATE Categories
SET name='Буханка', icon=NULL, color='#666'
WHERE id=8;
UPDATE Categories
SET name='Здоровье', icon=NULL, color='#4CCD99'
WHERE id=9;
UPDATE Categories
SET name='Ипотека', icon=NULL, color='#8B322C'
WHERE id=10;
UPDATE Categories
SET name='Коммунальные платежи', icon=NULL, color='#8B322C'
WHERE id=11;
UPDATE Categories
SET name='Сигареты', icon=NULL, color='#FFC55A'
WHERE id=12;
UPDATE Categories
SET name='Онлайн подписки и лицензии', icon=NULL, color='#007F73'
WHERE id=13;
UPDATE Categories
SET name='Продукты', icon=NULL, color='#93FCF8'
WHERE id=14;
UPDATE Categories
SET name='Развлечения', icon=NULL, color='#8F80F3'
WHERE id=15;
UPDATE Categories
SET name='Такси', icon=NULL, color='#777'
WHERE id=16;
UPDATE Categories
SET name='Товары для дома', icon=NULL, color='#FF8A00'
WHERE id=17;
UPDATE Categories
SET name='Общественный транспорт', icon=NULL, color='#555'
WHERE id=18;
UPDATE Categories
SET name='Хуйня всякая', icon=NULL, color='#F9F9E0'
WHERE id=19;
UPDATE Categories
SET name='Одежда', icon=NULL, color='#3BE9DE'
WHERE id=20;

-- migrate:down

