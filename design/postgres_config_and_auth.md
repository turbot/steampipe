# Postgres config files and database auth

Status: proposed. Part of it is on develop already. See [Rollout](#rollout).

## Summary

- Steampipe writes its own Postgres config files before every database start. Nobody edits
  them. Your changes go in a `.d` folder, and Steampipe never changes that folder.
- `pg_hba.conf` follows the same rule, with one exception: if you manage `pg_hba.conf`
  yourself, Steampipe never changes it. When your file stops Steampipe from working, Steampipe
  stops and tells you what to do. It never changes your file to fix it.
- Only the OS user who runs Steampipe can connect as `root`, the database superuser. `root`
  connects through the unix socket in the data directory.

## The files

Postgres reads these files from the data directory, `~/.steampipe/db/<version>/data/`.

| File | What it controls | Steampipe 2.4 and earlier |
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
top and uses the first rule that matches. If no rule matches, Postgres refuses the connection.
A rule has five parts:

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

Postgres reads `pg_hba.conf` only when it starts, or when someone tells it to reload. Steampipe
never tells it to reload.

## What is wrong in Steampipe 2.4 and earlier

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
   database upgrade each build their own postgres command line, and each writes different
   config files.

## Design

### One rule for every file

```
                         before every database start
  pg_hba.conf.d/*.conf  ──┐
  (yours)                 ├──▶  Steampipe writes pg_hba.conf
  Steampipe's rules     ──┘     (unless you manage pg_hba.conf yourself)

  postgresql.conf.d/*.conf ──┐
  (yours)                    ├──▶  postgresql.conf includes both
  steampipe.conf           ──┘     (Steampipe writes these, as it does today)
```

| Kind | Files | Who writes them | When |
|---|---|---|---|
| Steampipe's | `postgresql.conf`, `steampipe.conf`, `pg_hba.conf` | Steampipe | Before every start. Each starts with "DO NOT EDIT". |
| Yours | `postgresql.conf.d/*.conf`, `pg_hba.conf.d/*.conf` | You, or a platform | Steampipe never changes them. |
| Yours, if you take it over | `pg_hba.conf` without Steampipe's header | You, or a platform | Steampipe never changes it. |
| Command line | Port, listen addresses, socket, SSL files | Steampipe | Every start. No file can change them. |
| Postgres's | `pg_ident.conf`, `postgresql.auto.conf` | Postgres | Steampipe never touches them. |
| Certificates | `root.crt`, `server.crt` and their keys | Steampipe, or you | As today. |

The data directory holds only these files and Postgres's own data. Your connection files,
plugins, logs and Steampipe's runtime state stay in their own folders under the install
directory, as today.

### How Steampipe writes pg_hba.conf

Steampipe writes `pg_hba.conf` from three parts, in this order:

1. A header. Its first line marks the file as Steampipe's. Its last line is a checksum of every
   line after it.
2. Your rules: each file in `pg_hba.conf.d/` whose name ends in `.conf`, in name order.
   Steampipe skips names that start with `.`, as Postgres does for `include_dir`. Each file's
   rules follow a `# from pg_hba.conf.d/<name>` line.
3. Steampipe's rules.

Your rules come first, so they win. This is the same as `postgresql.conf.d/`, where your
settings also win.

Example: you want the `steampipe` user to give a password, also on this machine. You add the
file `pg_hba.conf.d/10-password.conf`:

```
host all steampipe samehost scram-sha-256
```

Steampipe writes this `pg_hba.conf`:

```
# Written by Steampipe each time the database starts. DO NOT EDIT.
# Put your own rules in pg_hba.conf.d/<name>.conf. They come first, so they win.
# To manage this file yourself, delete these header lines. Steampipe then never changes it.
# checksum: sha256:<hex digest of every line after this one>

# from pg_hba.conf.d/10-password.conf
host all steampipe samehost scram-sha-256

# Steampipe's rules
local all root      trust
host  all steampipe samehost trust
host  all steampipe all      scram-sha-256
```

Your rule matches first, so the `steampipe` user must give a password from this machine.

Two changes to Steampipe's rules, compared with Steampipe 2.4:

- **The database column is `all`, not the database name.** The grants already limit the
  `steampipe` user to its own database: the install revokes CONNECT from PUBLIC on the
  `postgres` and `template1` databases, and on the Steampipe database it grants CONNECT only to
  the Steampipe users. So the file does not need the database name, and Steampipe can write it
  before the database starts. A database that someone creates later by hand gives CONNECT to
  PUBLIC, so the `steampipe` user can connect to it.
- **The `hostssl` lines go.** Each one was a copy of a `host` line, and `host` matches SSL
  connections too.

Postgres 16 can include files in `pg_hba.conf`. Steampipe runs Postgres 14, so Steampipe joins
the files itself.

Steampipe writes the file to a temporary file in the data directory, then renames it to
`pg_hba.conf`. Steampipe 2.4 and earlier treat a missing `pg_hba.conf` as "not installed", and
they then delete everything in the data directory and run `initdb` again. So `pg_hba.conf` must never be
missing, also after a crash. The install writes Steampipe's file immediately after `initdb`,
over the file that `initdb` writes.

### Who owns pg_hba.conf

Before every start, Steampipe examines the `pg_hba.conf` that is there. The header decides who
owns it:

| The file that is there | Owner | What Steampipe does |
|---|---|---|
| None (a fresh install) | Steampipe | Writes it. |
| It starts with Steampipe's header, and the checksum matches. | Steampipe | Writes it again. |
| It is exactly a file that an older Steampipe wrote. | Steampipe | Writes the new file over it. |
| It starts with Steampipe's header, but the checksum does not match: someone edited it. | Not clear | Stops before postgres starts. See [When Steampipe stops](#when-steampipe-stops). |
| Anything else. | You | Never changes it. |

"Exactly a file that an older Steampipe wrote" means this: after Steampipe removes comments and
blank lines, and treats any amount of white space as one space, the rules are one of these two
lists, in this order. `<db>` is the same database name on every line.

```
hostssl all  root      samehost trust          hostssl all root samehost trust
host    all  root      samehost trust          host    all root samehost trust
hostssl <db> steampipe samehost trust
host    <db> steampipe samehost trust          (the root-only file that an
hostssl <db> steampipe all      scram-sha-256   interrupted install left)
host    <db> steampipe all      scram-sha-256
```

A file that has more rules, fewer rules, changed rules or rules in another order is yours.
Steampipe cannot tell which of its rules belong to you: you can have changed one of them, or
deleted one to block a connection, and order changes what the rules mean. So it keeps your
whole file, and changes nothing in it.

### Your own pg_hba.conf

When `pg_hba.conf` is yours:

- Steampipe never writes, moves or adds to it. Postgres reads it as it is.
- Steampipe does not use `pg_hba.conf.d/`. If that folder has `.conf` files, Steampipe shows a
  warning at start that names them and says that they are not used.
- Your file must let Steampipe connect as `root` through the unix socket. Steampipe needs this
  rule, or one like it:

  ```
  local all root trust
  ```

  Steampipe 2.4 and earlier never wrote a `local` rule. So a file that you edited for one of
  those versions does not have it, and Steampipe stops with the message below at its first
  start. A platform that already has a `local` rule for `root` keeps working with no change.

To give the file back to Steampipe, delete it. Steampipe writes its own file at the next start.

### When Steampipe stops

Steampipe stops in two cases:

1. **Steampipe's file was edited.** Steampipe finds this before postgres starts, so postgres
   does not start.
2. **Postgres refuses `root` through the socket.** Steampipe can learn this only by trying. It
   starts postgres, connects as `root`, and gets an authentication error (SQLSTATE class 28).
   It then stops postgres with a signal, which needs no database connection.

In both cases Steampipe has changed nothing and leaves nothing running. The command fails with
exit code 32 (`ExitCodeServiceStartupFailure`), and the error goes to stderr. You fix the file
and run the same command again.

The message shows the choices, with the real paths, and the lines that are yours. Example for
case 2 with your own file:

```
Error: pg_hba.conf has rules that Steampipe did not write, and they do not let
Steampipe connect. Steampipe will not change your file. Choose one:

  1. Move your rules into Steampipe's setup (recommended)
     Put your rules in pg_hba.conf.d/10-mine.conf, then delete pg_hba.conf.
     Steampipe writes a new pg_hba.conf at the next start.

  2. Keep managing pg_hba.conf yourself
     Add this line to it. Steampipe never changes a pg_hba.conf it did not write.
         local all root trust
     If your file lets root connect over TCP, it still will.

  3. Use Steampipe's defaults
     Delete pg_hba.conf. Your rules are lost.

Your rules (the lines that older Steampipe versions did not write):
    host all steampipe samehost scram-sha-256
    host all reporting 10.0.0.0/8 scram-sha-256

File: /Users/me/.steampipe/db/14.19.0/data/pg_hba.conf
```

"Your rules" are the rule lines that are not one of the older Steampipe rules in the lists
above.

For case 1, the choices are:

1. Move your change to `pg_hba.conf.d/<name>.conf`, then delete `pg_hba.conf` (recommended).
2. Delete the header lines, to manage `pg_hba.conf` yourself.
3. Delete `pg_hba.conf`, to discard your change. Steampipe writes it again at the next start.

The message shows the rule lines that are in the file but that Steampipe would not write now,
and the lines of Steampipe's that are missing from the file.

Case 2 can also happen when the file is Steampipe's, because a rule in `pg_hba.conf.d/` refuses
`root` (for example `local all all reject`). Then the message says that a rule in
`pg_hba.conf.d/` refuses `root`, and lists the files there.

Steampipe does not ask a question at start. `steampipe query` starts the database in the
background while the query prompt owns the terminal, and many starts have no terminal at all:
batch queries, systemd, containers, CI. One message works in all of them.

### Where the check runs

The check runs each time Steampipe starts postgres:

```
1. Examine pg_hba.conf, before postgres starts
     Steampipe's file, but edited  → stop now. Postgres never starts.
     Your file                     → leave it alone
     Steampipe's, old default,
       or none                     → write it
2. Start postgres
3. Connect as root through the socket
     works                         → continue (plugin manager, connection refresh)
     refused                       → stop postgres with a signal, show the message
```

| Situation | What happens |
|---|---|
| `steampipe service start` | The check runs. If it fails, Steampipe shows the message and exits with code 32. |
| `steampipe service start` when the service is running | "Already running". No check. No change. |
| `steampipe service restart` | Steampipe examines `pg_hba.conf` **before it stops the service**. If Steampipe's file was edited, it refuses and the service keeps running. If not, it stops the service and starts it as above. |
| `steampipe query` (interactive), no service running | The database starts in the background. If the check fails, the prompt closes and Steampipe shows the message, as it does for any start error today. |
| `steampipe query "<sql>"` (batch), no service running | The check runs. The message goes to stderr, and the exit code is not 0. Nothing goes to stdout. |
| `steampipe query`, service running | Connects as `steampipe` over TCP. No check. |
| Powerpipe, Flowpipe, psql and other tools | They connect as `steampipe` to a running service and never start it. No check. When no service runs, they get "connection refused", as today. |
| Plugin manager, connection refresh | `root` through the socket, from the start that has just shown that `root` works. |
| `steampipe service stop` | No check, because nothing starts. `root` connects through the socket to count the clients. |
| systemd, containers, CI | As for a batch query. The message goes to the service log. A restart loop fails the same way each time, and changes nothing. |
| Database upgrade to a new Postgres version | The backup starts the old postgres on the old data directory, so the check runs on the old `pg_hba.conf` before anything moves. If it fails, nothing is migrated, and the message names the old file. |
| Fresh install | Steampipe writes the file. There is nothing to examine. |

An edit to `pg_hba.conf` while the service runs has no effect until the next start, and the
check runs then.

### root: the unix socket in the data directory

```
before                                          after
CLI ──────── TCP 127.0.0.1 ──▶ root   ok        CLI ──── <data dir>/.s.PGSQL.9193 ──▶ root   ok
other user ─ TCP 127.0.0.1 ──▶ root   ok        other user ── socket ──▶ permission denied
container ── TCP 127.0.0.1 ──▶ root   ok        container ─── TCP ─────▶ root   rejected
tunnel ───── TCP 127.0.0.1 ──▶ root   ok        tunnel ────── TCP ─────▶ root   rejected
```

- Steampipe's `root` rule is `local all root trust`. There is no `host` rule for `root`, so
  Postgres rejects `root` over TCP.
- Steampipe starts postgres with two more command-line settings:
  - `unix_socket_directories=<data directory>`
  - `unix_socket_permissions=0700`
- Postgres does not start unless the data directory belongs to the user who runs it and has
  mode `0700`. So only that user can open the socket, and nobody else can make a file in that
  directory. This is true on every OS and on every mount where Postgres runs. Steampipe needs
  no check of the socket's owner.
- Every `root` connection from Steampipe uses the socket. This includes the plugin manager, and
  `pg_dump` and `pg_restore` during a database upgrade. `pg_dump` uses the old installation's
  data directory, and `pg_restore` the new one's. Connections through the socket do not use SSL.
- A socket path has a length limit: 103 bytes on macOS, 107 on Linux. The path is
  `<data directory>/.s.PGSQL.<port>`. With the default install directory it is about 55 bytes.
  Before postgres starts, Steampipe checks the length. If the path is too long, Steampipe stops
  and says to set a shorter `STEAMPIPE_INSTALL_DIR`. Without this check, postgres fails with
  "could not create any Unix-domain sockets".
- A service that Steampipe 2.4 or earlier started has no socket in the data directory. If
  Steampipe needs `root` and finds the service running with no socket, it says: "The running
  service was started by an older version of Steampipe. Run `steampipe service stop --force`
  once."

Why the data directory, and not another folder:

| Folder | What it holds | Socket there? |
|---|---|---|
| `/tmp` | Files from every user | No. Another user can make a file with the socket's name first, so Steampipe would need an owner check. A service manager or container can give the server and the CLI different `/tmp` folders. |
| `config/` | Your connection files | No. It is your content, and is often read-only or shared from another machine. |
| `plugins/` | Plugin binaries | No. It can be shared between installs, and can be read-only. |
| `logs/` | Log files | No. Log tools read and rotate it. |
| `internal/` | Steampipe's runtime state | No. It is mode `0755`, and a mounted folder can belong to another user. Steampipe would need code to make it private, and that code fails on a mount point it does not own. |
| `db/<version>/data/` | The database | **Yes.** Postgres makes sure it is private, and keeps its other runtime files there. Any file system that can hold the database can hold the socket. |

A socket in the data directory works on a Docker Desktop bind mount and on a named volume. This
was tested with Postgres in a container.

Other ways to protect `root`, and why not:

- **A password for root.** It needs a secret on disk, a way to recover when that file is lost,
  and the password in every `root` connection. The socket needs none of these.
- **`peer` auth with `pg_ident.conf`.** Steampipe must write the OS user name into
  `pg_ident.conf`, and Steampipe and Postgres must find the same name. It is a second file to
  manage. The private socket gives the same result.
- **A second rules file for Steampipe, through `hba_file`.** Steampipe 2.4 would then keep
  reading an unchanged `pg_hba.conf` after a downgrade. But there would be two rules files.
  Instead, a downgrade costs one line. See [Upgrade and downgrade](#upgrade-and-downgrade).

### The steampipe user does not change

The `steampipe` user keeps `samehost trust`. Everything that connects from this machine's
addresses can still run queries with your cloud credentials and no password. This includes the
other OS users, tunnels and containers in the list above. This is a deliberate choice:
Powerpipe, Flowpipe and local tools connect with no setup. The docs must say this clearly.

### Every postgres start uses the same steps

The service start, the install, and the backup during a database upgrade all start postgres.
All three:

- examine and, if Steampipe owns it, write `pg_hba.conf` first,
- get the socket settings from one shared list, and
- connect as `root` through the socket.

`postgresql.conf` and `steampipe.conf` stay as today: the service start writes them.

### What goes away

- The root-only `pg_hba.conf` that the install writes first, and the second write after it.
- The database name in `pg_hba.conf`.
- The `hostssl` copies of the `host` lines.
- The random `root` password that the install sets. Nothing keeps it or uses it.
- "`pg_hba.conf` exists" as the test for "the database is installed". Use `PG_VERSION` in the
  data directory, which `initdb` writes.
- `root` over TCP, and the socket in `/tmp`.

## Upgrade and downgrade

| Situation | What happens |
|---|---|
| Upgrade, default `pg_hba.conf` (most installs) | Steampipe writes the new file at the first start. `root` over TCP stops. |
| Upgrade, your own `pg_hba.conf` with a `local` rule for `root` | Nothing changes. |
| Upgrade, your own `pg_hba.conf` with no such rule | The first start stops with the message. You choose, and start again. |
| Upgrade while a Steampipe 2.4 service runs | Queries work. `steampipe service stop` says to use `--force` once. |
| Downgrade, Steampipe's `pg_hba.conf` | Steampipe 2.4 reads `pg_hba.conf` and connects as `root` over TCP, so it cannot start or stop the service. See the steps below. |
| Downgrade, your own `pg_hba.conf` | Steampipe 2.4 reads your file as it is. It works if your file lets `root` connect over TCP. |

`pg_hba.conf` holds a copy of your `pg_hba.conf.d/` rules, so Steampipe 2.4 still uses them
after a downgrade.

Steps to go back to Steampipe 2.4:

1. If a service runs, run `steampipe service stop --force`.
2. Add `host all root samehost trust` at the end of `pg_hba.conf`. Steampipe 2.4 never writes
   that file again.
3. Run `steampipe service start`.

When you upgrade again, the checksum does not match, so Steampipe stops and shows the choices.
Delete `pg_hba.conf` to go back to Steampipe's rules.

## Cases

| Case | Steampipe 2.4 | After |
|---|---|---|
| Steampipe CLI and plugin manager as `root` | TCP | Socket in the data directory |
| `steampipe` user over TCP from this machine (Powerpipe, Flowpipe, psql, BI tools) | No password | No password |
| `steampipe` user from another machine | Password | Password |
| `root` from another OS user, a tunnel or a container | Allowed | Rejected |
| `root` by hand with `psql -h localhost -U root` | Allowed | Rejected. Use `psql -h <data directory> -p 9193 -U root postgres`, or add `host all root samehost trust` in `pg_hba.conf.d/`. |
| Your own rules | Edit `pg_hba.conf`, which is never updated | Add a file in `pg_hba.conf.d/`, or manage `pg_hba.conf` yourself |
| A hand-edited `pg_hba.conf` | Kept | Kept, never changed. Stops once if it has no `local` rule for `root`. |
| A platform that copies its own `pg_hba.conf` before each start | Works | Works if the file has a `local` rule for `root`. The platform should copy its rules into `pg_hba.conf.d/` instead. |
| An edit to the `pg_hba.conf` that Steampipe wrote | Kept, because the file is never written again | Kept: Steampipe stops and shows the choices |
| Database upgrade on an install that an older Steampipe ran | Works | Works. The backup examines and writes `pg_hba.conf` first. |
| Custom database name (`STEAMPIPE_INITDB_DATABASE_NAME`) | The name is in `pg_hba.conf` | `all`, limited by the grants |
| A tool that connects through the socket in `/tmp` | Works if `pg_hba.conf` has a `local` rule | Use the data directory instead |
| An install directory with a long path | Works | Stops if the socket path is too long |
| Two installs on the same port | Their `/tmp` sockets clash | No clash |

## Testing

Unit tests:

- `pg_hba.conf` is the header, then the rules from `pg_hba.conf.d/` in name order, then
  Steampipe's rules. Other files, and names that start with `.`, are skipped. With no folder,
  the file has only Steampipe's rules. The checksum covers every line after it.
- Each row of [Who owns pg_hba.conf](#who-owns-pg_hbaconf): no file, Steampipe's file, both old
  default files (with the default and with a custom database name, with other white space and
  comments), an edited Steampipe file, and a file that is yours. These are yours: an old file
  with a rule added at the top, one added at the bottom, one with a changed rule, one with a
  deleted rule, and one with the rules in another order.
- A file that is yours is not changed, byte for byte.
- The "your rules" lines and the "would not write now" lines in the message.
- Steampipe writes through a temporary file and a rename.
- All three postgres starts get the socket settings.
- `root` connections use the socket and no SSL. `steampipe` connections do not change.
- The socket path length check.

Acceptance tests, in `pg_hba.bats`, which does not stop every service on the machine:

- `root` over TCP is rejected. `root` through the socket works.
- Another OS user cannot connect as `root` through the socket. Skip this test when `sudo -n` does
  not work. It works in CI.
- The `steampipe` user connects over TCP from this machine with no password. It cannot connect
  to the `postgres` database.
- An install with the old default `pg_hba.conf` gets the new file at its next start.
- A rule in `pg_hba.conf.d/` wins over Steampipe's rule.
- A hand-edited old `pg_hba.conf` with no `local` rule for `root`: `steampipe service start`
  exits with code 32, prints the message with the user's rules, leaves the file unchanged byte for
  byte, and leaves no postgres running. A batch `steampipe query` fails the same way, with the
  message on stderr and nothing on stdout.
- The same file with `local all root trust` added: the service starts, and the file is unchanged.
- An edited Steampipe `pg_hba.conf`: the start stops before postgres starts, and the file is
  unchanged. `steampipe service restart` refuses, and the running service keeps running.
- A Steampipe file with its header lines deleted is treated as yours.
- When `pg_hba.conf` is yours, a file in `pg_hba.conf.d/` gives the warning and is not used.
- Downgrade: after a start with this version, Steampipe 2.4.9 cannot start the service. After
  the steps in [Upgrade and downgrade](#upgrade-and-downgrade), Steampipe 2.4.9 can start the
  service, run a query and stop it. This version then stops with the "edited" message, and after
  `pg_hba.conf` is deleted it starts.

Manual checks before the release, because CI runs on Linux only:

- macOS: another OS user cannot open the socket.
  `sudo -u nobody psql -h <data directory> -p 9193 -U root postgres` must fail.
- Docker Desktop: a container cannot connect as `root` through `host.docker.internal`. It can
  still connect as `steampipe`.

## Rollout

Done on develop. These two pull requests built an earlier version of this design, which kept
the socket in `/tmp` with an owner check, and copied a changed `pg_hba.conf` to
`pg_hba.conf.d/00-preserved.conf`:

1. #5108: Steampipe writes `pg_hba.conf` before every start, with `pg_hba.conf.d/` rules first.
2. #5109: `root` connects through the unix socket only.

Before the next release, two more pull requests change develop to this design:

3. #5118: **The socket goes in the data directory.** Remove the owner check and the `/tmp`
   socket. Add the path length check and the message for a service that an older Steampipe
   started.
4. #5119: **Steampipe never changes a pg_hba.conf it did not write.** Add the header checksum, the
   ownership rules, the stop and its message, the check before `restart` stops the service,
   and the write through a temporary file. Remove the copy to `00-preserved.conf`.

Release notes:

- Only the OS user who runs Steampipe can connect as the database superuser (`root`), and only
  through the unix socket in the data directory. To connect as `root` by hand, use
  `psql -h <data directory> -p 9193 -U root postgres`.
- Steampipe writes `pg_hba.conf` each time the database starts. Put your own rules in
  `pg_hba.conf.d/<name>.conf`. If you manage `pg_hba.conf` yourself, Steampipe never changes it,
  but it must have the rule `local all root trust`. Steampipe stops and explains when it does
  not.
- A tool that connects through the socket in `/tmp` must use the data directory instead.
- The steps to go back to an older version.

Docs:

- The `STEAMPIPE_DATABASE_PASSWORD` page says that `root` can log in from this machine with no
  password. Change it to show how to connect as `root` through the socket.
- Document `pg_hba.conf.d/`, managing `pg_hba.conf` yourself, and the message. Say that the
  `steampipe` user needs no password from this machine.

## Out of scope

- A password for the `steampipe` user from this machine. That is a product decision. It would
  stop the no-setup connection that Powerpipe and Flowpipe use.
- SSL-only connections.
- `pg_ident.conf`.
- A question at start to choose what to do with `pg_hba.conf`.
