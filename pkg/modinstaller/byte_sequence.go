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

type ByteSequence struct {
	_underlying []byte
}

func NewByteSequence(b []byte) *ByteSequence {
	byteSequence := new(ByteSequence)
	byteSequence._underlying = make([]byte, len(b))
	copy(byteSequence._underlying, b)
	return byteSequence
}

func (b *ByteSequence) ApplyChanges(changeSet ChangeSet) {
	for _, change := range changeSet {
		switch change.Operation {
		case Insert:
			{
				b.insert(change)
			}
		case Delete:
			{
				b.clear(change)
			}
		case Replace:
			{
				b.clear(change)
				b.insert(change)
			}
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
func (bseq *ByteSequence) clear(change *Change) {
	left := bseq._underlying[:change.OffsetStart]
	right := bseq._underlying[change.OffsetEnd:]
	bseq._underlying = append(left, right...)
}

// insert inserts the given content at 'offset'
func (bseq *ByteSequence) insert(change *Change) {
	left := bseq._underlying[:change.OffsetStart]
	right := bseq._underlying[change.OffsetStart:]
	// prepend the content before the right part
	right = append(change.Content, right...)
	bseq._underlying = append(left, right...)
}
