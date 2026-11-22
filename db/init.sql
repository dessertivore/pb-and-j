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
