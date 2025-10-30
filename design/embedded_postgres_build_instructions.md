# PostgreSQL Source Build Instructions

This document provides step-by-step instructions for building the embedded PostgreSQL binaries required by Steampipe for a specific PostgreSQL version. It covers both macOS and Linux environments, including prerequisites, build steps, and packaging guidelines to ensure the resulting binaries are relocatable and suitable for Steampipe's use.

1. **Source Code:**
   [https://www.postgresql.org/ftp/source/](https://www.postgresql.org/ftp/source/)

2. **Build Documentation:**
   [https://www.postgresql.org/docs/current/install-make.html](https://www.postgresql.org/docs/current/install-make.html)

---

## 3. Download Source Code and Run

### For MacOS

#### 3.1. Pre-requisites

* `openssl`

---

#### 3.2. Steps to Build

1. Change to the PostgreSQL source directory:

   ```bash
   cd /postgres/source/dir
   ```

2. Set environment variables:

   ```bash
   export MACOSX_DEPLOYMENT_TARGET=11.0
   export CFLAGS="-mmacosx-version-min=11.0"
   export LDFLAGS="-mmacosx-version-min=11.0 -Wl,-rpath,@loader_path/../lib/postgresql"
   ```

   *(Rebuild with an older deployment target)*

3. Configure the build:

   ```bash
   ./configure --prefix=location/where/you/want/the/files \
   --libdir=/location/where/you/want/the/files/lib/postgresql \
   --datadir=/location/where/you/want/the/files/share/postgresql \
   --with-openssl \
   --with-includes=$(brew --prefix openssl)/include \
   --with-libraries=$(brew --prefix openssl)/lib
   ```

   *(Make sure the `libdir` and `datadir` args are passed correctly and point to the `postgresql` dir inside `lib` and `share` — this is needed for Steampipe.)*

4. Build PostgreSQL:

   ```bash
   make -j$(sysctl -n hw.ncpu)
   ```

5. Install binaries:

   ```bash
   make install
   ```

6. Verify that all binaries are built in the specified location.

7. Build contrib modules:

   ```bash
   make -C contrib
   ```

8. Install contrib modules:

   ```bash
   make -C contrib install
   ```

9. *(This builds extensions in the contrib directory — needed since we load `ltree` and `tablefunc`.)*

10. Verify installation structure:

    ```bash
    ls -al location/where/you/want/the/files
    ```

    You should see `lib`, `share`, `bin`, and `include` directories under that path.

11. Remove the `include` directory.

12. Remove unneeded binaries from `bin`.

13. Check that all extensions exist.

---

#### 3.3. Fix RPATHs

Run the `fix_rpath.sh` script to fix the rpaths of the binaries (`initdb`, `pg_restore`, `pg_dump`):

```bash
#!/bin/bash
set -euo pipefail

# --- CONFIGURE ---
# Adjust if your libpq lives in lib/ not lib/postgresql
LIB_SUBDIR="lib/postgresql"
BUNDLE_ROOT="$(pwd)"
LIBPQ_PATH="$BUNDLE_ROOT/$LIB_SUBDIR/libpq.5.dylib"

echo "🔧 Fixing libpq install name..."
install_name_tool -id "@rpath/libpq.5.dylib" "$LIBPQ_PATH"

echo "🔍 Processing binaries in bin/..."
for binfile in "$BUNDLE_ROOT"/bin/*; do
  [[ -x "$binfile" && ! -d "$binfile" ]] || continue
  echo "➡️  Patching $(basename "$binfile")"

  # Ensure an rpath to ../lib/postgresql exists
  install_name_tool -add_rpath "@loader_path/../$LIB_SUBDIR" "$binfile" 2>/dev/null || true

  # Rewrite any absolute reference to libpq
  install_name_tool -change     "$BUNDLE_ROOT/$LIB_SUBDIR/libpq.5.dylib"     "@rpath/libpq.5.dylib"     "$binfile" 2>/dev/null || true
done

echo "✅ Verification:"
for binfile in "$BUNDLE_ROOT"/bin/*; do
  [[ -x "$binfile" && ! -d "$binfile" ]] || continue
  echo "--- $(basename "$binfile") ---"
  otool -L "$binfile" | grep libpq || echo "⚠️  No libpq linkage"
  otool -l "$binfile" | grep -A2 LC_RPATH | grep path || echo "⚠️  No RPATH"
done
```

---

#### 3.4. Pack the Built Binaries

Create a `.txz` archive:

```bash
tar --disable-copyfile --exclude='._*' -cJf darwin-arm64.txz -C darwin-arm64 bin lib share
```

---

### For Linux (Ubuntu 24 / amd64 or arm64)

#### 3.5. Pre-requisites

```bash
apt update
apt install -y build-essential wget ca-certificates \
               libreadline-dev zlib1g-dev flex bison \
               libssl-dev patchelf file
```

---

#### 3.6. Steps to Build

1. Change to the PostgreSQL source directory:

   ```bash
   cd /postgres/source/dir
   ```

2. Set installation prefix and linker flags:

   ```bash
   export PREFIX=/postgres-binaries-14.19/linux-$(uname -m)
   mkdir -p "$PREFIX"
   export LDFLAGS='-Wl,-rpath,$ORIGIN/../lib/postgresql -Wl,--enable-new-dtags'
   ```

3. Configure:

   ```bash
   ./configure \
     --prefix="$PREFIX" \
     --libdir="$PREFIX/lib/postgresql" \
     --datadir="$PREFIX/share/postgresql" \
     --with-openssl \
     --with-includes=/usr/include \
     --with-libraries=/usr/lib/$(uname -m)-linux-gnu
   ```

4. Build and install:

   ```bash
   make -j2
   make install
   ```

5. Build contrib extensions:

   ```bash
   make -C contrib -j2
   make -C contrib install
   ```

6. Patch RPATHs for relocatability:

   ```bash
   cd "$PREFIX"
   for f in bin/*; do
     if [ -x "$f" ] && file "$f" | grep -q ELF; then
       patchelf --set-rpath '$ORIGIN/../lib/postgresql' "$f"
     fi
   done
   ```

7. Verify RPATH:

   ```bash
   readelf -d bin/initdb | grep -i rpath
   # → RUNPATH [$ORIGIN/../lib/postgresql]
   ```

8. Verify linkage:

   ```bash
   ldd bin/initdb | grep libpq
   # → libpq.so.5 => .../bin/../lib/postgresql/libpq.so.5
   ```

9. Remove `include` directory and unnecessary binaries:

   ```bash
   rm -rf "$PREFIX/include"
   ```

10. Pack into `.txz`:

```bash
cd $(dirname "$PREFIX")
tar -cJf postgres-14.19-$(uname -m).txz $(basename "$PREFIX")
```

✅ **Done.**
