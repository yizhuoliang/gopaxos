# Leader as a State Machine
## States
- Ballot Number
- Activity State
- Ongoing Proposals (pvalues)
- Heartbeat State (only checked when a Scout is running)

## State Transisions
**T1 - Scout Adoption**
1. Set ```Active``` to true
2. Update ```Proposals``` with the pvalues the Scout collected

**P1 - Adoption Predicates**
1. ```Active``` is originally false
    - I think I can give a logical proof, but it would be a bit long
2. Last heartbeat check failed, no other active leader in this leader's view

**T2 - Preemption**
1. Set ``` Active``` as false
2. Increment ```Balllot Number``` by 1
3. Spawn a Scout

**P2 - Preemption Predicates**
1. The leader receives a P1B or P2B that has a ballot number larger than the leader's ```Ballot Number```
    - Yes, preemption happens IFF a larger ballot is discovered

**T3 - New Proposal**
1. Add a new proposal to ```Proposals```

**P3 - New Proposal Predicate**
1. The new proposal's slot number is not proposed by any other propsoal in ```Proposals```