ALTER TABLE subscriptions ADD COLUMN from_city TEXT NOT NULL DEFAULT '';
ALTER TABLE subscriptions ADD COLUMN to_city TEXT NOT NULL DEFAULT '';

UPDATE subscriptions s
SET from_city = a.city
FROM airports a
WHERE a.iata_code = s.from_iata
  AND a.language = 'ru'
  AND a.city IS NOT NULL;

UPDATE subscriptions s
SET to_city = a.city
FROM airports a
WHERE a.iata_code = s.to_iata
  AND a.language = 'ru'
  AND a.city IS NOT NULL;
