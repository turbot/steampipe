package modinstaller

import (
	"bufio"
	"bytes"
	"sort"
	"strings"
)

type ChangeOpType int

const (
	INSERT ChangeOpType = iota
	DELETE
	REPLACE
)

type Change struct {
	Content     []byte
	Operation   ChangeOpType
	OffsetStart int
	OffsetEnd   int
}

type ChangeSet []*Change

type ByteSequence struct {
	_underlying []byte
}

func NewByteSequence(b []byte) *ByteSequence {
	bSeq := new(ByteSequence)
	bSeq._underlying = make([]byte, len(b))
	copy(bSeq._underlying, b)
	return bSeq
}

func (bseq *ByteSequence) ApplyChanges(changeSet ChangeSet) {
	// sort the changes into descending order of byte offset
	// this way, when a change is applied, even if it's replacement
	// does not have the exact same bytes, we don't lose the offset information
	// of the changes preceeding it
	sort.Slice(changeSet, func(i, j int) bool {
		return changeSet[i].OffsetStart > changeSet[j].OffsetStart
	})

	for _, change := range changeSet {
		switch change.Operation {
		case INSERT:
			{
				bseq.append(change.OffsetStart, change.Content)
			}
		case DELETE:
			{
				bseq.clear(change.OffsetStart, change.OffsetEnd)
			}
		case REPLACE:
			{
				bseq.clear(change.OffsetStart, change.OffsetEnd)
				bseq.append(change.OffsetStart, change.Content)
			}
		}
	}

}

// TrimBlanks compresses multiple empty lines to a single empty line
func (bseq *ByteSequence) TrimBlanks() {
	sc := bufio.NewScanner(bytes.NewReader(bseq._underlying))
	writer := bytes.NewBuffer([]byte{})
	skipBlankLine := false
	for sc.Scan() {
		isBlankLine := len(strings.TrimSpace(sc.Text())) == 0
		if isBlankLine && skipBlankLine {
			continue
		}

		if writer.Len() > 0 {
			// we don't want the first line to be a newline
			writer.WriteByte('\n')
		}

		// write out the line
		writer.Write(sc.Bytes())

		// if this was a blank line, we want to skip the next ones
		skipBlankLine = isBlankLine
	}

	bseq._underlying = writer.Bytes()
}

func (bseq *ByteSequence) Bytes() []byte {
	return bseq._underlying
}

// clear replaces whatever is within [start,end] with white spaces
func (bseq *ByteSequence) clear(start int, end int) {
	left := bseq._underlying[:start]
	right := bseq._underlying[end:]
	bseq._underlying = append(left, right...)
}

// append inserts the given content at 'offset'
func (bseq *ByteSequence) append(offset int, content []byte) {
	left := bseq._underlying[:offset]
	right := bseq._underlying[offset:]
	// prepend the content before the right part
	right = append(content, right...)
	bseq._underlying = append(left, right...)
}

// Apply applies the given function on the byte sequence
func (bseq *ByteSequence) Apply(apply func([]byte) []byte) {
	bseq._underlying = apply(bseq._underlying)
}
