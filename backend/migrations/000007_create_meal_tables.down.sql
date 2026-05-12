-- Drop meal tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS meal_payments;
DROP TABLE IF EXISTS meal_participations;
DROP TABLE IF EXISTS meal_members;
