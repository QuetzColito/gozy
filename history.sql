CREATE DATABASE cozy;
CREATE table list_items (
 name VARCHAR(200),
 index BIGINT,
 list BIGINT,
 color VARCHAR(50),
 decorator VARCHAR(50)
);

ALTER TABLE list_items ALTER COLUMN name SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN index SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN list SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN color SET NOT NULL;
ALTER TABLE list_items ALTER COLUMN decorator SET NOT NULL;

CREATE EXTENSION pgcrypto;

CREATE TABLE users (
 id SERIAL PRIMARY KEY,
 name VARCHAR(200) NOT NULL,
 password VARCHAR(100) NOT NULL
);

-- INSERT INTO users (name, password) VALUES($ADMIN, crypt('$ADMINPASS', gen_salt('md5')))

CREATE TABLE lists (
  id SERIAL PRIMARY KEY,
  owner INTEGER REFERENCES USERS ON DELETE CASCADE,
  name VARCHAR(200) NOT NULL,
  sub_names JSON NOT NULL,
  colors JSON NOT NULL,
  decorators JSON NOT NULL
);

INSERT INTO lists (owner, name, sub_names, colors, decorators) VALUES ((SELECT id FROM users),
  'Anime',
  '["Watchable","Watching","Watched"]',
  '{"":"fg","Romance":"red","RomCom":"orange","Alternate History":"yellow","Fantasy":"green","Urban Fantasy":"blue","Sci-Fi":"cyan","Drama":"purple"}',
  '{"":"","Trash":"🗑️","Bad":"👎","Mid":"🤏","Good":"👍","Very Good":"⭐️","Best":"✨"}'
);

ALTER TABLE list_items ADD COLUMN owner integer REFERENCES LISTS ON DELETE CASCADE;
UPDATE list_items SET owner = (SELECT id FROM lists);

ALTER TABLE users ADD CONSTRAINT unique_username UNIQUE (name);

--> server state
--> local state
