# go-state

> coverage: 96.5%

My event based elm-like app state architecture designed to make stateful event driven 
logic more maintainable.

[godoc.org/github.com/joeycumines/go-state](https://godoc.org/github.com/joeycumines/go-state)

This is a WIP, I am in the process of re-implementing the core logic to be as minimal as
possible, but there is a complete counter example (documentation + more tests to come).

It's super minimal, and runs all all the logic in the calling goroutine, and tries
very hard to obey context, and be as un-opinionated as possible (for such an abstract
implementation anyway). Three run modes are currently provided, which extend on each other,
`Run`, `Batch`, and `Aggregate`.

**UPDATE:** Implementation of `Batch` and `Aggregate` is complete, and close to 100%
test coverage (just waiting on time to finish off testing error cases in `Aggregate`).

## Dev Notes (so I don't forget, to be consolidated)

- Nil messages are ignored
- Nil command messages are ignored
- Commands are where external interactions should have, and their message return values should be the output if sync
- Replay is facilitated by simply skipping calling commands, which can be done via an option
