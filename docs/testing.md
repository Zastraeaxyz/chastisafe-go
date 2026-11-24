# Integration tests

Read-only, skipped unless `CHASTISAFE_DEV_CODE` is set:

```sh
CHASTISAFE_DEV_CODE=... go test -run TestIntegrationRead -v .
```

Destructive, skipped unless `CHASTISAFE_WRITE_TEST=1` is also set. These create
and delete a template, load a lock, and clean up:

```sh
CHASTISAFE_DEV_CODE=... CHASTISAFE_WRITE_TEST=1 \
  go test -run 'TestIntegration(Template|Timer)Lifecycle|TestIntegrationLockeeCardUpdate' -v .
```

Two-account keyholder test:

```sh
CHASTISAFE_WRITE_TEST=1 \
  CHASTISAFE_KEYHOLDER_DEV_CODE=... \
  CHASTISAFE_LOCKEE_DEV_CODE=... \
  go test -run TestIntegrationKeyholderActions -v .
```

Do not run the write tests against an account that matters.
