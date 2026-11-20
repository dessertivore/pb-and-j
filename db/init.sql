CREATE TABLE Champions (
  name varchar(64) PRIMARY KEY,
  slug varchar(64),
  url varchar(128)
);

CREATE TABLE Users (
  display_name varchar(64),
  username varchar(64) PRIMARY KEY,
  preferred_champions varchar(64)[],
  assigned_role varchar(64)
);
