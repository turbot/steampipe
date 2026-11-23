## Kind: design

PR #5208 is a Dependabot branch. `acceptance / Build` fails: `TestLocalConnectionStringKeepsSpecialCharacters` (`pkg/db/db_local/create_connection_test.go:43`) expects `config.TLSConfig.Certificates` to hold the client cert, but pgx 5.13.0 loads file-backed client certs per request via `GetClientCertificate`, so the slice is empty.

Only a commit fixes it (change the assertion to check `GetClientCertificate`). The fleet does not commit to a Dependabot branch it does not own, so I posted the fix-mark comment for head e4d78dd3c17d54121d04b5c06a97e1995a7d0acb. The dispatcher takes the PR over on that mark and starts a fix round; a person can instead fix it, close it, or ask Dependabot (`@dependabot recreate`).

Recommendation: fix round updates the assertion, keeps the bump.
