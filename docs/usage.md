# Usage

A small, typed Go client for the ChastiSafe lock service at
`https://chastisafe.com`.

ChastiKey is gone. Its old API host `api.chastikey.com` no longer resolves.
ChastiSafe is the replacement, but it has no published API. The endpoints,
payloads and quirks in this package were recovered from the ChastiSafe Flutter
web client and checked against a real account.

This package is intentionally independent of Discord and SQLite. A bot feature
can depend on it without pulling in bot code.

## Authentication

Send the account's developer code in the `chastisafedev` header and the
`chastisafe` cookie:

```
chastisafedev: <developer code>
Cookie: chastisafe=<developer code>
```

The live service currently accepts the cookie form and can return a server
error for header-only requests, so the client sends both. Copy the code from the
signed-in web app.

Treat the code as a password:

- never log it
- never commit it
- store it only for the account that owns it
- validate it with `Who` before trusting it

`Who` returns ChastiSafe's linked Discord snowflake. Compare it with the
Discord user who supplied the code before storing anything:

```go
identity, err := client.Who(ctx)
if err != nil {
    return err
}
if !identity.OwnedBy(discordUserID) {
    return errors.New("this ChastiSafe code belongs to another Discord user")
}
```

## Basic use

```go
client, err := chastisafe.New(chastisafe.Options{
    Token:  code,
    Logger: log,
})
if err != nil {
    return err
}

locks, err := client.Loaded(ctx)
if err != nil {
    return err
}
for _, lock := range locks {
    fmt.Println(chastisafe.LockSummary(&lock))
}
```

`Client` is bound to one account. `WithToken` shares the HTTP client with a
different account, which is useful for a multi-member Discord bot.

## Primitive and service layers

`Client` is the primitive layer. It maps one method to one endpoint and keeps
the API's raw shapes.

`Service` is the convenience layer for a bot. It adds short-lived caches,
reference resolution, typed drafts, and higher-level operations:

```go
service := chastisafe.NewService(client)

template, err := service.ResolveTemplate(ctx, "Steel")
description := "Updated description"
updated, err := service.UpdateTemplate(ctx, template.ID.String(), chastisafe.TemplatePatch{
    Description: &description,
})

view, err := service.Lock(ctx, "Steel")
detail := chastisafe.LockDetailText(view)
```

`Service` resolves templates by ID, exact name, or unique prefix. Locks resolve
by ID, lock name, lockee name, keyholder name, or unique prefix. Ambiguous
references return `ErrRefAmbiguous` with the candidate labels.

Typed helpers:

- `TemplateDraft` builds the large create payload ChastiSafe expects.
- `TemplatePatch` sends only the fields a caller sets.
- `LockSettingsPatch` converts typed durations into the units `KeyholderSettings` uses.
- `LockFilter` and `LockView` select running locks and attach the caller's role.
- `ParseDuration`, `ParseCards`, and `ParsePurposes` handle Discord option text.
- `LocksText`, `TemplateDetailText`, `LockListText`, `LockDetailText`, and
  `ProfileText` render text without combinations or passcodes.

Caches last 20 seconds by default. Every write invalidates the relevant cache.
Call `Service.Invalidate` after an account change such as a new login.

## Stats leaderboard

`Leaderboard` reads the public `#stats` page. The criteria are `Chastity`,
`Task`, `Bondage`, `Loyalty` and `Total`:

```go
rows, err := client.Leaderboard(ctx, chastisafe.StatsChastity)
for _, row := range rows {
    fmt.Println(row.Name, row.ChastityLocks, row.TotalLocks, row.Visible)
}
```

`Visible` is the account's public-profile flag. The endpoint includes hidden
profiles, so filter on `Visible` when the bot should only show public accounts.

## Public profiles

`ProfileFor` reads `/profile/{user}`. It accepts a numeric ChastiSafe user ID
or a bare username, and strips the `#0` discriminator from usernames.

If the requested account has hidden its profile, ChastiSafe returns the
caller's own profile instead of an error. `ProfileFor` detects that and returns
`ErrProfileNotPublic`, so the bot does not accidentally present the caller's
stats as someone else's.

## Units

The API mixes seconds and minutes.

Seconds:

- `timeremaining`
- `timerMin`, `timerMax`
- `addtime`, `remainingtime`
- `wait` (next card pick)
- `timetonextcheckin`
- `timetohygieneunlock`
- `freezeDuration`
- `nextTaskPick`, `nextTaskComplete`
- `remainingguaranteed`
- `trusttime`
- `lockduration`

Minutes:

- `regularity` (card pick interval)
- `checkinperiod`, `checkintolerance`
- `hygieneunlockperiod`, `hygieneunlockduration`
- `autoresetperiod`
- `taskMinRegularity`

The client exposes `FormatDuration`, `FormatMinutes` and
`FormatShortDuration` so callers do not have to remember this.
