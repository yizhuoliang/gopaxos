## Replica states
- `state`
- `slot_in`
- `slot_out`
- `requests`
- `proposals`
- `decisions`

## Transition functions
| curr_state                                                           | incoming_message          | next_state                                                                                        |
|----------------------------------------------------------------------|---------------------------|---------------------------------------------------------------------------------------------------|
| slot_in < slot_out + WINDOW<br>decisions[slot_in] !ok                  | <request, c>              | `proposals`[slot_in] = <proposal, slot_in, c><br>`slot_in`++                                            |
|                                                                      | array of <decision, s, c> | `decisions`[s] = decision                                                                           |
| decisions[slot_out] ok                                               | array of <decision, s, c> | `decisions`[s] = decision<br>c.op(`state`)<br>`slot_out`++                                                  |
| decisions[slot_out] ok<br>proposals[slot_out] ok                       | array of <decision, s, c> | `decisions`[s] = decision<br>c.op(`state`)<br>`slot_out`++<br>`proposals` \ proposals[slot_out]                 |
| d := decisions[slot_out] ok<br>p := proposals[slot_out] ok<br>d.c != p.c | array of <decision, s, c> | `decisions`[s] = decision<br>c.op(`state`)<br>`slot_out`++<br>`proposals` \ proposals[slot_out]<br>`requests` U p.c |

## Predicates for outgoing message <propose, slot_in, c>
- slot_in < slot_out + WINDOW
- decisions[slot_in] is empty
- replica must received <request, c>