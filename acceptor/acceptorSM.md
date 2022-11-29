# Acceptor as a State Machines
## Acceptor States
- Ballot number
- Accepted proposals (A map from ballot number to a list of proposals in out implementation)

## Accetptor State Transitions
**T1 - Ballot Number Update**
1. Update the ballot number when a larger ballot number appears in a Scout or a Commander message

**P1 - Ballot Number Update Predicates**
1. received a Scout/Commander with a higher ballot

**T2 - Accepted Proposals Update**
1. Add a proposal from a commander to ```Accepted Proposals```

**P2 - Accepted Update Predicates**
1. received a Commander with a higher ballot