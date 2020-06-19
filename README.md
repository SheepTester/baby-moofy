# baby-moofy
A Discord bot in Go that uses Markov chains and listens to conversations to learn how to speak

### Things to add

- [ ] command for changing local delay (so allow frequent responses in a spam channel)

- [ ] save channel settings to JSON file

- [x] allow "..." for moofy to continue without a message break: "how are you ... /" will start with "how are you" rather than "are you /"

	- [ ] maybe a message that is just "..." will use the previous message: "how are you / ... /" will start with "how are you"

- [x] include moofy's own contributions for context when learning because people react to moofy
