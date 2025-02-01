INSERT INTO apps (id, name, secret)
VALUES (1551816918, 'test', 'test-secret')
    ON CONFLICT DO NOTHING;