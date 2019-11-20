# go-state

> coverage: 97.9%

My event based elm-like app state architecture designed to make stateful event driven logic more maintainable.

[godoc.org/github.com/joeycumines/go-state](https://godoc.org/github.com/joeycumines/go-state)

It's super minimal, runs all the core logic in the calling goroutine, tries very hard to obey context, and be as
un-opinionated as possible (for such an abstract implementation anyway). Three run modes are currently provided, `Run`,
`Batch`, and `Aggregate`, where `Batch` extends `Run` to add bounded message processing, and `Aggregate` extends `Batch`
to add support for model snapshots (backed by a generic key-value data store).

This library is designed to be used in tandem with the `Buffer` implementation from
[bigbuff](https://github.com/joeycumines/go-bigbuff), but it will work with any unbounded FIFO queue implementation
meeting the required interfaces.

## Performance

Extremely fast. Probably won't get around to proving that here so you will need to try it out. I have validated this
implementation in two very different production environments, the first being as a batched message processing mechanism
for AWS Kinesis (running AWS Lambda consumers), and the second being various bots for personal projects. The lambda was
scaled to the point where it cost hundreds of dollars per month and was constantly consuming from 30+ shards, while the
most impressive bot currently runs in a very fast 8 core VM, and processes messages that are generated from ~70Mb/s of
primarily polling based HTTP traffic (N.B. the event loop is a single goroutine, and works very well with sync.Map
based view implementations).

## Roadmap

I don't expect to be able to move fast on this project, but I do intend to implement some more useful examples,
drawing on my experience using this implementation in production environments. The biggest issue that I have found
with this pattern is that developer experience can suffer, generally because it's significantly more complex than
basically any alternative, and is that much harder to reason about, as a result.

## Implementation Notes

- Nil messages are ignored
- Nil commands are ignored
- Replay is facilitated by simply skipping commands and is a configurable option
- If your views are expensive you may need to implement a "refresh rate" mechanism
- If the system is going to be saturated with CPU bounded operations you will very likely have a bad time unless you
  are very careful about ensuring anything that produces messages compensates for the main event loop falling behind
