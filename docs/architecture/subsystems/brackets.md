# Bracket Handling Subsystem

This document details the bracket handling subsystem that enables hierarchical data processing in our FBP implementation.

## Core Components

### IP Type System

```go
type IPType int

const (
    NormalIP IPType = iota
    OpenBracket
    CloseBracket
)

type IP[T any] struct {
    Type     IPType
    Data     T
    Metadata map[string]any
}
```

### Bracket Tracker

```go
type BracketTracker struct {
    depth     int
    mu        sync.RWMutex
    onClose   func()
}

func NewBracketTracker(onClose func()) *BracketTracker {
    return &BracketTracker{
        onClose: onClose,
    }
}
```

### Substream Processor

```go
type SubstreamProcessor[T any] struct {
    tracker    *BracketTracker
    buffer     []T
    processEnd func([]T)
}

func NewSubstreamProcessor[T any](processEnd func([]T)) *SubstreamProcessor[T] {
    sp := &SubstreamProcessor[T]{
        processEnd: processEnd,
    }
    sp.tracker = NewBracketTracker(sp.onSubstreamComplete)
    return sp
}
```

## Bracket Operations

### Opening Brackets

```go
func (bt *BracketTracker) OpenBracket() {
    bt.mu.Lock()
    defer bt.mu.Unlock()
    bt.depth++
}

func NewOpenBracket[T any]() *IP[T] {
    return &IP[T]{
        Type:     OpenBracket,
        Metadata: make(map[string]any),
    }
}
```

### Closing Brackets

```go
func (bt *BracketTracker) CloseBracket() {
    bt.mu.Lock()
    defer bt.mu.Unlock()
    bt.depth--
    if bt.depth == 0 && bt.onClose != nil {
        bt.onClose()
    }
}

func NewCloseBracket[T any]() *IP[T] {
    return &IP[T]{
        Type:     CloseBracket,
        Metadata: make(map[string]any),
    }
}
```

## Substream Processing

### Processing Logic

```go
func (sp *SubstreamProcessor[T]) ProcessIP(ip *IP[T]) {
    switch ip.Type {
    case OpenBracket:
        sp.tracker.OpenBracket()
    case CloseBracket:
        sp.tracker.CloseBracket()
    case NormalIP:
        sp.buffer = append(sp.buffer, ip.Data)
    }
}

func (sp *SubstreamProcessor[T]) onSubstreamComplete() {
    if sp.processEnd != nil {
        sp.processEnd(sp.buffer)
    }
    sp.buffer = sp.buffer[:0] // Clear buffer
}
```

## Bracketed Node Implementation

```go
type BracketedNode[In, Out any] struct {
    processor func([]In) []Out
}

func (bn *BracketedNode[In, Out]) Process(ctx context.Context, in Port[*IP[In]], out Port[*IP[Out]]) error {
    sp := NewSubstreamProcessor(func(data []In) {
        // Process entire substream when complete
        results := bn.processor(data)
        
        // Send results as a new bracketed stream
        out <- NewOpenBracket[Out]()
        for _, result := range results {
            out <- NewIP(result)
        }
        out <- NewCloseBracket[Out]()
    })

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case ip, ok := <-in:
            if !ok {
                return nil
            }
            sp.ProcessIP(ip)
        }
    }
}
```

## Example Usage

### Word Counter Example

```go
type WordCounter struct {
    *BracketedNode[string, int]
}

func NewWordCounter() *WordCounter {
    return &WordCounter{
        BracketedNode: NewBracketedNode(func(lines []string) []int {
            var wordCounts []int
            for _, line := range lines {
                words := strings.Fields(line)
                wordCounts = append(wordCounts, len(words))
            }
            return wordCounts
        }),
    }
}

// Example usage:
func Example_WordCounter() {
    counter := NewWordCounter()
    in := make(Port[*IP[string]], 10)
    out := make(Port[*IP[int]], 10)

    go func() {
        // First paragraph
        in <- NewOpenBracket[string]()
        in <- NewIP("Hello world")
        in <- NewIP("How are you")
        in <- NewCloseBracket[string]()

        // Second paragraph
        in <- NewOpenBracket[string]()
        in <- NewIP("Another paragraph")
        in <- NewIP("With more text")
        in <- NewIP("And another line")
        in <- NewCloseBracket[string]()

        close(in)
    }()

    go counter.Process(context.Background(), in, out)
}
```

## Best Practices

1. **Bracket Matching**
   - Always ensure brackets are properly matched
   - Use defer for automatic bracket closing
   - Handle errors without leaving unmatched brackets

2. **Memory Management**
   - Clear buffers after processing
   - Consider buffer size limits
   - Handle large substreams efficiently

3. **Error Handling**
   - Properly propagate errors during processing
   - Clean up resources on errors
   - Maintain bracket consistency during errors

4. **Thread Safety**
   - Use proper synchronization
   - Protect shared state
   - Handle concurrent substreams

## Common Use Cases

1. **Document Processing**
   - Processing paragraphs in text
   - Handling nested XML/JSON structures
   - Processing multi-record formats

2. **Hierarchical Data**
   - Processing tree structures
   - Handling nested transactions
   - Processing grouped records

3. **Batch Processing**
   - Processing record batches
   - Handling grouped operations
   - Managing transaction boundaries