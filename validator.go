package main

import "github.com/golang/protobuf/proto"

type validator struct {
	manifest           *manifest
	specsToExecute     []*specification
	runner             *testRunner
	conceptsDictionary *conceptDictionary
}

type specValidator struct {
	specification        *specification
	runner               *testRunner
	conceptsDictionary   *conceptDictionary
	stepValidationErrors []*stepValidationError
	stepValidationCache  map[string]bool
}

type stepValidationError struct {
	step     *step
	message  string
	fileName string
}

func (e *stepValidationError) Error() string {
	return e.message
}

type executionValidationErrors map[*specification][]*stepValidationError

func newValidator(manifest *manifest, specsToExecute []*specification, runner *testRunner, conceptsDictionary *conceptDictionary) *validator {
	return &validator{manifest: manifest, specsToExecute: specsToExecute, runner: runner, conceptsDictionary: conceptsDictionary}
}

func (self *validator) validate() executionValidationErrors {
	validationStatus := make(executionValidationErrors)
	specValidator := &specValidator{runner: self.runner, conceptsDictionary: self.conceptsDictionary, stepValidationCache: make(map[string]bool)}
	for _, spec := range self.specsToExecute {
		specValidator.specification = spec
		validationErrors := specValidator.validate()
		if len(validationErrors) != 0 {
			validationStatus[spec] = validationErrors
		}
	}
	if len(validationStatus) > 0 {
		return validationStatus
	} else {
		return nil
	}
}

func (self *specValidator) validate() []*stepValidationError {
	self.specification.traverse(self)
	return self.stepValidationErrors
}

func (self *specValidator) step(step *step) {
	if step.isConcept {
		for _, conceptStep := range step.conceptSteps {
			if _, ok := self.stepValidationCache[conceptStep.value]; !ok {
				self.stepValidationCache[conceptStep.value] = true
				self.step(conceptStep)
			}
		}
	} else if _, ok := self.stepValidationCache[step.value]; !ok {
		self.stepValidationCache[step.value] = true
		self.validateStep(step)
	}
}

func (self *specValidator) validateStep(step *step) {
	message := &Message{MessageType: Message_StepValidateRequest.Enum(),
		StepValidateRequest: &StepValidateRequest{StepText: proto.String(step.value), NumberOfParameters: proto.Int(len(step.args))}}
	response, err := getResponseForGaugeMessage(message, self.runner.connection)
	if err != nil {
		self.stepValidationErrors = append(self.stepValidationErrors, &stepValidationError{step: step, message: err.Error(), fileName: self.specification.fileName})
	}
	if response.GetMessageType() == Message_StepValidateResponse {
		validateResponse := response.GetStepValidateResponse()
		if !validateResponse.GetIsValid() {
			self.stepValidationErrors = append(self.stepValidationErrors, &stepValidationError{step: step, message: validateResponse.GetErrorMessage(), fileName: self.specification.fileName})
		}
	} else {
		self.stepValidationErrors = append(self.stepValidationErrors, &stepValidationError{step: step, message: "Invalid response from runner for Validation request", fileName: self.specification.fileName})
	}
}

func (self *specValidator) contextStep(step *step) {
	self.step(step)
}

func (self *specValidator) specHeading(heading *heading) {
	self.stepValidationErrors = make([]*stepValidationError, 0)
}

func (self *specValidator) specTags(tags *tags) {

}

func (self *specValidator) scenarioTags(tags *tags) {

}

func (self *specValidator) dataTable(dataTable *table) {

}

func (self *specValidator) scenario(scenario *scenario) {

}

func (self *specValidator) scenarioHeading(heading *heading) {
}

func (self *specValidator) comment(comment *comment) {

}
