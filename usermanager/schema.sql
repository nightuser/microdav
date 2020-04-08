create table "users" (
  "username" text not null unique,
  "salt" text not null,
  "password" text not null
)
