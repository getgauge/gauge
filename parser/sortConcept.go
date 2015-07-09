package parser

type ByLineNo []*Concept

func (s ByLineNo) Len() int {
	return len(s)
}

func (s ByLineNo) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByLineNo) Less(i, j int) bool {
	return s[i].ConceptStep.LineNo < s[j].ConceptStep.LineNo
}
