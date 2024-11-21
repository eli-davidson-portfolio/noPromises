# Advanced Bracket Patterns

## Overview

Brackets in Flow-Based Programming (FBP) provide a powerful mechanism for handling hierarchical data structures and grouped processing. This document details advanced patterns and implementations for bracket handling.

## Core Concepts

### Bracket Types

```go
type BracketType int

const (
    OpenBracket BracketType = iota
    CloseBracket
)

type IP[T any] struct {
    Type     IPType
    Data     T
    Metadata map[string]any
    BracketDepth int   // Current nesting level
}
```

### Substream Structure

```go
type Substream[T any] struct {
    Data     []T
    Metadata map[string]any
    Parent   *Substream[T]    // For nested substreams
    Children []*Substream[T]  // For nested substreams
    Depth    int             // Nesting level
}
```

## Advanced Patterns

### 1. Nested Substream Processing

```go
type NestedProcessor[T any] struct {
    currentSubstream *Substream[T]
    substreams      []*Substream[T]
    bracketDepth    int
}

func (p *NestedProcessor[T]) ProcessIP(ip *IP[T]) error {
    switch ip.Type {
    case OpenBracket:
        return p.handleOpenBracket(ip)
    case CloseBracket:
        return p.handleCloseBracket(ip)
    default:
        return p.processData(ip)
    }
}

func (p *NestedProcessor[T]) handleOpenBracket(ip *IP[T]) error {
    newSubstream := &Substream[T]{
        Metadata: ip.Metadata,
        Parent:   p.currentSubstream,
        Depth:    p.bracketDepth,
    }
    
    if p.currentSubstream != nil {
        p.currentSubstream.Children = append(p.currentSubstream.Children, newSubstream)
    }
    
    p.currentSubstream = newSubstream
    p.bracketDepth++
    return nil
}
```

### 2. Hierarchical Data Processing

```go
type HierarchicalProcessor[T any] struct {
    processor func(*Substream[T]) error
    bracketTracker *BracketTracker
}

func (p *HierarchicalProcessor[T]) Process(ctx context.Context, in Port[IP[T]], out Port[IP[T]]) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case ip := <-in:
            if err := p.handleIP(ip, out); err != nil {
                return err
            }
        }
    }
}

func (p *HierarchicalProcessor[T]) handleIP(ip IP[T], out Port[IP[T]]) error {
    switch ip.Type {
    case OpenBracket:
        p.bracketTracker.OpenBracket()
        return out.Send(ip)
    case CloseBracket:
        if err := p.processor(p.bracketTracker.CurrentSubstream()); err != nil {
            return err
        }
        p.bracketTracker.CloseBracket()
        return out.Send(ip)
    default:
        p.bracketTracker.AddData(ip.Data)
        return out.Send(ip)
    }
}
```

### 3. Bracket Synchronization

```go
type BracketSynchronizer[T any] struct {
    inputs     []Port[IP[T]]
    output     Port[IP[T]]
    trackers   []*BracketTracker
    sync       *sync.WaitGroup
}

func (s *BracketSynchronizer[T]) Synchronize(ctx context.Context) error {
    // Start tracking each input
    for i := range s.inputs {
        s.sync.Add(1)
        go s.trackInput(ctx, i)
    }
    
    // Wait for all inputs to complete
    s.sync.Wait()
    
    // Process synchronized substreams
    return s.processSynchronizedStreams()
}

func (s *BracketSynchronizer[T]) trackInput(ctx context.Context, index int) {
    defer s.sync.Done()
    
    for {
        select {
        case <-ctx.Done():
            return
        case ip := <-s.inputs[index]:
            s.trackers[index].Track(ip)
            if ip.Type == CloseBracket && s.trackers[index].Depth() == 0 {
                return // Substream complete
            }
        }
    }
}
```

### 4. Error Handling in Substreams

```go
type SubstreamErrorHandler[T any] struct {
    processor  ProcessFunc[T]
    onError    ErrorFunc[T]
    bracketTracker *BracketTracker
}

func (h *SubstreamErrorHandler[T]) Process(ip IP[T]) error {
    switch ip.Type {
    case OpenBracket:
        h.bracketTracker.OpenBracket()
        return nil
        
    case CloseBracket:
        defer h.bracketTracker.CloseBracket()
        
        // Process current substream
        substream := h.bracketTracker.CurrentSubstream()
        if err := h.processor(substream); err != nil {
            // Handle error while maintaining bracket integrity
            if err2 := h.handleSubstreamError(substream, err); err2 != nil {
                return fmt.Errorf("handling substream error: %w", err2)
            }
            return nil
        }
        return nil
        
    default:
        return h.processor(ip)
    }
}

func (h *SubstreamErrorHandler[T]) handleSubstreamError(substream *Substream[T], err error) error {
    // Create error substream
    errorStream := &Substream[T]{
        Metadata: map[string]any{
            "error": err.Error(),
            "originalMetadata": substream.Metadata,
        },
        Data: substream.Data,
    }
    
    return h.onError(errorStream)
}
```

## Best Practices

### 1. Bracket Integrity
- Always match open/close brackets
- Maintain proper nesting
- Handle errors without breaking bracket structure
- Clean up brackets on context cancellation

### 2. Substream Processing
- Process complete substreams
- Handle nested structures properly
- Maintain parent-child relationships
- Support recursive processing

### 3. Error Management
- Preserve bracket structure during errors
- Handle partial substreams
- Clean up resources properly
- Maintain data consistency

### 4. Performance Considerations
- Buffer substreams appropriately
- Handle large hierarchies efficiently
- Process substreams concurrently when possible
- Monitor memory usage

## Common Use Cases

### 1. Document Processing
```go
type DocumentProcessor struct {
    HierarchicalProcessor[Document]
}

func (p *DocumentProcessor) ProcessSubstream(substream *Substream[Document]) error {
    // Process document structure
    for _, child := range substream.Children {
        // Process sections
        for _, section := range child.Data {
            // Process content
        }
    }
    return nil
}
```

### 2. Transaction Grouping
```go
type TransactionGroup struct {
    HierarchicalProcessor[Transaction]
}

func (g *TransactionGroup) ProcessSubstream(substream *Substream[Transaction]) error {
    // Calculate group totals
    var total decimal.Decimal
    for _, tx := range substream.Data {
        total = total.Add(tx.Amount)
    }
    
    // Process group
    return g.processGroup(total, substream.Metadata)
}
```

### 3. Data Transformation
```go
type DataTransformer[In, Out any] struct {
    HierarchicalProcessor[In]
    transform func(In) Out
}

func (t *DataTransformer[In, Out]) ProcessSubstream(substream *Substream[In]) error {
    // Transform each item while maintaining structure
    for _, item := range substream.Data {
        transformed := t.transform(item)
        t.output.Send(NewIP(transformed))
    }
    return nil
}
```

## Testing Patterns

### 1. Bracket Structure Testing
```go
func TestBracketStructure(t *testing.T) {
    processor := NewHierarchicalProcessor[string]()
    
    input := []IP[string]{
        {Type: OpenBracket},
        {Data: "test1"},
        {Type: OpenBracket},
        {Data: "test2"},
        {Type: CloseBracket},
        {Data: "test3"},
        {Type: CloseBracket},
    }
    
    result := processor.Process(input)
    assert.Equal(t, 2, len(result.Substreams))
    assert.Equal(t, 1, len(result.Substreams[0].Children))
}
```

### 2. Error Handling Testing
```go
func TestSubstreamError(t *testing.T) {
    handler := NewSubstreamErrorHandler[string]()
    
    // Test error in nested substream
    input := []IP[string]{
        {Type: OpenBracket},
        {Data: "will-error"},
        {Type: CloseBracket},
    }
    
    err := handler.Process(input)
    assert.NoError(t, err) // Should handle error gracefully
    assert.Equal(t, 1, handler.ErrorCount())
}
```

## Implementation Notes

1. Always maintain bracket integrity even during errors
2. Use proper synchronization for concurrent processing
3. Implement efficient buffer management
4. Handle resource cleanup properly
5. Maintain clear parent-child relationships
6. Provide proper error context

This document serves as a comprehensive guide to advanced bracket handling in our FBP implementation. For basic bracket concepts, refer to the core documentation.