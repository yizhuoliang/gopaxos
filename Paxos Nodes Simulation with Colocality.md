# Paxos Nodes Simulation with Colocality

A replica, a leader, and an aceeptor run at the same place, assume no network issues inside a single bundle.



## Invariants we pay attention to

**R1, A5, L2:** no two different commands decided for the same slot across the system

**A1, 2, 3:** an acceptor only adopt strictly increasing ballot_numbers, only accept proposals when the ballot matches, and does not remove pvalues from ```accepted```

**A4:** if an acceptor accepted $$<b,s,c>$$ and another acceptor accepted $$<b,s,c'>\implies$$ $$c = c'$$



## Accepotr Simulation

### States

- ```ballot```
- ```accepted```

### Transformation Functions

$$
\begin{aligned}
&T(S, P1A): \\
&\qquad P1A.ballot > S.ballot \implies S.ballot = P1A.ballot\\
\\
&T(S, P2A): \\
&\qquad P2A.ballot = S.ballot \implies S.accepted.add(P2B.bsc)\\
\end{aligned}
$$

### Predicates

$$
\begin{align}
&\{P1B \mid ballot = S.ballot,\space accepted = accepted \}\\
&\{P2B \mid ballot = S.ballot \}
\end{align}
$$



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
& Commander \space such\space that\\
&cmd.ballot = S.ballot\\
&cmd.proposal \in S.accepted \implies\\ cmd.proposal.<s,c> &=S.accepted.<s,c>
\end{aligned}
$$


### Extra assertions that not happen with outgoing messages

1. when a leader scout receive pvalues, there cannot be $$<b,s,c>$$ and $$<b,s,c'>$$ , $$c$$ must equals $$c'$$
2. what about checking L2 directly?
3. nothing invented?



## Two ways to go

1. simply do the total simulation
2. do not do total simulation, then we can never ensure all invariants, but detect errors in some situations
3. total simulation + extra assertion