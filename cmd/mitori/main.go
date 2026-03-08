package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wastingnotime/mitori/internal/config"
	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/parser"
	"github.com/wastingnotime/mitori/internal/service"
	"github.com/wastingnotime/mitori/internal/store/file"
	"github.com/wastingnotime/mitori/internal/tui"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		fatal(err)
	}

	st := file.New(cfg.DataPath)
	projectSvc := service.NewProjectService(st)
	taskSvc := service.NewTaskService(st, projectSvc)
	boardSvc := service.NewBoardService(st)
	doctorSvc := service.NewDoctorService(st)

	args := os.Args[1:]
	if len(args) == 0 {
		if err := st.Init(ctx); err != nil {
			fatal(err)
		}
		model, err := tui.NewModel(ctx, taskSvc, projectSvc, boardSvc)
		if err != nil {
			fatal(err)
		}
		if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
			fatal(err)
		}
		return
	}

	switch args[0] {
	case "init":
		if err := st.Init(ctx); err != nil {
			fatal(err)
		}
		fmt.Printf("initialized %s\n", st.Path())
	case "add":
		if err := st.Init(ctx); err != nil {
			fatal(err)
		}
		if len(args) < 2 {
			fatal(fmt.Errorf(`usage: mitori add "task title"`))
		}
		raw := strings.TrimSpace(strings.Join(args[1:], " "))
		parsed := parser.ParseTaskInput(raw)
		task, err := taskSvc.Create(ctx, service.CreateTaskInput{
			Title:       parsed.Title,
			Initiative:  parsed.Initiative,
			ProjectName: parsed.Project,
			Type:        domain.TaskTypeChore,
			Loop:        domain.LoopProduct,
			EnergyType:  domain.EnergyProduces,
			Nature:      domain.NatureMushin,
			Lane:        domain.LaneBacklog,
		})
		if err != nil {
			fatal(err)
		}
		fmt.Printf("added %s (%s)\n", task.Title, task.ID)
	case "projects":
		if err := st.Init(ctx); err != nil {
			fatal(err)
		}
		projects, err := projectSvc.List(ctx)
		if err != nil {
			fatal(err)
		}
		if len(projects) == 0 {
			fmt.Println("no projects")
			return
		}
		for _, p := range projects {
			fmt.Printf("%s  [%s]  %s\n", p.Name, p.Initiative, p.ID)
		}
	case "doctor":
		if err := st.Init(ctx); err != nil {
			fatal(err)
		}
		report, err := doctorSvc.Check(ctx)
		if err != nil {
			fatal(err)
		}
		if report.OK() {
			fmt.Println("doctor: ok")
			return
		}
		fmt.Println("doctor: issues found")
		for _, issue := range report.Issues {
			fmt.Printf("- %s\n", issue)
		}
		os.Exit(1)
	default:
		fatal(fmt.Errorf("unknown command: %s", args[0]))
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "mitori: %v\n", err)
	os.Exit(1)
}
