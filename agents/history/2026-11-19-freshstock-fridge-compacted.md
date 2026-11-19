# Freshstock fridge and freezer inventory

The owner expanded Froststock to fridge food and requested a new name.
The app is now Freshstock. Its Windows executable name, app ID, launch route,
repository, package paths, My Drive Froststock folder, SQLite state, shares,
private key and browser queues keep the established update/data contract.
The managed entry keeps the registered Froststock name because existing
independent updaters validate it exactly. README explains that update entry.

Items gain an optional location value, fridge or freezer. Missing locations
stay unassigned. Existing records are not rewritten or classified by guessing.
The UI adds All food/Fridge/Freezer plus a conditional Unassigned filter,
storage labels, scan storage selection, location editing and food history.
The last scan location is remembered per device, defaulting to fridge.
Backup export retains locations and import accepts legacy schema-1 records.
Shared members use the same inventory and can change storage normally.
The app and release branding use Freshstock and the bundled Lucide boxes icon.

Go handler regression checks cover missing legacy locations, fridge updates,
unsupported locations and stale edits. Browser acceptance and independent
preview.10 release promotion are pending.

Preview.10 was published as Freshstock from 1c6abf0 with ten verified immutable
assets. The public update channel promotion is a1b769c. The anonymous
planner offered preview.9 -> preview.10, selected only four changed app
files, retained the manager and unrelated apps, and staged the exact
accepted executable, SHA256
90c2564f646705c44dfc416476bd782c69180bf94ff1091ba5139d0b91fa25ad.

Go tests, vet, recognition checks and browser acceptance passed. The exact
release executable upgraded synthetic preview.9 inventory and existing
shared sessions, scanned a fridge photo pair, moved food through a shared
member account, retained location through used/history restoration, exported
locations, imported new and legacy backups, and kept data/shares across
restart. Desktop and phone screenshots were visually reviewed. Fresh native
handler tests cover empty app data and legacy records. Public checking/staging
is verified. This task did not repeat a full suite install or rollback.

C filled during compilation. Final builds used task-specific NTFS temp on F
and the existing I Go cache. All task-owned preview processes were stopped.
No owner installation was updated. The remaining Froststock picker/update
label and My Drive folder are intentional compatibility names.
