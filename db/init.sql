CREATE TABLE champions (
  name varchar(64) PRIMARY KEY,
  slug varchar(64),
  url varchar(128)
);

CREATE TABLE users (
  display_name varchar(64),
  username varchar(64) PRIMARY KEY,
  assigned_role varchar(32)
);

CREATE TABLE user_to_champions (
    username VARCHAR(64) NOT NULL,
    champion_name VARCHAR(64) NOT NULL,
    PRIMARY KEY (username, champion_name),

    CONSTRAINT fk_user_to_champions_user
        FOREIGN KEY (username)
        REFERENCES users(username)
        ON UPDATE CASCADE
        ON DELETE CASCADE

);

-- I want to make this change:
-- CREATE TABLE users (
--     username VARCHAR(64) PRIMARY KEY,
--     display_name VARCHAR(64) NOT NULL
-- );

-- CREATE TABLE roles (
--     role_name VARCHAR(32) PRIMARY KEY
-- );

-- CREATE TABLE champions (
--     name VARCHAR(64) PRIMARY KEY,
--     slug VARCHAR(64),
--     url VARCHAR(128)
-- );

-- CREATE TABLE user_role_champions (
--     username VARCHAR(64) NOT NULL,
--     role_name VARCHAR(32) NOT NULL,
--     champion_name VARCHAR(64) NOT NULL,

--     PRIMARY KEY (username, role_name, champion_name),

--     FOREIGN KEY (username)
--         REFERENCES users(username)
--         ON UPDATE CASCADE
--         ON DELETE CASCADE,

--     FOREIGN KEY (role_name)
--         REFERENCES roles(role_name),

--     FOREIGN KEY (champion_name)
--         REFERENCES champions(name)
--         ON UPDATE CASCADE
--         ON DELETE CASCADE
-- );