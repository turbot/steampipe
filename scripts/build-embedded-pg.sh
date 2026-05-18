#!/usr/bin/env bash
#
# build-embedded-pg.sh - build the embedded PostgreSQL binaries from source,
# exactly per design/embedded_postgres_build_instructions.md, producing the
# relocatable .txz archive Steampipe expects.
#
# This is a thin, version-parameterised wrapper around the committed recipe.
# It does NOT invent build steps - it drives the documented macOS (§3.2-3.4)
# and Linux (§3.5-3.6) steps. To build for a different PostgreSQL major
# (e.g. the PG18 upgrade) only PG_VERSION changes.
#
# Usage:
#   scripts/build-embedded-pg.sh [PG_VERSION] [OUTPUT_DIR]
#
#   PG_VERSION   PostgreSQL source version, e.g. 14.19 (default) or 18.0
#   OUTPUT_DIR   where to place the built tree + .txz (default: ./pg-build)
#
# Output: $OUTPUT_DIR/<prefix>/{bin,lib,share} and the packed archive named
# per Steampipe's pkg/db/platform/paths_*.go TarFileName:
#   darwin arm64  -> postgres-darwin-arm_64.txz
#   darwin x86_64 -> postgres-darwin-x86_64.txz
#   linux  arm64  -> postgres-linux-arm_64.txz
#   linux  x86_64 -> postgres-linux-x86_64.txz
#
set -euo pipefail

PG_VERSION="${1:-14.19}"
OUTPUT_DIR="${2:-$(pwd)/pg-build}"

OS="$(uname -s)"
ARCH="$(uname -m)"

# Map to Steampipe's expected TarFileName (pkg/db/platform/paths_*.go).
case "$OS/$ARCH" in
  Darwin/arm64)  TARFILE="postgres-darwin-arm_64.txz" ;;
  Darwin/x86_64) TARFILE="postgres-darwin-x86_64.txz" ;;
  Linux/aarch64) TARFILE="postgres-linux-arm_64.txz" ;;
  Linux/arm64)   TARFILE="postgres-linux-arm_64.txz" ;;
  Linux/x86_64)  TARFILE="postgres-linux-x86_64.txz" ;;
  *) echo "ERROR: unsupported platform $OS/$ARCH" >&2; exit 1 ;;
esac

PREFIX="$OUTPUT_DIR/postgres-${PG_VERSION}-$(echo "$OS" | tr '[:upper:]' '[:lower:]')-${ARCH}"
SRC_TARBALL="postgresql-${PG_VERSION}.tar.bz2"
SRC_URL="https://ftp.postgresql.org/pub/source/v${PG_VERSION}/${SRC_TARBALL}"
SRC_DIR="$OUTPUT_DIR/postgresql-${PG_VERSION}"

echo "== build-embedded-pg =="
echo "  PG_VERSION : $PG_VERSION"
echo "  platform   : $OS/$ARCH  -> $TARFILE"
echo "  prefix     : $PREFIX"
echo "  output dir : $OUTPUT_DIR"

mkdir -p "$OUTPUT_DIR"
cd "$OUTPUT_DIR"

# --- fetch + extract source (postgresql.org/ftp/source) ---
if [[ ! -f "$SRC_TARBALL" ]]; then
  echo "== downloading $SRC_URL"
  curl -fL --retry 3 -o "$SRC_TARBALL" "$SRC_URL"
fi
rm -rf "$SRC_DIR"
tar -xf "$SRC_TARBALL"
cd "$SRC_DIR"

mkdir -p "$PREFIX"

if [[ "$OS" == "Darwin" ]]; then
  # ---- macOS: design doc §3.2 ----
  export MACOSX_DEPLOYMENT_TARGET=11.0
  export CFLAGS="-mmacosx-version-min=11.0"
  export LDFLAGS="-mmacosx-version-min=11.0 -Wl,-rpath,@loader_path/../lib/postgresql"
  OPENSSL_PREFIX="$(brew --prefix openssl)"

  # PostgreSQL 16+ builds with ICU by default and `configure` hard-fails
  # without icu-uc/icu-i18n (PG18 in particular). PG14/15 build fine
  # without it (and the proven PG14 path is left byte-identical). When
  # building >=16, require icu4c (keg-only on Homebrew) and put its
  # pkg-config on PKG_CONFIG_PATH, mirroring how Homebrew builds its own
  # postgresql@18 (--with-icu + icu4c pkgconfig).
  PG_MAJOR="${PG_VERSION%%.*}"
  ICU_CONFIGURE=()
  if [[ "$PG_MAJOR" -ge 16 ]]; then
    ICU_PREFIX="$(brew --prefix icu4c 2>/dev/null || true)"
    if [[ -z "$ICU_PREFIX" || ! -d "$ICU_PREFIX/lib/pkgconfig" ]]; then
      # fall back to the highest installed versioned icu4c keg
      ICU_PREFIX="$(ls -d /opt/homebrew/opt/icu4c@* 2>/dev/null | sort -V | tail -1 || true)"
    fi
    if [[ -z "$ICU_PREFIX" || ! -d "$ICU_PREFIX/lib/pkgconfig" ]]; then
      echo "ERROR: PostgreSQL $PG_VERSION needs ICU but icu4c was not found; run 'brew install icu4c'" >&2
      exit 1
    fi
    echo "== PG$PG_MAJOR: building --with-icu using $ICU_PREFIX"
    export PKG_CONFIG_PATH="$ICU_PREFIX/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
    ICU_CONFIGURE=(--with-icu)
  fi
  # PG<16: no ICU flag at all - identical to the validated PG14 recipe.

  ./configure --prefix="$PREFIX" \
    --libdir="$PREFIX/lib/postgresql" \
    --datadir="$PREFIX/share/postgresql" \
    --with-openssl \
    --with-includes="$OPENSSL_PREFIX/include" \
    --with-libraries="$OPENSSL_PREFIX/lib" \
    "${ICU_CONFIGURE[@]}"

  make -j"$(sysctl -n hw.ncpu)"
  make install
  make -C contrib
  make -C contrib install

  # §3.2 step 11: remove the include directory.
  rm -rf "$PREFIX/include"
  # §3.2 step 12: keep only the binaries the shipped artifact ships. The
  # committed doc said "remove unneeded binaries" without enumerating them;
  # the current shipped darwin-arm64 .txz was inspected and contains exactly
  # these five (see output/exec3-oracle-parity-14x.md). Pruning to match
  # keeps the from-source artifact structurally faithful to what Steampipe
  # ships and uses (initdb / postgres / pg_ctl + pg_dump / pg_restore for the
  # migration path).
  ( cd "$PREFIX/bin"
    for b in *; do
      case "$b" in
        initdb|pg_ctl|pg_dump|pg_restore|postgres) ;;
        *) rm -f "$b" ;;
      esac
    done
  )

  # ---- macOS rpath fix: design doc §3.3 (fix_rpath.sh, inlined) ----
  (
    cd "$PREFIX"
    LIB_SUBDIR="lib/postgresql"
    BUNDLE_ROOT="$(pwd)"
    LIBPQ_PATH="$BUNDLE_ROOT/$LIB_SUBDIR/libpq.5.dylib"
    echo "🔧 Fixing libpq install name..."
    install_name_tool -id "@rpath/libpq.5.dylib" "$LIBPQ_PATH"
    echo "🔍 Processing binaries in bin/..."
    for binfile in "$BUNDLE_ROOT"/bin/*; do
      [[ -x "$binfile" && ! -d "$binfile" ]] || continue
      # the shipped artifact carries BOTH rpaths (../lib and
      # ../lib/postgresql) - match it for faithful relocatability parity.
      install_name_tool -add_rpath "@loader_path/../lib" "$binfile" 2>/dev/null || true
      install_name_tool -add_rpath "@loader_path/../$LIB_SUBDIR" "$binfile" 2>/dev/null || true
      install_name_tool -change "$BUNDLE_ROOT/$LIB_SUBDIR/libpq.5.dylib" "@rpath/libpq.5.dylib" "$binfile" 2>/dev/null || true
    done
  )

  # ---- macOS pack: design doc §3.4 ----
  ( cd "$PREFIX" && tar --disable-copyfile --exclude='._*' -cJf "$OUTPUT_DIR/$TARFILE" bin lib share )

else
  # ---- Linux: design doc §3.6 ----
  export LDFLAGS='-Wl,-rpath,$ORIGIN/../lib/postgresql -Wl,--enable-new-dtags'
  ./configure \
    --prefix="$PREFIX" \
    --libdir="$PREFIX/lib/postgresql" \
    --datadir="$PREFIX/share/postgresql" \
    --with-openssl \
    --with-includes=/usr/include \
    --with-libraries="/usr/lib/${ARCH}-linux-gnu"

  make -j"$(nproc)"
  make install
  make -C contrib -j"$(nproc)"
  make -C contrib install

  ( cd "$PREFIX"
    for f in bin/*; do
      if [[ -x "$f" ]] && file "$f" | grep -q ELF; then
        patchelf --set-rpath '$ORIGIN/../lib/postgresql' "$f"
      fi
    done
  )
  rm -rf "$PREFIX/include"
  ( cd "$PREFIX" && tar -cJf "$OUTPUT_DIR/$TARFILE" bin lib share )
fi

echo "== done: $OUTPUT_DIR/$TARFILE"
ls -la "$OUTPUT_DIR/$TARFILE"
