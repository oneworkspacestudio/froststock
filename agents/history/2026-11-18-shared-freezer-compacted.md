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
