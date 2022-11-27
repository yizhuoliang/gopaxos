# Replica states
- `state`
- `slot_in`
- `slot_out`
- `requests`
- `proposals`
- `decisions`

# Conditions for states update
- states (except `requests`) update happen iff `replicaStateUpdateChannel` is non empty
- two ways to trigger populating `replicaStateUpdateChannel`:
  - any `CollectorRoutine` receiving a valid response (type 1)
  - handling `Request` gRPC call from client (`requests` is populated) and notifying all `MessengerRoutine` (type 2)
  - decision and proposal have same `slot_out` but different `CommandId` (type 2)

## type 1
- if received decisions is non-empty, `decisions` is 'refreshed'
- if `slot_out` is mapped to a decision `d` in `decisions`, `state` is updated and `slot_out` is incremented
  - if `slot_out` is mapped to a proposal `p` in `proposals`, `p` is deleted from `proposals`
    - if `d` and `p` have different `CommandId`, `p.Command` will be put back to `requests`

## type 2
- if no decision in `decisions` is mapped to by `slot_in`, a new proposal will be taken from `requests` and mapped to by `slot_in` in `proposals`
- after all `MessengerRoutine` have processed type 2 update (regardless of the result of the previous if statement), `slot_in` is incremented