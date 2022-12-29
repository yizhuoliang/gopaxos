# Paxos Nodes Simulation with Colocality

## Accepotr Simulation

### States

- ```ballot```
- ```accepted```

### Transformation Functions

| curr_state             | incoming_message                                  | next_state                  |
| ---------------------- | ------------------------------------------------- | --------------------------- |
| S.ballot<br>S.accepted | m <P1A, b, $$\lambda$$> <br>m.b > S.ballot        | S.ballot = s.ballot         |
|                        | m <P1A, b, $$\lambda$$> <br>m.b <= S.ballot       | S                           |
|                        | m <P2A, $$\lambda$$, <b,s,c>><br>m.b = S.ballot   | S.accepted[m.s] = m.<b,s,c> |
|                        | m <P2A, $$\lambda$$, <b,s,c>><br/>m.b != S.ballot | S                           |
|                        |                                                   |                             |

### Basic Assertions on Outgoing Messeges (nothing going backwards, nothing deleted)

**m <P1B, b, accepted>** 

- S.ballot <= m.b
- S.accepted is a subset of m.accepted or they are equal

**m <P2B, b>**

- S.ballot <= m.b

### Inference State from Outgoing Messeges

**m <P1B, b, accepted>**

- S.ballot = m.b
- S.accepted = m.accepted

**m <P2B, b>**

- S.ballot = m.b

## Leader Simulation

### States

- ```ballot```
- ```proposals```
- ```pvalues```, ```p1b_count```,```p2b_count``` (for simulating the job of Scouts and commanders)

### Transofrmation Fucntions

| curr_state                                            | incoming_message                                        | next_state                      |
| ----------------------------------------------------- | ------------------------------------------------------- | ------------------------------- |
| S.ballot<br>S.proposals<br>S.p1b_count<br>S.p2b_count | m <propose, s, c> <br>m.propose $$\in$$ S.proposals     | S                               |
|                                                       | m <propose, s, c> <br/>m.propose $$\notin$$ S.proposals | S.proposals.add(m.<s, c>)       |
|                                                       | m <adoption,  b, pvalues><br>                           | update S.proposals with pvalues |
|                                                       | m <preemption, b><br/>m.b > S.ballot                    | S                               |



## Outgoing Message Allowed for each message

**Scout**
$$
\begin{aligned}
Scout\space such&\space that\\
&scout.ballot > S.ballot\\
\end{aligned}
$$
**Commander**
$$
\begin{aligned}
&Commander \space such\space that\\
&cmd.ballot = S.ballot\\
&cmd.proposal \in S.accepted \implies\\ cmd.proposal.<s,c> &=S.accepted.<s,c>
\end{aligned}
$$

## Three ways to go

1. simply do the total simulation, enforce all invariants
2. do not do total simulation, then we can never ensure all invariants, but detect errors in some situations by extra assertions
3. total simulation + redundency extra assertion