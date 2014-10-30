package main

type conceptDictionary struct {
	conceptsMap     map[string]*concept
	constructionMap map[string]*step
	referenceMap    map[*step][]*step
}

type concept struct {
	conceptStep *step
	fileName    string
}

type conceptParser struct {
	currentState   int
	currentConcept *step
}

//concept file can have multiple concept headings
func (parser *conceptParser) parse(text string) ([]*step, *parseError) {
	defer parser.resetState()

	specParser := new(specParser)
	tokens, err := specParser.generateTokens(text)
	if err != nil {
		return nil, err
	}
	return parser.createConcepts(tokens)
}

func (parser *conceptParser) resetState() {
	parser.currentState = initial
	parser.currentConcept = nil
}

func (parser *conceptParser) createConcepts(tokens []*token) ([]*step, *parseError) {
	parser.currentState = initial
	concepts := make([]*step, 0)
	var error *parseError
	for _, token := range tokens {
		if parser.isConceptHeading(token) {
			if isInState(parser.currentState, conceptScope, stepScope) {
				concepts = append(concepts, parser.currentConcept)
			}
			parser.currentConcept, error = parser.processConceptHeading(token)
			if error != nil {
				return nil, error
			}
			addStates(&parser.currentState, conceptScope)
		} else if parser.isStep(token) {
			if !isInState(parser.currentState, conceptScope) {
				return nil, &parseError{lineNo: token.lineNo, message: "Step is not defined inside a concept heading", lineText: token.lineText}
			}
			if err := parser.processConceptStep(token); err != nil {
				return nil, err
			}
			addStates(&parser.currentState, stepScope)
		} else if parser.isTableHeader(token) {
			if !isInState(parser.currentState, stepScope) {
				return nil, &parseError{lineNo: token.lineNo, message: "Table doesn't belong to any step", lineText: token.lineText}
			}
			parser.processTableHeader(token)
			addStates(&parser.currentState, tableScope)
		} else if parser.isTableDataRow(token) {
			parser.processTableDataRow(token, &parser.currentConcept.lookup)
		}
	}
	if !isInState(parser.currentState, stepScope) && parser.currentState != initial {
		return nil, &parseError{lineNo: parser.currentConcept.lineNo, message: "Concept should have atleast one step", lineText: parser.currentConcept.lineText}
	}

	if parser.currentConcept != nil {
		concepts = append(concepts, parser.currentConcept)
	}
	return concepts, nil
}

func (parser *conceptParser) isConceptHeading(token *token) bool {
	return token.kind == specKind || token.kind == scenarioKind
}

func (parser *conceptParser) isStep(token *token) bool {
	return token.kind == stepKind
}

func (parser *conceptParser) isTableHeader(token *token) bool {
	return token.kind == tableHeader
}

func (parser *conceptParser) isTableDataRow(token *token) bool {
	return token.kind == tableRow
}

func (parser *conceptParser) processConceptHeading(token *token) (*step, *parseError) {
	processStep(new(specParser), token)
	if !parser.hasOnlyDynamicParams(token) {
		return nil, &parseError{lineNo: token.lineNo, message: "Concept heading can have only Dynamic Parameters"}
	}
	concept, err := new(specification).createStepUsingLookup(token, nil)
	if err != nil {
		return nil, err
	}
	concept.isConcept = true
	parser.createConceptLookup(concept)
	return concept, nil

}

func (parser *conceptParser) processConceptStep(token *token) *parseError {
	processStep(new(specParser), token)
	conceptStep, err := new(specification).createStepUsingLookup(token, &parser.currentConcept.lookup)
	if err != nil {
		return err
	}
	parser.currentConcept.conceptSteps = append(parser.currentConcept.conceptSteps, conceptStep)
	return nil
}

func (parser *conceptParser) processTableHeader(token *token) {
	steps := parser.currentConcept.conceptSteps
	currentStep := steps[len(steps)-1]
	addInlineTableHeader(currentStep, token)
}

func (parser *conceptParser) processTableDataRow(token *token, argLookup *argLookup) {
	steps := parser.currentConcept.conceptSteps
	currentStep := steps[len(steps)-1]
	addInlineTableRow(currentStep, token, argLookup)
}

func (parser *conceptParser) hasOnlyDynamicParams(token *token) bool {
	_, kinds := extractStepValueAndParameterTypes(token.value)
	for _, argKind := range kinds {
		if argKind != "dynamic" {
			return false
		}
	}
	return true
}

func (parser *conceptParser) createConceptLookup(concept *step) {
	for _, arg := range concept.args {
		concept.lookup.addArgName(arg.value)
	}
}

func (conceptDictionary *conceptDictionary) isConcept(step *step) bool {
	_, ok := conceptDictionary.conceptsMap[step.value]
	return ok

}
func (conceptDictionary *conceptDictionary) add(concepts []*step, conceptFile string) *parseError {
	if conceptDictionary.conceptsMap == nil {
		conceptDictionary.conceptsMap = make(map[string]*concept)
	}
	if conceptDictionary.constructionMap == nil {
		conceptDictionary.constructionMap = make(map[string]*step)
	}
	for _, conceptStep := range concepts {
		if _, exists := conceptDictionary.conceptsMap[conceptStep.value]; exists {
			return &parseError{message: "Duplicate concept definition found", lineNo: conceptStep.lineNo, lineText: conceptStep.lineText}
		}
		conceptDictionary.replaceNestedConceptSteps(conceptStep)
		conceptDictionary.conceptsMap[conceptStep.value] = &concept{conceptStep, conceptFile}
	}
	conceptDictionary.updateLookupForNestedConcepts()
	return nil
}

func (conceptDictionary *conceptDictionary) search(stepValue string) *concept {
	if concept, ok := conceptDictionary.conceptsMap[stepValue]; ok {
		return concept
	}
	return nil
}

func (conceptDictionary *conceptDictionary) replaceNestedConceptSteps(conceptStep *step) {
	conceptDictionary.updateStep(conceptStep)
	for i, stepInsideConcept := range conceptStep.conceptSteps {
		if nestedConcept := conceptDictionary.search(stepInsideConcept.value); nestedConcept != nil {
			//replace step with actual concept
			conceptStep.conceptSteps[i] = nestedConcept.conceptStep
		} else {
			conceptDictionary.updateStep(stepInsideConcept)
		}
	}
}

//mutates the step with concept steps so that anyone who is referencing the step will now refer a concept
func (conceptDictionary *conceptDictionary) updateStep(step *step) {
	if conceptDictionary.constructionMap[step.value] == nil {
		conceptDictionary.constructionMap[step.value] = step
	} else {
		conceptDictionary.constructionMap[step.value].isConcept = step.isConcept
		conceptDictionary.constructionMap[step.value].conceptSteps = step.conceptSteps
		conceptDictionary.constructionMap[step.value].lookup = step.lookup
	}
}

func (conceptDictionary *conceptDictionary) updateLookupForNestedConcepts() {
	for _, concept := range conceptDictionary.conceptsMap {
		for _, stepInsideConcept := range concept.conceptStep.conceptSteps {
			if nestedConcept := conceptDictionary.search(stepInsideConcept.value); nestedConcept != nil {
				for i, arg := range nestedConcept.conceptStep.args {
					nestedConcept.conceptStep.lookup.addArgValue(arg.name, &stepArg{argType: dynamic, value: stepInsideConcept.args[i].name})
				}
			}
		}
	}
}

func (self *concept) deepCopy() *concept {
	return &concept{fileName: self.fileName, conceptStep: self.conceptStep.getCopy()}
}
