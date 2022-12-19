# Paxos Nodes Simulation with Colocality

A replica, a leader, and an aceeptor run at the same place, assume no network issues inside a single bundle.

## Invariants we pay attention to

**R1, A5, L2:** no two different commands decided for the same slot across the system

**A1, 2, 3:** an acceptor only adopt strictly increasing ballot_numbers, only accept proposals when the ballot matches, and does not remove pvalues from ```accepted```

**A4:** if an acceptor accepted $$<b,s,c>$$ and another acceptor accepted $$<b,s,c'>\implies$$ $$c = c'$$

### Extra assertions that not happen with outgoing messages

1. when a leader scout receive pvalues, there cannot be $$<b,s,c>$$ and $$<b,s,c'>$$ , $$c$$ must equals $$c'$$ [A4]
2. Check if a replica receive same slot decision but different command
3. Chek if a leader ever spwans same bs but different c
4. check if the proposals sent are ever received, check if the pvalues sent are legally accepeted

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

### Assertions on Outgoing Messeges (at least make sure everything is non-decreasing)

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
- There exist a **p <b, s, c>** in S.accepted that p.b = m.b // delete? 

## Leader Simulation

### Transofrmation Fucntions

$$
\begin{aligned}
T(S, prop):\qquad \qquad \qquad \qquad \\
propo.slot \notin proposals &\rightarrow S.proposals.add(prop)
\end{aligned}
$$

$$
\begin{aligned}
T(S, p1b):\qquad \qquad \qquad \qquad \\
prop.slot \in proposals &\rightarrow S\\
propo.slot \notin proposals &\rightarrow S.proposals.add(prop)
\end{aligned}
$$

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