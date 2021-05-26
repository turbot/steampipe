package modconfig

// ReportTreeItemDiffs is a struct representing the differences between 2 ReportTreeItems (of same type)
type ReportTreeItemDiffs struct {
	Name              string
	Item              ReportTreeItem
	ChangedProperties []string
	AddedPanels       []string
	RemovedPanels     []string
	AddedReports      []string
	RemovedReports    []string
}

func (d *ReportTreeItemDiffs) AddPropertyDiff(propertyName string) {
	d.ChangedProperties = append(d.ChangedProperties, propertyName)
}

func (d *ReportTreeItemDiffs) AddAddedPanel(panelName string) {
	d.ChangedProperties = append(d.AddedPanels, panelName)
}

func (d *ReportTreeItemDiffs) AddDeletedPanel(panelName string) {
	d.ChangedProperties = append(d.RemovedPanels, panelName)
}

func (d *ReportTreeItemDiffs) AddAddedReport(reportName string) {
	d.ChangedProperties = append(d.AddedReports, reportName)
}

func (d *ReportTreeItemDiffs) AddDeletedReport(reportName string) {
	d.ChangedProperties = append(d.RemovedReports, reportName)
}

func (d *ReportTreeItemDiffs) populateChildDiffs(old ReportTreeItem, new ReportTreeItem) {
	// build map of panel and report names
	childPanelMap := make(map[string]bool)
	otherChildPanelMap := make(map[string]bool)
	childReportMap := make(map[string]bool)
	otherChildReportMap := make(map[string]bool)
	for _, childPanel := range old.GetPanels() {
		childPanelMap[childPanel.Name()] = true
	}
	for _, childPanel := range new.GetPanels() {
		otherChildPanelMap[childPanel.Name()] = true
	}
	for _, childReport := range old.GetReports() {
		childReportMap[childReport.Name()] = true
	}
	for _, childReport := range new.GetReports() {
		otherChildReportMap[childReport.Name()] = true
	}
	for panelName := range childPanelMap {
		if !otherChildPanelMap[panelName] {
			d.AddDeletedPanel(panelName)
		}
	}
	for panelName := range otherChildPanelMap {
		if !childPanelMap[panelName] {
			d.AddAddedPanel(panelName)
		}
	}
	for reportName := range childReportMap {
		if !otherChildReportMap[reportName] {
			d.AddDeletedReport(reportName)
		}
	}
	for reportName := range otherChildReportMap {
		if !childReportMap[reportName] {
			d.AddAddedReport(reportName)
		}
	}
}

func (d *ReportTreeItemDiffs) HasChanges() bool {
	return len(d.ChangedProperties)+
		len(d.AddedPanels)+
		len(d.AddedReports)+
		len(d.RemovedPanels)+
		len(d.RemovedReports) > 0
}
