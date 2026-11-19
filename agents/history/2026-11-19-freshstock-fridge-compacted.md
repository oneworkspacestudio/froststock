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
