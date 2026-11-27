# Verified behaviour and quirks

## Verified behaviour

Checked against a live account:

- header and cookie authentication with `chastisafedev` and `chastisafe`
- `/who`, `/profile`, `/locks`, `/loaded`, `/recent`
- `/stats/{criterion}` leaderboard
- `/profile/{id}` public profiles, including hidden-profile detection
- `POST /load` returns `{"locks":[...]}`
- `GET /open/{id}` opens a ready lock and returns its code
- `GET /abandon/{id}` ends a lock
- `GET /pick/{id}` returns the card and the updated lock
- `POST /update` with `addtime` adds timer seconds for a lockee when the
  template allows updates
- `POST /update` with `cards` works for a lockee when the template allows
  updates and the card is a yellow variant
- a lockee is denied when trying to modify a green card
- template create, partial update, card range edit, background, simulate and
  delete
- with a second account: `/lockees`, keyholder freeze, hide, remaining time,
  purpose change, card edits, split and unlock
- lockee trust extension
- a self-lock returns `locknotfound` for `frozenbykh`, `hidden`,
  `remainingtime` and similar keyholder fields

Not yet checked against a real keyholder account:

- keyholder task, pillory and script actions

The request shapes for those actions come from the web client. A self-lock
cannot exercise them because its keyholder is `<self>` and ChastiSafe denies
keyholder writes with `locknotfound`.

## Quirks

- `keyholderId` is a JSON string in loaded locks and a JSON number in
  templates. `ID` accepts both.
- Loaded timestamps are RFC3339 strings. Log and history timestamps are epoch
  seconds. `Time` accepts both.
- `status` is sometimes `0`/`1` and sometimes `false`/`true`. `Status` accepts
  both.
- Some responses put the updated lock at the top level. Others return
  `{"status":0}`. The client handles both.
- `POST /update` succeeds for a self-lock with `addtime` but not with
  keyholder fields.
- Loaded-lock `frozenbykh`, `hidden` and `cumulative` require JSON booleans.
  Template `hidden`, `startfrozen` and `cumulative` are the strings `yes`,
  `no` or `randomize`.
- `frozenByCard` can stay `true` after a card freeze ends, until the next pick,
  and `freezeDuration` is not a reliable active-freeze signal. The app draws
  whenever a pick is available, so `IsFrozen` treats a card freeze as active
  only while `DrawsAvailable()` is zero. A keyholder freeze always counts.
- `POST /simulate` returns `hiddenlocksimulationnotshowntolockee` when the
  keyholder has hidden the simulation from the lockee. `errors.Is(err,
  ErrSimulationDisabled)` detects it.
- `POST /trust` is a lockee action. It works with a duration in seconds.
  Calling it as the keyholder returns an internal error.
- `SetPurpose` works for an allowed purpose. In testing, changing a lock whose
  purpose was already `TEST` back to `CHASTITY` returned `locknotfound`.
- `GET /open/{id}` is the lockee path. `GET /unlock/{id}` is the keyholder
  path.
- `deck` is a map of card name to remaining count. The app also accepts the
  lowercase alias `yellow`, and a loaded deck can report a specific yellow
  variant as `YELLOW`. Use `Deck.Yellows()` when the variant does not matter.
- `requiredgreens` values above 1000 mean "all greens". Values of 1000 or less
  mean a fixed number of greens.
- `GET /profile/{user}` accepts a numeric user ID or a bare username. Strip the
  `#0` discriminator before building the path. A hidden profile returns the
  caller's profile with a different ID, which `ProfileFor` rejects.
