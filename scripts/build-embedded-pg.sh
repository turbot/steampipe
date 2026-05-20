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

# Clean the install prefix between runs. Without this, stale files from
# a prior build (e.g. an ICU bundle from a different ICU major) survive
# into the new artifact because `make install` and the wrapper's
# bundling step only write, never delete. Mirrors how SRC_DIR is
# cleaned above. Idempotent for first runs (rm -rf on a missing dir is
# a no-op).
rm -rf "$PREFIX"
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
  # the current shipped darwin-arm64 .txz contains exactly these five.
  # Pruning to match keeps the from-source artifact structurally faithful
  # to what Steampipe ships and uses (initdb / postgres / pg_ctl for the
  # service; pg_dump / pg_restore for the migration path).
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

  # ---- macOS ICU bundling (PG16+; Option A: full unfiltered ICU) ----
  # PG18 is built --with-icu, so postgres links libicui18n/libicuuc/
  # libicudata at the build machine's absolute icu4c path - not
  # relocatable. Bundle the three ICU dylibs into lib/postgresql and
  # rewrite references (same pattern as libpq, extended to the 3-lib
  # chain): binary->icu = @rpath (bins already have the ../lib/postgresql
  # rpath); icu->icu = @loader_path (co-located in the same dir); each
  # bundled lib id = @rpath. The ICU major is pinned to whatever
  # $ICU_PREFIX provides at build time and is a collation-stability
  # contract: changing it can alter text sort order, so treat an ICU
  # major bump as a deliberate, separately-tested change.
  if [[ "$PG_MAJOR" -ge 16 ]]; then
    (
      cd "$PREFIX"
      LIB_SUBDIR="lib/postgresql"
      BUNDLE_ROOT="$(pwd)"
      echo "📦 Bundling ICU from $ICU_PREFIX"
      for stem in libicudata libicuuc libicui18n; do
        # `|| true`: under `set -o pipefail` a no-match from grep makes the
        # whole substitution exit non-zero; on a bare assignment `set -e`
        # would then abort here BEFORE the fallback/error below could run.
        src="$(ls "$ICU_PREFIX"/lib/${stem}.*.dylib 2>/dev/null | grep -E "/${stem}\.[0-9]+\.dylib$" | head -1)" || true
        [[ -z "$src" ]] && src="$(ls "$ICU_PREFIX"/lib/${stem}.*.dylib 2>/dev/null | head -1)" || true
        [[ -z "$src" ]] && { echo "ERROR: ICU lib $stem not found in $ICU_PREFIX/lib" >&2; exit 1; }
        cp -L "$src" "$BUNDLE_ROOT/$LIB_SUBDIR/$(basename "$src")"
        install_name_tool -id "@rpath/$(basename "$src")" "$BUNDLE_ROOT/$LIB_SUBDIR/$(basename "$src")"
      done
      # rewrite every reference into the build-machine ICU prefix:
      # within a bundled ICU lib -> @loader_path (siblings); elsewhere
      # (the bin/* binaries) -> @rpath.
      relocate_icu() {
        local f="$1" anchor="$2"
        otool -L "$f" | awk 'NR>1{print $1}' | while read -r dep; do
          case "$dep" in
            "$ICU_PREFIX"/*|*/icu4c*/lib/libicu*)
              install_name_tool -change "$dep" "$anchor/$(basename "$dep")" "$f" 2>/dev/null || true ;;
          esac
        done
      }
      for stem in libicudata libicuuc libicui18n; do
        for l in "$BUNDLE_ROOT/$LIB_SUBDIR/${stem}".*.dylib; do
          [[ -f "$l" ]] && relocate_icu "$l" "@loader_path"
        done
      done
      for binfile in "$BUNDLE_ROOT"/bin/*; do
        [[ -x "$binfile" && ! -d "$binfile" ]] || continue
        relocate_icu "$binfile" "@rpath"
      done
      echo "✅ ICU bundled: $(ls "$BUNDLE_ROOT/$LIB_SUBDIR"/libicu*.dylib | xargs -n1 basename | tr '\n' ' ')"
    )
  fi

  # ---- macOS re-sign (MANDATORY on Apple Silicon) ----
  # install_name_tool invalidates the Mach-O code signature; arm64 macOS
  # then SIGKILLs the binary/dylib on load ("Killed: 9", no output - NOT
  # a dyld 'library not loaded' error). Ad-hoc re-sign every Mach-O the
  # rpath/ICU fixups touched. The committed recipe's fix_rpath.sh omits
  # this; ICU bundling (many more install_name_tool ops) makes it fatal.
  (
    cd "$PREFIX"
    for f in bin/* lib/postgresql/*.dylib; do
      [[ -f "$f" ]] || continue
      file "$f" | grep -q 'Mach-O' || continue
      codesign --force --sign - "$f" 2>/dev/null || true
    done
  )

  # ---- verify contrib extensions exist (recipe §3.2 step 13) ----
  # a broken `make -C contrib install` would otherwise ship an artifact
  # silently missing ltree/tablefunc, which Steampipe loads. The loadable
  # module suffix follows PostgreSQL's DLSUFFIX: .so on Linux and PG<=17
  # macOS, but .dylib on macOS PG18 (src/Makefile.global). Accept either -
  # this check is for a *missing* module, not the suffix.
  # NOTE: the macOS shipped-convention mismatch (Steampipe and the current
  # artifact use .so; from-source PG18-macOS emits .dylib - the same
  # family as the FDW `inst` .so/.dylib quirk) is a packaging decision
  # for the release step, deliberately NOT forced here. The shipped
  # builds are Linux, where DLSUFFIX is .so and this is moot.
  for ext in ltree tablefunc; do
    { [[ -f "$PREFIX/lib/postgresql/${ext}.so" || -f "$PREFIX/lib/postgresql/${ext}.dylib" ]] \
      && [[ -f "$PREFIX/share/postgresql/extension/${ext}.control" ]]; } \
      || { echo "ERROR: required extension '$ext' missing after contrib build (no lib/postgresql/${ext}.{so,dylib} or share/postgresql/extension/${ext}.control)" >&2; exit 1; }
  done

  # ---- macOS pack: design doc §3.4 ----
  ( cd "$PREFIX" && tar --disable-copyfile --exclude='._*' -cJf "$OUTPUT_DIR/$TARFILE" bin lib share )

else
  # ---- Linux: design doc §3.6 ----
  export LDFLAGS='-Wl,-rpath,$ORIGIN/../lib/postgresql -Wl,--enable-new-dtags'

  # PG16+ requires --with-icu at configure time (PG18 hard-fails
  # without it). Mirror of the macOS ICU-detection block above; on
  # Linux/Ubuntu libicu-dev provides icu-uc.pc / icu-i18n.pc, so detect
  # via pkg-config (vs Homebrew prefix on macOS).
  PG_MAJOR_LC="${PG_VERSION%%.*}"
  ICU_CONFIGURE=()
  if [[ "$PG_MAJOR_LC" -ge 16 ]]; then
    if ! pkg-config --exists icu-uc icu-i18n 2>/dev/null; then
      echo "ERROR: PostgreSQL $PG_VERSION needs ICU but icu-uc/icu-i18n pkg-config not found; install libicu-dev" >&2
      exit 1
    fi
    echo "== PG$PG_MAJOR_LC: building --with-icu (linux) using $(pkg-config --variable=libdir icu-uc)"
    ICU_CONFIGURE=(--with-icu)
  fi

  ./configure \
    --prefix="$PREFIX" \
    --libdir="$PREFIX/lib/postgresql" \
    --datadir="$PREFIX/share/postgresql" \
    --with-openssl \
    --with-includes=/usr/include \
    --with-libraries="/usr/lib/${ARCH}-linux-gnu" \
    "${ICU_CONFIGURE[@]}"

  make -j"$(nproc)"
  make install
  make -C contrib -j"$(nproc)"
  make -C contrib install

  # Keep only the 5 binaries Steampipe ships — mirror of the macOS
  # prune (this file, "step 12: keep only the binaries the shipped
  # artifact ships" block). The Linux branch previously skipped this,
  # so the .txz carried ~30 unused binaries (clusterdb, pgbench,
  # pg_upgrade, psql, ...).
  ( cd "$PREFIX/bin"
    for b in *; do
      case "$b" in
        initdb|pg_ctl|pg_dump|pg_restore|postgres) ;;
        *) rm -f "$b" ;;
      esac
    done
  )

  ( cd "$PREFIX"
    for f in bin/*; do
      if [[ -x "$f" ]] && file "$f" | grep -q ELF; then
        patchelf --set-rpath '$ORIGIN/../lib/postgresql' "$f"
      fi
    done
  )
  # ---- Linux ICU bundling (PG16+; Option A) ----
  # ELF resolves by soname via rpath, so bundling = drop the 3 ICU
  # .so.<major> into lib/postgresql (bins already have
  # $ORIGIN/../lib/postgresql rpath) and give the ICU libs an $ORIGIN
  # rpath so icu->icu resolves within the bundle. NOTE: this Linux path
  # is implemented but has only been logic-reviewed, not run (built/
  # proven on macOS only); verify it on a Linux build host/CI runner.
  PG_MAJOR_L="${PG_VERSION%%.*}"
  if [[ "$PG_MAJOR_L" -ge 16 ]]; then
    ICU_LIBDIR="$(pkg-config --variable=libdir icu-uc 2>/dev/null || echo /usr/lib/${ARCH}-linux-gnu)"
    for stem in libicudata libicuuc libicui18n; do
      # `|| true`: pipefail + bare assignment would abort under set -e on
      # a grep no-match before the fallback/error below could run.
      src="$(ls "$ICU_LIBDIR"/${stem}.so.* 2>/dev/null | grep -E "/${stem}\.so\.[0-9]+$" | head -1)" || true
      [[ -z "$src" ]] && src="$(ls "$ICU_LIBDIR"/${stem}.so.* 2>/dev/null | head -1)" || true
      [[ -z "$src" ]] && { echo "ERROR: ICU lib $stem not found in $ICU_LIBDIR" >&2; exit 1; }
      cp -L "$src" "$PREFIX/lib/postgresql/$(basename "$src")"
      patchelf --set-rpath '$ORIGIN' "$PREFIX/lib/postgresql/$(basename "$src")" 2>/dev/null || true
    done
    echo "✅ ICU bundled (linux): $(ls "$PREFIX"/lib/postgresql/libicu*.so.* | xargs -n1 basename | tr '\n' ' ')"
  fi

  rm -rf "$PREFIX/include"

  # ---- verify contrib extensions exist (recipe §3.2 step 13) ----
  # Linux DLSUFFIX is .so; accept .dylib too for symmetry with the macOS
  # path. This catches a *missing* contrib module, not the suffix.
  for ext in ltree tablefunc; do
    { [[ -f "$PREFIX/lib/postgresql/${ext}.so" || -f "$PREFIX/lib/postgresql/${ext}.dylib" ]] \
      && [[ -f "$PREFIX/share/postgresql/extension/${ext}.control" ]]; } \
      || { echo "ERROR: required extension '$ext' missing after contrib build (no lib/postgresql/${ext}.{so,dylib} or share/postgresql/extension/${ext}.control)" >&2; exit 1; }
  done

  # --hard-dereference: follow symlinks at archive time so the .txz
  # carries plain files (matches zonky's Linux pack pattern; avoids
  # downstream surprises when an upstream lib is a symlink).
  ( cd "$PREFIX" && tar --hard-dereference -cJf "$OUTPUT_DIR/$TARFILE" bin lib share )
fi

echo "== done: $OUTPUT_DIR/$TARFILE"
ls -la "$OUTPUT_DIR/$TARFILE"
