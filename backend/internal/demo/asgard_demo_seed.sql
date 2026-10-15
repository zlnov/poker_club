-- ============================================================================
-- DEMO TEMPLATE SEED for poker_club_demo
-- Creates a single Asgard (Demo) club with is_template = TRUE.
-- No suffixes: this script is loaded once; sessions are cloned in Go.
-- Preserves email and tg_user_name. Generated from asgard_demo_seed_auto_fill-5.sql
-- ============================================================================

BEGIN;

-- Idempotent: skip if a template club already exists.
DO $guard$
BEGIN
    IF EXISTS (SELECT 1 FROM clubs WHERE is_template = TRUE) THEN
        RAISE EXCEPTION 'DEMO_TEMPLATE_EXISTS: Asgard demo template already loaded';
    END IF;
END
$guard$;

-- ============================================================================
-- 1. TEMP MAPPING TABLES
-- ============================================================================
CREATE TEMP TABLE club_map   (key TEXT PRIMARY KEY, id BIGINT) ON COMMIT DROP;
CREATE TEMP TABLE player_map (key TEXT PRIMARY KEY, id BIGINT) ON COMMIT DROP;
CREATE TEMP TABLE member_map (key TEXT PRIMARY KEY, id BIGINT) ON COMMIT DROP;
CREATE TEMP TABLE game_map   (key TEXT PRIMARY KEY, id BIGINT) ON COMMIT DROP;

-- ============================================================================
-- 2. CLUB (template)
-- ============================================================================
WITH inserted AS (
    INSERT INTO clubs (tg_chat_id, name, is_template)
    VALUES (NULL, 'Asgard (Demo)', TRUE)
    RETURNING id
)
INSERT INTO club_map (key, id)
SELECT 'asgard', id FROM inserted;

-- ============================================================================
-- 3. PLAYERS (email + tg_user_name filled; stable synthetic tg_user_id / phone)
-- ============================================================================
WITH inserted AS (
    INSERT INTO players (
        first_name, last_name, nickname, phone_number, email, password,
        tg_user_id, tg_user_name
    )
    SELECT
        p.first_name,
        p.last_name,
        p.nickname,
        p.phone,
        p.email,
        p.password,
        p.tg_user_id,
        p.tg_user_name
    FROM (VALUES
        ('Thor',      'Odinson',        'Thor',          '8889261230000', 'thor@hpclub.com',          '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999980::BIGINT, 'Thor=$'),
        ('Tony',      'Stark',          'Iron Man',      '8889261230001', 'iron_man@hpclub.com',      '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999981, 'IronMan=$'),
        ('Steve',     'Rogers',         'Captain',       '8889261230002', 'captain@hpclub.com',       '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999982, 'Captain=$'),
        ('Bruce',     'Banner',         'Hulk',          '8889261230003', 'hulk@hpclub.com',          '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999983, 'Hulk=$'),
        ('Hank',      'Pym',            'Ant-Man',       '8889261230004', 'ant-man@hpclub.com',       '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999984, 'Ant-Man=$'),
        ('Janet',     'Van Dyne',       'Wasp',          '8889261230005', 'wasp@hpclub.com',          '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999985, 'Wasp=$'),
        ('Clint',     'Barton',         'Hawkeye',       '8889261230006', 'hawkeye@hpclub.com',       '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999986, 'Hawkeye=$'),
        ('Peter',     'Parker',         'Spider-Man',    '8889261230007', 'spider-man@hpclub.com',    '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999987, 'Spider-Man=$'),
        ('Wanda',     'Maximoff',       'Scarlet Witch', '8889261230008', 'scarlet_witch@hpclub.com', '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999988, 'Scarlet Witch=$'),
        ('Natasha',   'Romanoff',       'Black Widow',   '8889261230009', 'black_widow@hpclub.com',   '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999989, 'Black Widow=$'),
        ('Vision',    '(no last name)', 'Vision',        '8889261230010', 'vision@hpclub.com',        '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999990, 'Vision=$'),
        ('T''Challa', '(no last name)', 'Black Panther', '8889261230011', 'black_panther@hpclub.com', '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999991, 'Black Panther=$'),
        ('Peter',     'Quill',          'Star-Lord',     '8889261230012', 'star-lord@hpclub.com',     '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999992, 'Star-Lord=$'),
        ('Gamora',    '(no last name)', 'Gamora',        '8889261230013', 'gamora@hpclub.com',        '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999993, 'Gamora=$'),
        ('Groot',     '(no last name)', 'Groot',         '8889261230014', 'groot@hpclub.com',         '$2a$12$3LO4iZahfCGdlEu0S5bcvObg4QItcW4dxcOvJKFCf6DVbovQCAZZe', 9999999999994, 'Groot=$')
    ) AS p(first_name, last_name, nickname, phone, email, password, tg_user_id, tg_user_name)
    RETURNING id, nickname
)
INSERT INTO player_map (key, id)
SELECT nickname, id FROM inserted;

-- ============================================================================
-- 4. CLUB MEMBERS
-- ============================================================================
WITH inserted AS (
    INSERT INTO club_members (club_id, player_id, role, status, accepted)
    SELECT
        (SELECT id FROM club_map WHERE key = 'asgard'),
        pm.id,
        CASE pm.key
            WHEN 'Thor'     THEN 'owner'
            WHEN 'Iron Man' THEN 'admin'
            ELSE                 'member'
        END,
        'active',
        TRUE
    FROM player_map pm
    RETURNING id, player_id
)
INSERT INTO member_map (key, id)
SELECT pm.key, ins.id
FROM inserted ins
JOIN player_map pm ON pm.id = ins.player_id;

-- ============================================================================
-- 5. GAMES — FINISHED (20 games)
-- ============================================================================
WITH inserted AS (
    INSERT INTO games (
        club_id, banker_id, game_type, currency, money_model, chip_value,
        buy_in_amount, rebuy_allowed, rebuy_price, max_rebuys, duration,
        start_time, end_time, status, min_players, max_players,
        ranking_primary, ranking_secondary,
        timer_paused_at, timer_paused_duration, timer_notified
    )
    SELECT
        (SELECT id FROM club_map WHERE key = 'asgard'),
        (SELECT id FROM member_map WHERE key = 'Iron Man'),
        'cash', 'RUB', 'real', 1,
        g.buy_in_amount, TRUE, g.rebuy_price, 10, '07:00:00',
        g.start_time, g.end_time, 'finished', 4, 10,
        'profit', 'roi',
        NULL, '00:00:00', FALSE
    FROM (VALUES
        (1,  300::NUMERIC, 250::NUMERIC, '2026-05-30 19:00:00+03'::TIMESTAMPTZ, '2026-05-31 02:00:00+03'::TIMESTAMPTZ),
        (2,  450, 300, '2026-06-06 19:00:00+03', '2026-06-07 02:00:00+03'),
        (3,  200, 200, '2026-06-13 19:00:00+03', '2026-06-14 02:00:00+03'),
        (4,  500, 350, '2026-06-20 19:00:00+03', '2026-06-21 02:00:00+03'),
        (5,  350, 250, '2026-06-27 19:00:00+03', '2026-06-28 02:00:00+03'),
        (6,  250, 300, '2026-07-04 19:00:00+03', '2026-07-05 02:00:00+03'),
        (7,  400, 400, '2026-07-11 19:00:00+03', '2026-07-12 02:00:00+03'),
        (8,  300, 250, '2026-07-18 19:00:00+03', '2026-07-19 02:00:00+03'),
        (9,  450, 350, '2026-07-25 19:00:00+03', '2026-07-26 02:00:00+03'),
        (10, 200, 300, '2026-08-01 19:00:00+03', '2026-08-02 02:00:00+03'),
        (11, 500, 250, '2026-08-08 19:00:00+03', '2026-08-09 02:00:00+03'),
        (12, 350, 400, '2026-08-15 19:00:00+03', '2026-08-16 02:00:00+03'),
        (13, 250, 200, '2026-08-22 19:00:00+03', '2026-08-23 02:00:00+03'),
        (14, 400, 300, '2026-08-29 19:00:00+03', '2026-08-30 02:00:00+03'),
        (15, 300, 350, '2026-09-05 19:00:00+03', '2026-09-06 02:00:00+03'),
        (16, 450, 250, '2026-09-12 19:00:00+03', '2026-09-13 02:00:00+03'),
        (17, 200, 400, '2026-09-19 19:00:00+03', '2026-09-20 02:00:00+03'),
        (18, 500, 300, '2026-09-26 19:00:00+03', '2026-09-27 02:00:00+03'),
        (19, 350, 250, '2026-10-03 19:00:00+03', '2026-10-04 02:00:00+03'),
        (20, 400, 350, '2026-10-10 19:00:00+03', '2026-10-11 02:00:00+03')
    ) AS g(n, buy_in_amount, rebuy_price, start_time, end_time)
    RETURNING id
)
INSERT INTO game_map (key, id)
SELECT 'game_' || n, id FROM (
    SELECT ROW_NUMBER() OVER () AS n, id FROM inserted
) x;

-- ============================================================================
-- 🆕 5A. GAMES — PLANNED (1 game)
-- ============================================================================
WITH inserted AS (
    INSERT INTO games (
        club_id, banker_id, game_type, currency, money_model, chip_value,
        buy_in_amount, rebuy_allowed, rebuy_price, max_rebuys, duration,
        start_time, end_time, status, min_players, max_players,
        ranking_primary, ranking_secondary,
        timer_paused_at, timer_paused_duration, timer_notified
    )
    SELECT
        (SELECT id FROM club_map WHERE key = 'asgard'),
        (SELECT id FROM member_map WHERE key = 'Iron Man'),
        'cash', 'RUB', 'real', 1,
        400::NUMERIC, TRUE, 200::NUMERIC, 10, '07:00:00',
        NOW() + INTERVAL '3 days',   -- start_time = сегодня + 3 дня
        NULL,                         -- end_time = NULL для planned
        'planned', 4, 10,
        'profit', 'roi',
        NULL, '00:00:00', FALSE
    RETURNING id
)
INSERT INTO game_map (key, id)
SELECT 'game_planned_1', id FROM inserted;

-- ============================================================================
-- 🆕 5B. GAMES — ACTIVE (1 game)
-- ============================================================================
WITH inserted AS (
    INSERT INTO games (
        club_id, banker_id, game_type, currency, money_model, chip_value,
        buy_in_amount, rebuy_allowed, rebuy_price, max_rebuys, duration,
        start_time, end_time, status, min_players, max_players,
        ranking_primary, ranking_secondary,
        timer_paused_at, timer_paused_duration, timer_notified
    )
    SELECT
        (SELECT id FROM club_map WHERE key = 'asgard'),
        (SELECT id FROM member_map WHERE key = 'Iron Man'),
        'cash', 'RUB', 'real', 1,
        400::NUMERIC, TRUE, 200::NUMERIC, 10, '07:00:00',
        NOW(),                        -- start_time = момент создания
        NULL,                         -- end_time = NULL для active
        'active', 4, 10,
        'profit', 'roi',
        NULL, '00:00:00', FALSE
    RETURNING id
)
INSERT INTO game_map (key, id)
SELECT 'game_active_1', id FROM inserted;

-- ============================================================================
-- 6. GAME PARTICIPANTS — FINISHED
-- ============================================================================
INSERT INTO game_participants (
    game_id, player_id, buy_in_count, rebuy_count,
    chips_end, payout_amount, place, status
)
SELECT
    gm.id,
    pm.id,
    p.buy_in_count,
    p.rebuy_count,
    p.chips_end,
    p.chips_end,
    p.place,
    'confirmed'
FROM (VALUES
    -- Game 1
    (1,  'Thor',          1, 3, 2600, 1),
    (1,  'Iron Man',      1, 5, 2350, 2),
    (1,  'Scarlet Witch', 1, 6, 1500, 3),
    (1,  'Captain',       1, 2, 500,  4),
    (1,  'Spider-Man',    1, 4, 700,  5),
    (1,  'Hulk',          1, 7, 900,  6),
    -- Game 2
    (2,  'Thor',          1, 2, 3500, 1),
    (2,  'Black Widow',   1, 5, 3200, 2),
    (2,  'Iron Man',      1, 8, 3100, 3),
    (2,  'Wasp',          1, 3, 1200, 4),
    (2,  'Hawkeye',       1, 6, 1800, 5),
    (2,  'Ant-Man',       1, 4, 800,  6),
    (2,  'Vision',        1, 7, 800,  7),
    -- Game 3
    (3,  'Iron Man',      1, 4, 1500, 1),
    (3,  'Star-Lord',     1, 3, 1100, 2),
    (3,  'Captain',       1, 9, 2200, 3),
    (3,  'Hulk',          1, 2, 300,  4),
    (3,  'Black Panther', 1, 5, 500,  5),
    -- Game 4
    (4,  'Thor',          1, 3, 3500, 1),
    (4,  'Scarlet Witch', 1, 4, 2500, 2),
    (4,  'Groot',         1, 3, 1650, 3),
    (4,  'Spider-Man',    1, 7, 3200, 4),
    (4,  'Iron Man',      1, 6, 2800, 5),
    (4,  'Ant-Man',       1, 2, 800,  6),
    (4,  'Captain',       1, 5, 1500, 7),
    (4,  'Gamora',        1, 8, 1000, 8),
    -- Game 5
    (5,  'Black Widow',   1, 6, 2600, 1),
    (5,  'Iron Man',      1, 3, 1500, 2),
    (5,  'Wasp',          1, 5, 1800, 3),
    (5,  'Vision',        1, 4, 1200, 4),
    (5,  'Hawkeye',       1, 2, 400,  5),
    (5,  'Hulk',          1, 8, 1600, 6),
    -- Game 6
    (6,  'Thor',          1, 4, 2100, 1),
    (6,  'Ant-Man',       1, 6, 2600, 2),
    (6,  'Iron Man',      1, 7, 2500, 3),
    (6,  'Black Panther', 1, 3, 900,  4),
    (6,  'Gamora',        1, 8, 2350, 5),
    (6,  'Star-Lord',     1, 5, 1400, 6),
    (6,  'Captain',       1, 2, 400,  7),
    -- Game 7
    (7,  'Spider-Man',    1, 5, 3400, 1),
    (7,  'Black Widow',   1, 4, 2400, 2),
    (7,  'Groot',         1, 3, 1800, 3),
    (7,  'Iron Man',      1, 2, 800,  4),
    (7,  'Scarlet Witch', 1, 9, 2800, 5),
    -- Game 8
    (8,  'Thor',          1, 3, 1900, 1),
    (8,  'Wasp',          1, 8, 2800, 2),
    (8,  'Captain',       1, 7, 2400, 3),
    (8,  'Iron Man',      1, 5, 1800, 4),
    (8,  'Black Panther', 1, 3, 1100, 5),
    (8,  'Hawkeye',       1, 4, 900,  6),
    (8,  'Hulk',          1, 2, 300,  7),
    (8,  'Vision',        1, 6, 700,  8),
    -- Game 9
    (9,  'Ant-Man',       1, 6, 3100, 1),
    (9,  'Iron Man',      1, 4, 2200, 2),
    (9,  'Star-Lord',     1, 5, 2400, 3),
    (9,  'Scarlet Witch', 1, 7, 2600, 4),
    (9,  'Gamora',        1, 3, 1150, 5),
    (9,  'Spider-Man',    1, 2, 700,  6),
    -- Game 10
    (10, 'Thor',          1, 2, 1800, 1),
    (10, 'Iron Man',      1, 8, 3000, 2),
    (10, 'Vision',        1, 4, 1400, 3),
    (10, 'Black Widow',   1, 6, 1800, 4),
    (10, 'Hulk',          1, 3, 900,  5),
    (10, 'Groot',         1, 7, 1800, 6),
    (10, 'Captain',       1, 5, 1200, 7),
    -- Game 11
    (11, 'Hawkeye',       1, 4, 2100, 1),
    (11, 'Iron Man',      1, 3, 1500, 2),
    (11, 'Star-Lord',     1, 5, 1550, 3),
    (11, 'Black Panther', 1, 2, 800,  4),
    (11, 'Wasp',          1, 7, 1800, 5),
    -- Game 12
    (12, 'Ant-Man',       1, 8, 4200, 1),
    (12, 'Spider-Man',    1, 3, 2100, 2),
    (12, 'Thor',          1, 4, 2400, 3),
    (12, 'Iron Man',      1, 6, 3000, 4),
    (12, 'Gamora',        1, 2, 1100, 5),
    (12, 'Scarlet Witch', 1, 7, 2800, 6),
    (12, 'Captain',       1, 2, 500,  7),
    (12, 'Black Widow',   1, 5, 1500, 8),
    -- Game 13
    (13, 'Wasp',          1, 8, 2200, 1),
    (13, 'Iron Man',      1, 5, 1500, 2),
    (13, 'Black Panther', 1, 4, 1200, 3),
    (13, 'Groot',         1, 6, 1500, 4),
    (13, 'Vision',        1, 2, 300,  5),
    (13, 'Hulk',          1, 3, 400,  6),
    -- Game 14
    (14, 'Thor',          1, 2, 1600, 1),
    (14, 'Spider-Man',    1, 6, 2600, 2),
    (14, 'Iron Man',      1, 7, 2800, 3),
    (14, 'Star-Lord',     1, 5, 1800, 4),
    (14, 'Hawkeye',       1, 3, 1100, 5),
    (14, 'Captain',       1, 4, 1200, 6),
    (14, 'Gamora',        1, 8, 2200, 7),
    -- Game 15
    (15, 'Ant-Man',       1, 6, 2800, 1),
    (15, 'Iron Man',      1, 2, 1200, 2),
    (15, 'Vision',        1, 3, 1500, 3),
    (15, 'Scarlet Witch', 1, 4, 1500, 4),
    (15, 'Black Widow',   1, 7, 2200, 5),
    -- Game 16
    (16, 'Thor',          1, 3, 2100, 1),
    (16, 'Groot',         1, 3, 1800, 2),
    (16, 'Iron Man',      1, 5, 2000, 3),
    (16, 'Wasp',          1, 8, 2700, 4),
    (16, 'Spider-Man',    1, 4, 1600, 5),
    (16, 'Hulk',          1, 2, 600,  6),
    (16, 'Captain',       1, 7, 1800, 7),
    (16, 'Black Panther', 1, 6, 1500, 8),
    -- Game 17
    (17, 'Iron Man',      1, 4, 2200, 1),
    (17, 'Black Widow',   1, 5, 2600, 2),
    (17, 'Gamora',        1, 3, 1600, 3),
    (17, 'Star-Lord',     1, 6, 2400, 4),
    (17, 'Hawkeye',       1, 8, 3000, 5),
    (17, 'Scarlet Witch', 1, 2, 600,  6),
    -- Game 18
    (18, 'Thor',          1, 2, 1800, 1),
    (18, 'Iron Man',      1, 6, 2600, 2),
    (18, 'Groot',         1, 8, 3200, 3),
    (18, 'Vision',        1, 7, 2400, 4),
    (18, 'Ant-Man',       1, 3, 1100, 5),
    (18, 'Captain',       1, 5, 1600, 6),
    (18, 'Black Panther', 1, 4, 1300, 7),
    -- Game 19
    (19, 'Iron Man',      1, 3, 1400, 1),
    (19, 'Spider-Man',    1, 5, 1900, 2),
    (19, 'Star-Lord',     1, 4, 1300, 3),
    (19, 'Wasp',          1, 2, 600,  4),
    (19, 'Hulk',          1, 7, 1800, 5),
    -- Game 20
    (20, 'Thor',          1, 4, 2400, 1),
    (20, 'Black Widow',   1, 5, 2600, 2),
    (20, 'Iron Man',      1, 7, 3100, 3),
    (20, 'Gamora',        1, 2, 1050, 4),
    (20, 'Hawkeye',       1, 3, 1200, 5),
    (20, 'Ant-Man',       1, 6, 2200, 6),
    (20, 'Captain',       1, 2, 800,  7),
    (20, 'Scarlet Witch', 1, 8, 2800, 8)
) AS p(game_n, player_key, buy_in_count, rebuy_count, chips_end, place)
JOIN game_map   gm ON gm.key = 'game_' || p.game_n
JOIN player_map pm ON pm.key = p.player_key;

-- ============================================================================
-- 🆕 6A. GAME PARTICIPANTS — PLANNED GAME
-- ============================================================================
-- Confirmed:  Thor, Iron Man, Groot
-- Accepted:   Black Widow, Star-Lord, Wasp
-- Invited:    все остальные члены клуба
INSERT INTO game_participants (
    game_id, player_id, buy_in_count, rebuy_count,
    chips_end, payout_amount, place, status
)
SELECT
    (SELECT id FROM game_map WHERE key = 'game_planned_1'),
    pm.id,
    0, 0,
    NULL, NULL, NULL,
    CASE pm.key
        WHEN 'Thor'        THEN 'confirmed'
        WHEN 'Iron Man'    THEN 'confirmed'
        WHEN 'Groot'       THEN 'confirmed'
        WHEN 'Black Widow' THEN 'accepted'
        WHEN 'Star-Lord'   THEN 'accepted'
        WHEN 'Wasp'        THEN 'accepted'
        ELSE                    'invited'
    END
FROM player_map pm;

-- ============================================================================
-- 🆕 6B. GAME PARTICIPANTS — ACTIVE GAME
-- ============================================================================
-- Confirmed: Iron Man, Thor, Captain, Hulk, Spider-Man, Scarlet Witch
INSERT INTO game_participants (
    game_id, player_id, buy_in_count, rebuy_count,
    chips_end, payout_amount, place, status
)
SELECT
    (SELECT id FROM game_map WHERE key = 'game_active_1'),
    pm.id,
    1, 0,           -- buy_in_count=1, rebuy_count=0
    NULL, NULL, NULL,
    'confirmed'
FROM (VALUES
    ('Iron Man'),
    ('Thor'),
    ('Captain'),
    ('Hulk'),
    ('Spider-Man'),
    ('Scarlet Witch')
) AS p(player_key)
JOIN player_map pm ON pm.key = p.player_key;

-- ============================================================================
-- 🆕 6C. EVENTS (журнал событий)
-- ============================================================================
-- Event 1: Game 3, Star-Lord, chips_end 900 → 1100
-- Event 2: Game 4, Gamora,   chips_end 1200 → 1000
-- Event 3: Game 7, Groot,    chips_end 1600 → 1800
-- created_by = Thor (owner)
INSERT INTO events (
    game_id, player_id, type,
    old_value, new_value, metadata,
    created_at, created_by
)
SELECT
    gm.id,
    pm.id,
    'correction',
    e.old_value,
    e.new_value,
    jsonb_build_object(
        'field',          'chips_end',
        'old_chips_end',  e.old_value,
        'new_chips_end',  e.new_value
    ),
    NOW(),
    (SELECT id FROM member_map WHERE key = 'Thor')
FROM (VALUES
    (3, 'Star-Lord', 900::NUMERIC,  1100::NUMERIC),
    (4, 'Gamora',    1200::NUMERIC, 1000::NUMERIC),
    (7, 'Groot',     1600::NUMERIC, 1800::NUMERIC)
) AS e(game_n, player_key, old_value, new_value)
JOIN game_map   gm ON gm.key = 'game_' || e.game_n
JOIN player_map pm ON pm.key = e.player_key;

-- ============================================================================
-- 7. PLAYER STATISTICS (computed from finished games only)
-- ============================================================================
INSERT INTO player_statistics (
    player_id, club_id,
    total_games, total_buy_in_amount, total_rebuy_amount, total_rebuys_count,
    total_invested, total_chips, total_profit,
    biggest_win, biggest_loss, games_won, podiums, roi, itm
)
SELECT
    pm.id,
    (SELECT id FROM club_map WHERE key = 'asgard'),
    COUNT(*)::INT,
    SUM(g.buy_in_amount)::NUMERIC(12,2),
    SUM(g.rebuy_price * gp.rebuy_count)::NUMERIC(12,2),
    SUM(gp.rebuy_count)::INT,
    SUM(g.buy_in_amount + g.rebuy_price * gp.rebuy_count)::NUMERIC(12,2),
    SUM(gp.chips_end)::NUMERIC(12,2),
    SUM(gp.payout_amount - (g.buy_in_amount + g.rebuy_price * gp.rebuy_count))::NUMERIC(12,2),
    GREATEST(MAX(gp.payout_amount - (g.buy_in_amount + g.rebuy_price * gp.rebuy_count)), 0)::NUMERIC(12,2),
    LEAST(MIN(gp.payout_amount - (g.buy_in_amount + g.rebuy_price * gp.rebuy_count)), 0)::NUMERIC(12,2),
    COUNT(*) FILTER (WHERE gp.place = 1)::INT,
    0,
    CASE
        WHEN SUM(g.buy_in_amount + g.rebuy_price * gp.rebuy_count) > 0
        THEN ROUND(
            SUM(gp.payout_amount - (g.buy_in_amount + g.rebuy_price * gp.rebuy_count)) * 100.0
            / SUM(g.buy_in_amount + g.rebuy_price * gp.rebuy_count),
            4
        )
        ELSE 0
    END,
    0
FROM game_participants gp
JOIN games  g  ON g.id = gp.game_id
JOIN player_map pm ON pm.id = gp.player_id
WHERE g.status = 'finished'   -- только завершённые игры
GROUP BY pm.id;

COMMIT;
