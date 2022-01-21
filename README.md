# Week Work Planner

This project implements a simple week work planner API with these features:

- There are two kinds of users: workers and admins.
- Workers can pick available shifts. Admins can't pick shifts.
- Picking already picked shifts is not allowed.
- A worker can't have two shifts on the same day.
- Maximum shifts per worker per week is 6.
- Workers can delete their own shifts.
- Admins can delete any worker shifts.
- Workers can see their shifts.
- Admins can see all shifts, per user or in total.
- There are 3 shifts per day.

## Tokens

API users are defined by a token that has to be sent in every request. They must contain the header:

```
{
  "alg": "HS256",
  "typ": "JWT"
}
```
And the payload:

```
{
  "lvl": "0",
  "nam": "Xavier Marco",
  "uid": "2",
  "usr": "xavier"
}
```
The key `lvl` tells the API if a user is a worker (level 0), or an admin (level 1). For the purpose of this API, tokens does not expire. The last part of the token is the signature. The API gets the signing key from an environment variable name `SIGNKEY17`, and has to be encoded in base64. Obviously the tokens must be signed with the same key, otherwise the API will not accept the token and the request will result in a forbidden response.

## Database

There is no database, week planner shifts will be saved in memory. If the server stops, data is lost. User data like level, name, or userid, is stored in the tokens payload's claims.

## Endpoints

There is one endpoint with three methods. Assuming we are running in localhost, in port 80, we will have:

### For Workers

- `GET http://localhost/`: return user's shifts.
- `POST http://localhost/{day}{shift}`: request the shift.
- `DELETE http://localhost/{day}{shift}`: delete the shift.

### For Admins

- `GET http://localhost/`: return all working plan and user data.
- `GET http://localhost/{userid}`: return shifts from specific user.
- `DELETE http://localhost/{userid}{day}{shift}`: delete the shift.

## Format

Errors are returned in plain text. All other data is returned as a JSON.