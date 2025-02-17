// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func selectPullRequestPrompt(pullRequests []apiclient.GitPullRequest, projectOrder int, selectionListOptions views.SelectionListOptions, choiceChan chan<- string, navChan chan<- string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, pr := range pullRequests {
		newItem := item[string]{
			id:             pr.Name,
			title:          pr.Name,
			choiceProperty: pr.Name,
			desc:           fmt.Sprintf("Branch: %s", views.GetBranchNameLabel(pr.Branch)),
		}
		items = append(items, newItem)
	}

	if !selectionListOptions.IsPaginationDisabled {
		items = AddLoadMoreOptionToList(items)
	}

	l := views.GetStyledSelectList(items, selectionListOptions)

	title := "Choose a Pull/Merge Request"
	if projectOrder > 1 {
		title += fmt.Sprintf(" (Project #%d)", projectOrder)
	}
	l.Title = views.GetStyledMainTitle(title)
	l.Styles.Title = titleStyle
	m := model[string]{list: l}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[string]); ok && m.choice != nil {
		choice := *m.choice
		if !selectionListOptions.IsPaginationDisabled && choice == views.ListNavigationText {
			navChan <- choice
		} else {
			choiceChan <- choice
		}
	} else {
		choiceChan <- ""
	}
}

func GetPullRequestFromPrompt(pullRequests []apiclient.GitPullRequest, projectOrder int, selectionListOptions views.SelectionListOptions) (*apiclient.GitPullRequest, string) {
	choiceChan := make(chan string)
	navChan := make(chan string)

	go selectPullRequestPrompt(pullRequests, projectOrder, selectionListOptions, choiceChan, navChan)

	select {
	case pullRequestName := <-choiceChan:
		for _, pr := range pullRequests {
			if pr.Name == pullRequestName {
				return &pr, ""
			}
		}
		return nil, ""
	case navigate := <-navChan:
		return nil, navigate
	}
}
