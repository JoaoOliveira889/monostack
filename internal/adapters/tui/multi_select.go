package tui

import tea "github.com/charmbracelet/bubbletea"

type multiSelectSet map[int]bool

func newMultiSelectSet() multiSelectSet {
	return make(multiSelectSet)
}

func (ms multiSelectSet) toggle(index int) {
	if ms[index] {
		delete(ms, index)
	} else {
		ms[index] = true
	}
}

func (ms multiSelectSet) count() int {
	return len(ms)
}

func (ms multiSelectSet) contains(index int) bool {
	return ms[index]
}

func (ms multiSelectSet) clear() {
	for k := range ms {
		delete(ms, k)
	}
}

func (ms multiSelectSet) items() []int {
	result := make([]int, 0, len(ms))
	for k := range ms {
		result = append(result, k)
	}
	return result
}

func (m *Model) toggleMultiSelect() {
	ctx, index, ok := m.logSelectionContext()
	if !ok {
		return
	}

	ms := m.multiSelectFor(ctx)
	if ms == nil {
		return
	}

	if !m.multiSelectActive || m.selectionContext != ctx {
		m.clearSelection()
		m.multiSelectActive = true
		m.selectionContext = ctx
		for _, s := range m.allMultiSelects() {
			s.clear()
		}
		ms.toggle(index)
		return
	}

	ms.toggle(index)
	if ms.count() == 0 {
		m.multiSelectActive = false
		m.selectionContext = selectionNone
	}
}

func (m *Model) clearMultiSelect() {
	m.multiSelectActive = false
	for _, s := range m.allMultiSelects() {
		s.clear()
	}
}

func (m *Model) multiSelectFor(ctx selectionContext) *multiSelectSet {
	switch ctx {
	case selectionS3Buckets:
		return &m.s3MultiSelected
	case selectionS3Objects:
		return &m.s3ObjectsMultiSelected
	case selectionSQSQueues:
		return &m.sqsMultiSelected
	case selectionSQSTopics:
		return &m.sqsTopMultiSelected
	case selectionSNSTopics:
		return &m.snsMultiSelected
	case selectionSNSSubs:
		return &m.snsSubsMultiSelected
	case selectionSecrets:
		return &m.secretsMultiSelected
	}
	return nil
}

func (m *Model) allMultiSelects() []*multiSelectSet {
	return []*multiSelectSet{
		&m.s3MultiSelected,
		&m.s3ObjectsMultiSelected,
		&m.sqsMultiSelected,
		&m.sqsTopMultiSelected,
		&m.snsMultiSelected,
		&m.snsSubsMultiSelected,
		&m.secretsMultiSelected,
	}
}

func (m *Model) multiSelectCountForPanel() int {
	if !m.multiSelectActive {
		return 0
	}
	ms := m.multiSelectFor(m.selectionContext)
	if ms == nil {
		return 0
	}
	return ms.count()
}

func (m *Model) deleteMultiSelectedCmd() tea.Cmd {
	return m.executeMultiDeleteCmd()
}
