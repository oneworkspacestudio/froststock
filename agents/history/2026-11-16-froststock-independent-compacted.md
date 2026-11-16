# Froststock app repository

The Workspace owner requested independent Froststock releases. The first
standalone package continues the Froststock preview.7 interface, recognition
and data formats from Workspace preview.139. The source moved from the suite
into this repository. Workspace preview.140 discovers this feed and omits
Froststock from its own installer payload.

The service remains a managed Froststock app with MyAccounts, My Drive and
the existing OIDC, owner-only API, storage path and DeepSeek private-settings
contract. The Go module pins the public Workspace AppHost source commit.
Launch metadata and the original five Froststock package paths are in
workspace-app.json. Keep the same app ID and paths across future updates.

Build and publish with the app-local scripts. Verify the Windows package in
isolated Workspace fixtures and promote the shared update index afterward.
Do not publish another Workspace release for ordinary Froststock updates.
