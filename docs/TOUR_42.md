<b><a href="./TOUR_0.html">Some Practical Rye</a> > Sqlite and Postgres</b>

# Sqlite and Postgres

SQL dialect is a dialect of Rye values (not string) tha generates prepared SQL
statements withouth the mistake-prone need for positional arguments.

## SQLite

```rye
title: "some title"
content: "... content ..."

db: open sqlite://file.db
exec db { insert into notes ( title , content ) values ( ?title , ?content ) }

query db { select * from notes } |print
// prints:
// |title     |content        |
// |some title|... content ...|
```

## PostgreSQL

```rye
id: 101
db: open postgres://user1:password@demo
generate-token 32 :tok
exec db
{ insert into tokens ( id_user , token ) values ( ?id , ?tok )
  on conflict ( id_user ) do update set token = ?tok }
```
