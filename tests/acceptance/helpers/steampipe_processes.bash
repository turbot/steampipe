# Counts processes left behind by the install under test ($STEAMPIPE_INSTALL_DIR)
# and by installs the tests create under a temp directory (--install-dir).
#
# A process counts when its command line names an install dir that is one of those
# and it is one of: the plugin manager (--install-dir <dir>), the postgres server
# (-D <dir>/db/...), or a plugin (<dir>/plugins/.../steampipe-plugin-*.plugin).
# Anything else with "steampipe" in its command line - editors, scripts, another
# checkout's service, the developer's own ~/.steampipe - is not counted.
#
# macOS temp dirs are /var/folders/... but may appear as /private/var/folders/...,
# so a leading /private is stripped before comparing.
count_steampipe_processes() {
  local install_dir="${STEAMPIPE_INSTALL_DIR:-$HOME/.steampipe}"
  ps -axo command | awk -v install_dir="${install_dir%/}" -v tmp_dir="${TMPDIR:-/tmp}" '
    function norm(p) { sub(/^\/private/, "", p); sub(/\/+$/, "", p); return p }
    function in_scope(dir) {
      dir = norm(dir)
      if (dir == norm(install_dir) || index(dir, norm(install_dir) "/") == 1) return 1
      if (index(dir, norm(tmp_dir) "/") == 1) return 1
      if (index(dir, "/tmp/") == 1 || index(dir, "/var/folders/") == 1) return 1
      return 0
    }
    {
      hit = 0
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
