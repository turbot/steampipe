package modinstaller

import (
	"sort"
)

type ChangeOperation int

const (
	Insert ChangeOperation = iota
	Delete
	Replace
)

type Change struct {
	Content     []byte
	Operation   ChangeOperation
	OffsetStart int
	OffsetEnd   int
}

type ChangeSet []*Change

func NewChangeSet(changeSets ...ChangeSet) ChangeSet {
	changeSet := ChangeSet{}
	for _, cs := range changeSets {
		changeSet = append(changeSet, cs...)
	}
	return changeSet
}

func (c ChangeSet) SortByOffset() {
	// sort the changes into descending order of byte offset
	// this way, when a change is applied, even if it's replacement
	// does not have the exact same bytes, we don't lose the offset information
	// of the changes preceeding it
	sort.Slice(c, func(i, j int) bool {
		return c[i].OffsetStart > c[j].OffsetStart
	})
}

type OperatorFunc func(*Change, []byte) []byte

type ByteSequence struct {
	operators   map[ChangeOperation]OperatorFunc
	_underlying []byte
}

func NewByteSequence(b []byte) *ByteSequence {
	byteSequence := new(ByteSequence)
	byteSequence._underlying = make([]byte, len(b))
	copy(byteSequence._underlying, b)

	byteSequence.operators = map[ChangeOperation]OperatorFunc{
		Insert:  insert,
		Delete:  clear,
		Replace: replace,
	}

	return byteSequence
}

func (b *ByteSequence) ApplyChanges(changeSet ChangeSet) {
	for _, change := range changeSet {
		operation := change.Operation
		if operator, ok := b.operators[operation]; ok {
			b._underlying = operator(change, b._underlying)
		}
	}
}

// Apply applies the given function on the byte sequence
func (bseq *ByteSequence) Apply(apply func([]byte) []byte) {
	bseq._underlying = apply(bseq._underlying)
}

func (bseq *ByteSequence) Bytes() []byte {
	return bseq._underlying
}

// clear replaces whatever is within [start,end] with white spaces
func clear(change *Change, source []byte) []byte {
	left := source[:change.OffsetStart]
	right := source[change.OffsetEnd:]
	return append(left, right...)
}

// insert inserts the given content at 'offset'
func insert(change *Change, source []byte) []byte {
	left := source[:change.OffsetStart]
	right := source[change.OffsetStart:]
	// prepend the content before the right part
	right = append(change.Content, right...)
	return append(left, right...)
}

func replace(change *Change, source []byte) []byte {
	return insert(change, clear(change, source))
}
