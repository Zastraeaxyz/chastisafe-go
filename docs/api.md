# Endpoints used

## Read

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/who` | Current account and linked Discord ID |
| GET | `/profile` | Profile, stats, developer code |
| GET | `/profile/{user}` | Another public profile (numeric ID or bare username) |
| GET | `/locks` | Managed lock templates |
| GET | `/loadable/{keyholder}` | Templates a keyholder offers |
| GET | `/getlock/{id}` | One template |
| GET | `/getlockforedit/{id}` | One template with edit fields |
| GET | `/loaded` | Active locks for this account |
| GET | `/loaded/{id}` | One active lock, plus `isLockee` |
| GET | `/lockees` | Active locks of this account's lockees |
| GET | `/lockee/{id}` | One lockee's active lock |
| GET | `/recent` | Completed lock history |
| GET | `/stats/{criterion}` | Public stats leaderboard |
| POST | `/log` | Lock event log |
| GET | `/tasks`, `/task/{id}` | Task rollers |
| GET | `/taskstatus/{lock}` | Task state for a lock |
| GET | `/pillory`, `/pillory/{lock}` | Pillory sessions |
| GET | `/scripts/{template}` | Saved scripts |
| GET | `/bots/card`, `/bots/timer` | Built-in keyholder bots |

## Lockee actions

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/load` | Start a lock from a template |
| GET | `/open/{id}` | Open a ready lock and reveal its code |
| GET | `/rate/{rating}/{id}` | Rate a completed lock from 1 to 5 |
| GET | `/abandon/{id}` | End a lock without a code |
| GET | `/reset/{id}` | Restart a lock |
| GET | `/random/{id}` | Random unlock attempt |
| GET | `/pick/{id}` | Draw a card |
| POST | `/update` | Add or remove timer time |
| GET | `/checkin/{id}` | Record a check-in |
| GET | `/keyholderdecide/{id}` | Ask the keyholder to decide |
| POST | `/hygiene` | Start a hygiene unlock with a code |
| GET | `/hygieneskip/{id}` | Skip a hygiene unlock |
| GET | `/putgreenback/{id}` | Return the last green card |
| POST | `/respond/{id}` | Answer a keyholder prompt |
| POST | `/trust` | Extend the lockee trust period |
| GET | `/taskpick/{lock}` | Draw a task |
| GET | `/taskcomplete/{lock}` | Complete a task |
| GET | `/taskabandon/{lock}` | Abandon a task |
| POST | `/vote` | Vote in a pillory |
| POST | `/pillorytaskdone` | Finish a pillory task |

## Keyholder actions

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/unlock/{id}` | Unlock a lockee's ready lock |
| POST | `/update` | Freeze, hide, purpose, remaining time, settings |
| POST | `/update` | Add or remove cards |
| POST | `/clear` | Clear warning flags |
| POST | `/split` | Create fake locks |
| POST | `/taskassign` | Assign a task roller item |
| POST | `/startpillory` | Start a pillory |
| GET | `/cancelpillory/{id}` | Cancel a pillory |
| POST | `/script` | Run a script |
| POST | `/button` | Press a script-backed button |

## Template management

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/create` | Create a template, or replace when `id` is set |
| POST | `/updatetemplate` | Partial update using `templateId` |
| GET | `/delete/{id}` | Delete a template |
| POST | `/setbackground` | Change a template background |
| POST | `/updatecardrange` | Change one card range on a template |
| POST | `/simulate` | Preview card duration and card counts |
