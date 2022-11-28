# Permissions PoC

This repo aims to serve as a proof of concept for the proposed permission system.

## Setup

In order to run this project, you'll have to create a Postgres DB for the permissions table. You should be able to do this with `psql`:

```psql
psql

CREATE DATABASE perms;
```

Next you need to point the project to this Database with a postgres connection URL. There's a default URL defined
in the project, however you can define yours via an environment variable.

```sh
export DATABASE_URL=postgresql://username:password@localhost:5432/permissions_test?sslmode=disable
```

obviously substituting the username/password/db host/database name as necessary.

## Running

You should now be ready to just run the project:

```sh
go run .
```

[Walkthrough]()