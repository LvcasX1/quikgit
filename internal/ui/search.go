package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	ghClient "github.com/lvcasx1/quikgit/internal/github"
)

// performSearchEnhanced executes the repository search
func (a *Application) performSearchEnhanced() {
	// Get form values
	query := strings.TrimSpace(a.searchForm.GetFormItemByLabel("Query").(*tview.InputField).GetText())
	if query == "" {
		a.error = fmt.Errorf("search query cannot be empty")
		a.showSearch()
		return
	}

	// Get dropdown values
	_, language := a.searchForm.GetFormItemByLabel("Language").(*tview.DropDown).GetCurrentOption()
	_, sortBy := a.searchForm.GetFormItemByLabel("Sort by").(*tview.DropDown).GetCurrentOption()
	includeForks := a.searchForm.GetFormItemByLabel("Include forks").(*tview.Checkbox).IsChecked()

	// Show loading
	a.showSearchProgress("Searching repositories...")

	// Perform search in background
	go func() {
		results, err := a.searchRepositories(query, language, sortBy, includeForks)

		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.error = err
				a.showSearch()
			} else {
				a.searchResults = results
				a.showSearchResults()
			}
		})
	}()
}

// searchRepositories performs the actual GitHub search
func (a *Application) searchRepositories(query, language, sortBy string, includeForks bool) ([]*ghClient.Repository, error) {
	if a.githubClient == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	// Build search query
	searchQuery := query
	if language != "Any" {
		searchQuery += " language:" + strings.ToLower(language)
	}
	if !includeForks {
		searchQuery += " fork:false"
	}

	// Map sort options
	var sort string
	switch sortBy {
	case "Stars":
		sort = "stars"
	case "Forks":
		sort = "forks"
	case "Updated":
		sort = "updated"
	case "Created":
		sort = "created"
	default:
		sort = "" // best match
	}

	opts := ghClient.SearchOptions{
		Query: searchQuery,
		Sort:  sort,
		Order: "desc",
		Page:  1,
		Limit: 30,
	}

	// Perform search
	repos, _, err := a.githubClient.SearchRepositories(a.ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return repos, nil
}

// showSearchProgress shows search in progress
func (a *Application) showSearchProgress(message string) {
	progressText := tview.NewTextView()
	progressText.SetText(message + "\n\nPlease wait...")
	progressText.SetTextAlign(tview.AlignCenter)

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(progressText, 5, 0, true).
			AddItem(nil, 0, 1, false), 40, 0, true).
		AddItem(nil, 0, 1, false)

	// Set background color to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)

	a.pages.AddAndSwitchToPage("search_progress", flex, true)
}

// showSearchResults displays the search results
func (a *Application) showSearchResults() {
	if a.resultsList == nil {
		a.createResultsListEnhanced()
	}

	// Clear and populate results list
	a.resultsList.Clear()
	a.resultsList.SetTitle(fmt.Sprintf(" %s Search Results (%d found) ", IconMap["🔍"], len(a.searchResults)))

	if len(a.searchResults) == 0 {
		a.resultsList.AddItem("📦 No repositories found", "Try adjusting your search criteria", 0, nil)
	} else {
		for i, repo := range a.searchResults {
			// Create beautiful card-style layout
			selectedIcon := ""
			if a.selectedIndices[i] {
				selectedIcon = "[green]" + IconMap["✅"] + "[white] "
			}

			// Card header with repo name and language
			primary := fmt.Sprintf("%s[white::b]%s %s[-::-]", selectedIcon, IconMap["📁"], repo.FullName)
			if repo.Language != "" {
				primary += fmt.Sprintf(" [yellow]● %s[-]", repo.Language)
			}

			// Card body with description and stats
			var cardBody []string

			if repo.Description != "" {
				desc := repo.Description
				if len(desc) > 80 {
					desc = desc[:77] + "..."
				}
				cardBody = append(cardBody, fmt.Sprintf("[gray]%s[-]", desc))
			}

			// Stats with icons and colors
			statsLine := fmt.Sprintf("[yellow]%s %d[-]  [blue]%s %d[-]", IconMap["⭐"], repo.Stars, IconMap["🍴"], repo.Forks)
			if repo.UpdatedAt.Year() > 1 {
				statsLine += fmt.Sprintf("  [gray]%s %s[-]", IconMap["📅"], repo.UpdatedAt.Format("2006-01-02"))
			}
			cardBody = append(cardBody, statsLine)

			// Join card body with line breaks
			secondary := strings.Join(cardBody, "\n")

			index := i // Capture loop variable
			a.resultsList.AddItem(primary, secondary, 0, func() {
				// Simplified workflow - directly clone instead of showing details
				a.selectedRepos = []*ghClient.Repository{a.searchResults[index]}
				a.startCloning()
			})
		}
	}

	// Set up key bindings for results
	a.resultsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			a.showSearch()
			return nil
		case tcell.KeyCtrlA:
			a.selectAllResults()
			return nil
		case tcell.KeyCtrlD:
			a.deselectAllResults()
			return nil
		case tcell.KeyCtrlC:
			// Clone all selected repositories
			if len(a.selectedRepos) > 0 {
				a.startCloning()
			} else {
				a.cloneSelectedResults()
			}
			return nil
		}

		switch event.Rune() {
		case ' ':
			// Toggle selection of current item
			currentIndex := a.resultsList.GetCurrentItem()
			a.toggleSelection(currentIndex)
			return nil
		case 'a':
			a.selectAllResults()
			return nil
		case 'd':
			a.deselectAllResults()
			return nil
		case 'c':
			// Clone selected repositories immediately
			if len(a.selectedRepos) > 0 {
				a.startCloning()
			} else {
				a.cloneSelectedResults()
			}
			return nil
		case 's':
			a.showSearch()
			return nil
		case 'h':
			a.showSearch()
			return nil
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case 'l':
			return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		case 'i':
			// Show info/details for current repository
			currentIndex := a.resultsList.GetCurrentItem()
			if currentIndex >= 0 && currentIndex < len(a.searchResults) {
				a.showRepositoryDetails(currentIndex)
			}
			return nil
		}

		return event
	})

	// Create centered layout with selection count
	selectedCount := len(a.selectedRepos)
	infoText := tview.NewTextView()
	if selectedCount > 0 {
		infoText.SetText(fmt.Sprintf("[-]Keys:[-] Enter/l=Clone, Space=Toggle, c=Clone %d, i=Info, a=All, d=None, s/h=Back", selectedCount))
	} else {
		infoText.SetText("[-]Keys:[-] Enter/l=Clone, Space=Select, i=Info, c=Clone Selected, a=All, d=Clear, s/h=Back")
	}
	infoText.SetTextAlign(tview.AlignCenter)
	infoText.SetDynamicColors(true)
	infoText.SetBackgroundColor(tcell.ColorDefault)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.resultsList, 0, 1, true).
		AddItem(infoText, 2, 0, false)

	centeredFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(flex, 0, 4, true).
		AddItem(nil, 0, 1, false)

	// Set background colors to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)
	centeredFlex.SetBackgroundColor(tcell.ColorDefault)

	// Fix view stacking by removing old page first
	a.pages.RemovePage("results")
	a.pages.AddPage("results", centeredFlex, true, true)
	a.pages.SwitchToPage("results")
	a.currentState = StateResults
	a.app.SetFocus(a.resultsList)
}

// showRepositoryDetails shows detailed view of a repository
func (a *Application) showRepositoryDetails(index int) {
	if index >= len(a.searchResults) {
		return
	}

	repo := a.searchResults[index]

	detailsText := fmt.Sprintf(`[-]%s[-]

[-]Description:[-] %s

[-]Language:[-] %s
[-]Stars:[-] %d
[-]Forks:[-] %d
[-]Owner:[-] %s
[-]Private:[-] %t

[-]Clone URL:[-] %s
[-]SSH URL:[-] %s

[-]Updated:[-] %s

[-]Press 'c' to clone this repository, 'b'/'h' to go back[-]`,
		repo.FullName,
		repo.Description,
		repo.Language,
		repo.Stars,
		repo.Forks,
		repo.Owner,
		repo.Private,
		repo.CloneURL,
		repo.SSHURL,
		repo.UpdatedAt.Format("2006-01-02 15:04:05"))

	detailsView := tview.NewTextView()
	detailsView.SetText(detailsText)
	detailsView.SetBorder(true)
	detailsView.SetTitle(fmt.Sprintf(" %s Repository Details ", IconMap["📁"]))
	detailsView.SetDynamicColors(true)
	detailsView.SetWordWrap(true)

	detailsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'c', 'C':
			a.selectedRepos = []*ghClient.Repository{repo}
			a.startCloning()
			return nil
		case 'b', 'B', 'h':
			a.showSearchResults()
			return nil
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}

		switch event.Key() {
		case tcell.KeyEscape:
			a.showSearchResults()
			return nil
		}

		return event
	})

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(detailsView, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 4, true).
		AddItem(nil, 0, 1, false)

	// Set background color to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)

	a.pages.AddAndSwitchToPage("details", flex, true)
	a.app.SetFocus(detailsView)
}

// toggleSelection toggles the selection of a specific result
func (a *Application) toggleSelection(index int) {
	if index < 0 || index >= len(a.searchResults) {
		return
	}

	if a.selectedIndices[index] {
		delete(a.selectedIndices, index)
	} else {
		a.selectedIndices[index] = true
	}

	// Update selected repos array
	a.updateSelectedRepos()
	a.showSearchResults() // Refresh display
}

// selectAllResults selects all search results
func (a *Application) selectAllResults() {
	a.selectedIndices = make(map[int]bool)
	for i := range a.searchResults {
		a.selectedIndices[i] = true
	}
	a.updateSelectedRepos()
	a.message = fmt.Sprintf("Selected all %d repositories", len(a.selectedRepos))
	a.showSearchResults() // Refresh display
}

// deselectAllResults deselects all search results
func (a *Application) deselectAllResults() {
	a.selectedIndices = make(map[int]bool)
	a.selectedRepos = nil
	a.message = "Deselected all repositories"
	a.showSearchResults() // Refresh display
}

// updateSelectedRepos updates the selectedRepos array based on selectedIndices
func (a *Application) updateSelectedRepos() {
	a.selectedRepos = nil
	for index := range a.selectedIndices {
		if index < len(a.searchResults) {
			a.selectedRepos = append(a.selectedRepos, a.searchResults[index])
		}
	}
}

// cloneSelectedResults starts cloning selected repositories
func (a *Application) cloneSelectedResults() {
	if len(a.selectedRepos) == 0 {
		a.error = fmt.Errorf("no repositories selected")
		return
	}

	a.startCloning()
}

// createResultsListEnhanced creates the enhanced results list component
func (a *Application) createResultsListEnhanced() {
	a.resultsList = tview.NewList()
	a.resultsList.SetBorder(true)
	a.resultsList.ShowSecondaryText(true)
	a.resultsList.SetHighlightFullLine(true)

	// Fix highlighting colors for better visibility
	a.resultsList.SetSelectedTextColor(tcell.ColorBlack)           // Dark text on light background
	a.resultsList.SetSelectedBackgroundColor(tcell.ColorLightBlue) // Light blue selection
	a.resultsList.SetMainTextColor(tcell.ColorDefault)             // Default text color
	a.resultsList.SetSecondaryTextColor(tcell.ColorDefault)        // Default secondary text
	a.resultsList.SetBackgroundColor(tcell.ColorDefault)           // Transparent background

	// Card-style padding and rounded borders
	a.resultsList.SetBorderPadding(1, 1, 2, 2)
	a.resultsList.SetMainTextStyle(tcell.StyleDefault)
}
