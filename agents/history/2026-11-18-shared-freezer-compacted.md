# Shared freezer access

The owner requested sharing Froststock with other OneAccounts users. The
default keeps an existing freezer private. Settings lets the owner add and
remove verified email invitations on the same configured issuer. The first
accepted invitation binds to the stable app account ID. Later email
reassignment does not inherit access. Every inventory, photo and recognition
request rechecks sharing. Owners manage shares and the saved DeepSeek key.
Shared users can read and change the common My Drive inventory and use its
configured label reader.

The existing item format, revision conflict handling, per-account browser
upload queues and five owned package paths continue. Sharing is migration 1
in the existing AppHost SQLite database, included in normal app backups.
Froststock preview.9 requires the existing preview.140 updater contract.
Focused handler tests cover member uploads, owner/member inventory equality,
photos, conflicting edits, foreign/unverified identities, CSRF, duplicate
invites, binding, email changes, revocation and owner-only settings. Go tests,
vet and recognition checks passed. Browser acceptance and independent release
publication remain to be verified before delivery.

Preview.9 was published from b85731f with ten immutable assets. Public channel
promotion 9a1853d and the normal anonymous planner select preview.9 from the
Froststock repository. Verified staged executable SHA256 is
9ca3a467cfa034ac2fd1033b7c992e20692f648b35d063a7628bc52c7447dd0c.
The planner selected only the three changed Froststock files and preserved
the manager and all unrelated app entries.

The release executable passed browser acceptance with real signed OIDC
callbacks from a synthetic issuer, separate owner/member/uninvited browser
sessions, existing preview.8 data, shared uploads and UI edits, persistence
across restart, and immediate revocation on the next request. Desktop and
phone settings were visually checked. A fresh isolated Windows Workspace140
installation selected Accounts and Drive, then installed preview.9 through
the public updater, checked all app health states, rolled back and reapplied.
The manager binary, owner and file root remained unchanged. Test processes
were stopped. No owner installation was changed.
