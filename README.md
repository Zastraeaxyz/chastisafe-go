# chastisafe-go

A Go library for interacting with the [chastisafe.com](https://chastisafe.com)
API.

It is used by [locksmith](https://github.com/Zastraeaxyz/locksmith), a Discord
bot that lets users manage their ChastiSafe locks from Discord instead of the
Flutter PWA.

ChastiSafe has no published API. The endpoints and payloads here were recovered
from its Flutter web client and checked against a real account.

## Install

```sh
go get github.com/zastraeaxyz/chastisafe-go
```

## Quick start

```go
client, err := chastisafe.New(chastisafe.Options{Token: developerCode})
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

The token is your ChastiSafe developer code. Treat it like a password.

## Documentation

- [Usage](docs/usage.md): authentication, the client and service layers, stats
  and profiles
- [API](docs/api.md): endpoints the library wraps
- [Behaviour](docs/behaviour.md): what is verified, and the API's quirks
- [Testing](docs/testing.md): integration tests
