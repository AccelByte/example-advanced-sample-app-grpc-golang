## Crew and Game MatchMakers

### Summary

The `Crew` server will take in Match Ticket Requests and match them together to form a Crew (which can be thought of as a Party, Team, etc).
When matched, a `crewType` will be added to the MatchAttributes which will identify them as a `mercenary` Crew. These Attributes are then passed to 
the Session that is created for this Crew and can be appended to a new Match Ticket Request (under the TicketAttribute key) when they queue up for a Game.

The `Game` server will Validate that the ticket has the correct `crewType` TicketAttribute to it before allowing it
to be considered in Match Making.

### Getting the passed-along Match Attributes

When the Crew server builds a crew, it will add attributes to the MatchAttributes key in the Match struct.
These added Attributes are then added to the Session's Attributes (remember, a Session is a Game, Party, etc). To add these
Attributes to the Match Ticket, simply send a GET using the SessionID of the Crew and append those Attributes to any other
Match Ticket attributes before sending a Match Request to the Game server.