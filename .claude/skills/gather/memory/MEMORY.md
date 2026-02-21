# Browser Automation Findings

## playwright-cli Setup

- Package: `@playwright/cli` (installed globally via `npm install -g @playwright/cli@latest`)
- Skills installed to `.claude/skills/playwright-cli` via `playwright-cli install --skills`
- Default browser: Chrome (detected automatically)

## Browser Launch Modes

### `--headed` (no persistence)
- Works reliably, fresh session each time
- User must re-login every session
- Downloads go to `.playwright-cli/<filename>` in the working directory
- **Best for: one-off sessions**

### `--headed --profile=<path>` (persistent)
- Stores cookies/session in the specified directory
- **Major issue:** Chrome profile locks. If the browser crashes or playwright-cli daemon dies, the profile gets locked and Chrome shows "Something went wrong when opening your profile."
- Profile lock files: `SingletonLock`, `SingletonSocket`, `SingletonCookie` in the profile dir
- Removing lock files can corrupt the profile
- Can't reuse profile if any Chrome instance is already running with it
- **Verdict: fragile, avoid for now**

### `--extension` (attach to existing Chrome)
- Requires "Playwright MCP Bridge" Chrome extension
- Connects to user's actual Chrome session (with all existing auth)
- **Issue:** If the active tab is a `chrome-extension://` URL, connection fails with "Cannot access a chrome-extension:// URL of different extension"
- Must have a regular webpage as the active tab before connecting
- **Issue:** Initial connection sometimes times out — retry usually works
- **Verdict: promising but finicky, needs the active-tab workaround**

## Download Behavior

- Downloads go to `.playwright-cli/` in the working directory (NOT `~/Downloads/`)
- Filename is whatever the server sends (e.g., `DownloadTxnHistory.csv`)
- If downloading multiple files with the same name, later downloads may overwrite — move files immediately after each download

## Reliability Issues

- Browser session can die between commands (especially after downloads)
- Need to check if browser is still open before each command block
- `playwright-cli kill-all` kills daemon processes but not necessarily the Chrome window
- Chrome windows from `--profile` mode persist after daemon death

## Element Selection

- Named selectors (e.g., `combobox "Duration"`) often fail with "not found in current page snapshot" even when the element exists in the YAML
- Using element refs (e.g., `e321`) from the snapshot YAML is much more reliable
- Pattern: always `snapshot` → read YAML → click by ref

## Recommendations for gather skill

1. Use `--headed` mode (no profile) as the default — accept the re-login cost
2. Consider `--extension` as an advanced option for users who install the bridge extension
3. Always check browser status before navigation commands
4. Move downloaded files immediately after each download
5. Don't rely on the browser surviving across long pauses
6. Always use element refs from snapshot YAML, not named selectors
