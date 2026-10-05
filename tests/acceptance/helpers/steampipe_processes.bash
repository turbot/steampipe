# Counts steampipe processes left behind by a test run. A process counts when:
#  - its command line names an install dir (via --install-dir or -D <dir>/db/..., or as
#    a plugin path <dir>/plugins/.../steampipe-plugin-*.plugin) that is
#      * the install under test ($STEAMPIPE_INSTALL_DIR, default $HOME/.steampipe), or
#      * anywhere under $TMPDIR, /tmp or /var/folders - so a concurrent run from another
#        checkout is counted too; or
#  - its executable (argv[0]) is the steampipe binary under test (`command -v steampipe`),
#    or the bare name "steampipe" (a shell passes the name as typed, so the path is lost).
#    This covers CLI processes and plugin managers of any install, including ones whose
#    install dir comes from the STEAMPIPE_INSTALL_DIR environment variable and so never
#    appears in the command line.
# Anything else with "steampipe" in its command line - editors, scripts, a service from a
# different binary under a directory outside those roots - is not counted. Install dirs
# containing spaces are not matched.
#
# macOS temp dirs are /var/folders/... but may appear as /private/var/folders/...,
# so a leading /private is stripped before comparing.
count_steampipe_processes() {
  local install_dir="${STEAMPIPE_INSTALL_DIR:-$HOME/.steampipe}"
  local bin_path bin_dir
  bin_path=$(command -v steampipe)
  if [ -n "$bin_path" ]; then
    bin_dir=$(cd "$(dirname "$bin_path")" 2>/dev/null && pwd -P)
    [ -n "$bin_dir" ] && bin_path="$bin_dir/$(basename "$bin_path")"
  fi
  ps -axo command | awk -v install_dir="${install_dir%/}" -v tmp_dir="${TMPDIR:-/tmp}" -v bin_path="$bin_path" '
    function norm(p) { sub(/^\/private/, "", p); sub(/\/+$/, "", p); return p }
    function in_scope(dir) {
      dir = norm(dir)
      if (dir == norm(install_dir) || index(dir, norm(install_dir) "/") == 1) return 1
      if (index(dir, norm(tmp_dir) "/") == 1) return 1
      if (index(dir, "/tmp/") == 1 || index(dir, "/var/folders/") == 1) return 1
      return 0
    }
    {
      hit = (bin_path != "" && (norm($1) == norm(bin_path) || $1 == "steampipe"))
      for (i = 1; i <= NF; i++) {
        t = $i
        if (t == "--install-dir" && i < NF && in_scope($(i + 1))) hit = 1
        else if (t ~ /^--install-dir=/ && in_scope(substr(t, 15))) hit = 1
        else if (t == "-D" && i < NF && $(i + 1) ~ /\/db\// && in_scope($(i + 1))) hit = 1
        else if (t ~ /\/steampipe-plugin-[^\/]*\.plugin$/ && t ~ /\/plugins\// && in_scope(t)) hit = 1
      }
      if (hit) n++
    }
    END { print n + 0 }'
}
