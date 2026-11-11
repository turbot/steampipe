# Postgres config files and database auth

Status: proposed. It is built in two steps. See [Rollout](#rollout).

## Summary

Steampipe will manage every Postgres config file with one rule:

- Steampipe writes its own files before every database start. Nobody edits them.
- Your changes go in a `.d` folder. Steampipe never changes that folder.

Today `pg_hba.conf` is the only file that breaks this rule. Step 1 brings it into line. Step 2
builds on step 1: only the OS user who runs Steampipe can connect as `root`, the database
superuser.

## The files

Postgres reads these files from the data directory, `~/.steampipe/db/<version>/data/`.

| File | What it controls | Today |
|---|---|---|
| `postgresql.conf` | Server settings. It includes `steampipe.conf`, then `postgresql.conf.d/*.conf`. | Steampipe writes it before every start. |
| `steampipe.conf` | Steampipe's default settings: time zone, logging, locks. | Steampipe writes it before every start. |
| `postgresql.conf.d/` | Your settings. They win over `steampipe.conf`. | Steampipe creates the folder. You own the files in it. |
| `pg_hba.conf` | Who can connect, from where, and how they prove who they are. | Steampipe writes it once, at install. It never writes it again. |
| `pg_ident.conf` | A table that maps OS user names to database user names. | `initdb` creates it empty. Nothing uses it. |
| `postgresql.auto.conf` | Settings made with `ALTER SYSTEM`. | `initdb` creates it. Steampipe never touches it. |
| `root.crt`, `server.crt` and their keys | SSL certificates. | Steampipe creates them when they are missing, and renews the ones it issued. It never touches certificates that you supply. |

Steampipe also gives postgres some settings on its command line (`-c`), for example the port
and the listen addresses. A command-line setting wins over every file.

### pg_hba.conf and pg_ident.conf

`pg_hba.conf` is a list of rules. For each new connection, Postgres reads the rules from the
top and uses the first rule that matches. A rule has five parts:

```
# type  database  user       address   method
host    all       steampipe  samehost  trust
```

- `host` matches a TCP connection. `local` matches a connection through the unix socket file.
- `samehost` matches a connection that comes from one of this machine's own IP addresses.
- `trust` lets the connection in with no password. `scram-sha-256` asks for a password.

`pg_ident.conf` is used only by rules with the `peer` or `ident` method and a `map=` option.
These methods ask the OS which user opened the connection, and `pg_ident.conf` changes that OS
user name to a database user name. This design does not use them. So `pg_ident.conf` stays as
`initdb` wrote it.

## What is wrong today

1. **pg_hba.conf has a rule of its own.** Steampipe writes it at install and never again.
   - A change to the access rules never gets to an existing install.
   - There is no safe place for your own rules. A platform that runs Steampipe must copy its
     whole file over Steampipe's file before every start.
   - The install writes the file two times: a root-only file first, then the full file.
   - Steampipe uses "`pg_hba.conf` exists" to mean "the database is installed".
2. **root trusts every connection from this machine's addresses.** `samehost` tests the
   network address. It does not test who you are. All of these connect from this machine's
   addresses:
   - other OS users on the same machine
   - SSH port forwards, and tunnels such as ngrok
   - containers on Docker Desktop
   
   `root` is the superuser, and a superuser can run OS commands as the user who runs Steampipe.
   Only Steampipe itself connects as `root`, so this trust gives users nothing.
3. **Three ways to start postgres.** The service start, the install, and the backup during a
   database upgrade each build their own postgres command line. Each writes different config
   files: the service start writes `postgresql.conf`, the install writes `pg_hba.conf`, and the
   backup writes none.

## Design

### One rule for every file

```
                         before every database start
  pg_hba.conf.d/*.conf  ──┐
  (yours)                 ├──▶  Steampipe writes pg_hba.conf
  Steampipe's rules     ──┘

  postgresql.conf.d/*.conf ──┐
  (yours)                    ├──▶  postgresql.conf includes both
  steampipe.conf           ──┘     (Steampipe writes these, as it does today)
```

| Kind | Files | Who writes them | When |
|---|---|---|---|
| Steampipe's | `postgresql.conf`, `steampipe.conf`, `pg_hba.conf` | Steampipe | Before every start. Each starts with "DO NOT EDIT". |
| Yours | `postgresql.conf.d/*.conf`, `pg_hba.conf.d/*.conf` | You, or a platform | Steampipe never changes them. |
| Command line | Port, listen addresses, socket, SSL files | Steampipe | Every start. No file can change them. |
| Postgres's | `pg_ident.conf`, `postgresql.auto.conf` | Postgres | Steampipe never touches them. |
| Certificates | `root.crt`, `server.crt` and their keys | Steampipe, or you | As today. |

### How Steampipe writes pg_hba.conf

Before every start, Steampipe writes `pg_hba.conf` from three parts, in this order:

1. A header that marks the file as Steampipe's.
2. Your rules: each file in `pg_hba.conf.d/` whose name ends in `.conf`, in name order.
   Steampipe skips names that start with `.`, as Postgres does for `include_dir`.
3. Steampipe's rules.

Your rules come first, so they win. This is the same as `postgresql.conf.d/`, where your
settings also win.

Example: you want the `steampipe` user to give a password, also on this machine. You add the
file `pg_hba.conf.d/10-password.conf`:

```
host all steampipe samehost scram-sha-256
```

After step 2, Steampipe writes this `pg_hba.conf`:

```
# Written by Steampipe each time the database starts. DO NOT EDIT.
# Put your own rules in pg_hba.conf.d/<name>.conf. They come first, so they win.

# from pg_hba.conf.d/10-password.conf
host all steampipe samehost scram-sha-256

# Steampipe's rules
local all root      trust
host  all steampipe samehost trust
host  all steampipe all      scram-sha-256
```

Your rule matches first, so the `steampipe` user must give a password from this machine.

Two changes to Steampipe's rules, compared with today:

- **The database column is `all`, not the database name.** The grants already limit the
  `steampipe` user to its own database: the install revokes CONNECT from PUBLIC on every
  database, and grants it on the Steampipe database only. So the file does not need the
  database name, and Steampipe can write it before the database starts.
- **The `hostssl` lines go.** Each one was a copy of a `host` line, and `host` matches SSL
  connections too.

Postgres 16 can include files in `pg_hba.conf`. Steampipe runs Postgres 14, so Steampipe joins
the files itself. When Steampipe moves to Postgres 16 or later, it can use `include_dir`
instead. Users will see no change.

### root: the unix socket, for the OS user who runs Steampipe

```
before                                          after
CLI ──────── TCP 127.0.0.1 ──▶ root   ok        CLI ──── /tmp/.s.PGSQL.9193 ──▶ root   ok
other user ─ TCP 127.0.0.1 ──▶ root   ok        other user ── socket ──▶ permission denied
container ── TCP 127.0.0.1 ──▶ root   ok        container ─── TCP ─────▶ root   rejected
tunnel ───── TCP 127.0.0.1 ──▶ root   ok        tunnel ────── TCP ─────▶ root   rejected
```

- Steampipe's `root` rule is `local all root trust`. There is no `host` rule for `root`, so
  Postgres rejects `root` over TCP.
- Steampipe starts postgres with two more command-line settings:
  - `unix_socket_directories=/tmp`. Postgres already puts its socket there today. The
    setting makes this explicit, so the CLI and the server always agree.
  - `unix_socket_permissions=0700`. Only the OS user who runs postgres can open the socket.
- Every `root` connection from Steampipe uses the socket. This includes the plugin manager, and
  `pg_dump` and `pg_restore` during a database upgrade. Connections through the socket do not
  use SSL.
- Before Steampipe connects as `root`, it checks that the socket file is a socket that the
  current OS user owns. `/tmp` is shared, so another user can make a file with that name before
  postgres starts. Steampipe must not send `root` commands to that user's server. While
  postgres is starting, a missing socket file means "wait". A socket file that another user
  owns means "stop with an error".

Why this, and not one of the other ways:

- **A password for root.** It needs a secret on disk, a way to recover when that file is lost,
  and the password in every `root` connection. The socket needs none of these.
- **`peer` auth with `pg_ident.conf`.** Steampipe must write the OS user name into
  `pg_ident.conf`, and Steampipe and Postgres must find the same name. It is a second file to
  manage. The socket permission gives the same result with one setting.
- **A socket in the data directory.** That folder is private, but macOS limits a socket path
  to 103 bytes. With a long install directory, the database would not start. It also moves a
  socket that tools use in `/tmp` today.

### The steampipe user does not change

The `steampipe` user keeps `samehost trust`. Everything that connects from this machine's
addresses can still run queries with your cloud credentials and no password. This includes the
other OS users, tunnels and containers in the list above. This is a deliberate choice:
Powerpipe, Flowpipe and local tools connect with no setup. The docs must say this clearly.

### Existing installs

Before Steampipe writes `pg_hba.conf`, it examines the file that is there:

| The file that is there | What Steampipe does |
|---|---|
| It starts with Steampipe's header. | Writes the new file over it. |
| It has only rules that an older Steampipe wrote. | Writes the new file over it. |
| Anything else: you or a platform changed it. | Moves it to `pg_hba.conf.d/00-preserved.conf`, over any older copy, and logs a warning. Then it writes the new file, which puts those rules first. |

"Rules that an older Steampipe wrote" means this: after Steampipe removes comments and blank
lines, each line is one of these three, with any database name and any amount of white space:

```
host|hostssl  all   root       samehost  trust
host|hostssl  <db>  steampipe  samehost  trust
host|hostssl  <db>  steampipe  all       scram-sha-256
```

`initdb` also writes a `pg_hba.conf`, which trusts every local connection. The install deletes
it immediately after `initdb`, so that the first start does not keep it as your rules.

Examples:

- **A default install.** The file has only old rules. Steampipe writes over it. After step 2,
  `root` over TCP stops.
- **You added a rule by hand.** Steampipe keeps your whole file as your rules, with its `root`
  lines. Nothing that you set up stops working. To close `root` over TCP, remove those lines
  from `pg_hba.conf.d/00-preserved.conf`.
- **A platform copies its own `pg_hba.conf` before each start.** Steampipe moves it to
  `pg_hba.conf.d/00-preserved.conf` each time, so the platform's rules still apply. The
  platform should copy its file into `pg_hba.conf.d/` instead.

### Every postgres start writes pg_hba.conf

The service start, the install, and the backup during a database upgrade all start postgres.
All three:

- write `pg_hba.conf` first, and
- get the socket settings from one shared list.

The backup starts the old postgres version on the old data directory. That directory can
still have the `pg_hba.conf` of an older Steampipe. If the backup did not write the file first,
it could not connect as `root`.

`postgresql.conf` and `steampipe.conf` stay as today: the service start writes them.

### What goes away

- The root-only `pg_hba.conf` that the install writes first, and the second write after it.
- The database name in `pg_hba.conf`.
- The `hostssl` copies of the `host` lines.
- The random `root` password that the install sets. Nothing keeps it or uses it.
- "`pg_hba.conf` exists" as the test for "the database is installed". Use `PG_VERSION` in the
  data directory, which `initdb` writes.

## Cases

| Case | Before | After step 2 |
|---|---|---|
| Steampipe CLI and plugin manager as `root` | TCP | Socket |
| `steampipe` user over TCP from this machine (Powerpipe, Flowpipe, psql, BI tools) | No password | No password |
| `steampipe` user from another machine | Password | Password |
| `root` from another OS user, a tunnel or a container | Allowed | Rejected |
| `root` by hand with `psql -h localhost -U root` | Allowed | Rejected. Use `psql -h /tmp -p 9193 -U root postgres`. |
| A hand-edited `pg_hba.conf` | Kept, never updated | Kept, as `pg_hba.conf.d/00-preserved.conf` |
| A platform that copies its own `pg_hba.conf` before each start | Works | Works. Steampipe moves the file each start. |
| Database upgrade on an install that an older Steampipe ran | Works | Works. The backup writes `pg_hba.conf` first. |
| Custom database name (`STEAMPIPE_INITDB_DATABASE_NAME`) | The name is in `pg_hba.conf` | `all`, limited by the grants |
| An older Steampipe on an install that this version used | Works | `service start` and `service stop` fail. See [Rollout](#rollout). |
| Steampipe commands that do not share `/tmp` with the service (for example systemd `PrivateTmp=yes`) | Works | `root` commands cannot get to the socket |
| Two installs on the same port with different listen addresses | The `/tmp` sockets clash | No change |

## Testing

Unit tests:

- `pg_hba.conf` is the header, then the rules from `pg_hba.conf.d/` in name order, then
  Steampipe's rules. Other files, and names that start with `.`, are skipped. With no folder,
  the file has only Steampipe's rules.
- These files are written over: the old default file (with the default and with a custom
  database name), the old root-only file, and a file with the new header. A file with any other
  rule moves to `00-preserved.conf`.
- All three postgres starts get the socket settings.
- `root` connections use the socket and no SSL. `steampipe` connections do not change.
- The owner check refuses a path that another user owns (a path that `root` owns, such as `/`,
  is enough), and a path that is not a socket.

Acceptance tests go in a new test file that does not stop every service on the machine, so that
it can run on a developer's machine and in CI:

- `root` over TCP is rejected. `root` over the socket works.
- Another OS user cannot connect as `root` over the socket. Skip this test when `sudo -n` does
  not work. It works in CI.
- The `steampipe` user connects over TCP from this machine with no password. It cannot connect
  to the `postgres` database.
- An install with the old default `pg_hba.conf` gets the new file at its next start.
- A fresh install has no `pg_hba.conf.d/00-preserved.conf`.
- A hand-edited `pg_hba.conf` moves to `pg_hba.conf.d/00-preserved.conf`, and its rules still
  apply.
- A rule in `pg_hba.conf.d/` wins over Steampipe's rule.

Manual checks before the release, because CI runs on Linux only:

- macOS: another OS user cannot open the socket.
  `sudo -u nobody psql -h /tmp -p 9193 -U root postgres` must fail.
- Docker Desktop: a container cannot connect as `root` through `host.docker.internal`. It can
  still connect as `steampipe`.

## Rollout

Two pull requests, in this order, in the next minor release:

1. **pg_hba.conf works like postgresql.conf.** Steampipe writes it before every start, reads
   `pg_hba.conf.d/`, keeps changed files, and drops the database name and the `hostssl` lines.
   Steampipe's rules give the same access as today, so no one's access changes:

   ```
   host all root      samehost trust
   host all steampipe samehost trust
   host all steampipe all      scram-sha-256
   ```
2. **root over the socket.** Change Steampipe's `root` rule. Add the socket settings and the
   owner check. Move every `root` connection to the socket. Remove the random `root` password.

Release notes for step 2:

- Only the OS user who runs Steampipe can connect as the database superuser (`root`), and only
  through the unix socket in `/tmp`.
- An older Steampipe cannot start or stop the service on an install that this version used. To go back to an older version, add `host all root samehost trust` as the first
  rule in `pg_hba.conf`. Older versions never write that file again.

Docs:

- The `STEAMPIPE_DATABASE_PASSWORD` page says that `root` can log in from this machine with no
  password. Change it to show how to connect as `root` through the socket.
- Document `pg_hba.conf.d/`. Say that the `steampipe` user needs no password from this machine.

## Out of scope

- A password for the `steampipe` user from this machine. That is a product decision. It would
  stop the no-setup connection that Powerpipe and Flowpipe use.
- SSL-only connections.
- `pg_ident.conf`.
