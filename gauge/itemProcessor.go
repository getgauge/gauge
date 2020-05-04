/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

type ItemProcessor interface {
	Specification(*Specification)
	Heading(*Heading)
	Tags(*Tags)
	Table(*Table)
	DataTable(*DataTable)
	Scenario(*Scenario)
	Step(*Step)
	TearDown(*TearDown)
	Comment(*Comment)
}
