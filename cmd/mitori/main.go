package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wastingnotime/mitori/internal/config"
	"github.com/wastingnotime/mitori/internal/domain"
	iexport "github.com/wastingnotime/mitori/internal/export"
	"github.com/wastingnotime/mitori/internal/parser"
	"github.com/wastingnotime/mitori/internal/service"
	"github.com/wastingnotime/mitori/internal/store"
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
			Type:        parsed.TaskType,
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
	case "archive":
		if err := st.Init(ctx); err != nil {
			fatal(err)
		}
		query := strings.TrimSpace(strings.Join(args[1:], " "))
		tasks, projects, err := taskSvc.ListArchivedFiltered(ctx, service.TaskFilter{Query: query})
		if err != nil {
			fatal(err)
		}
		if len(tasks) == 0 {
			fmt.Println("no archived tasks")
			return
		}
		for _, t := range tasks {
			project := "-"
			if p, ok := projects[t.ProjectID]; ok {
				project = p.Name
			}
			archivedAt := "-"
			if t.ArchivedAt != nil {
				archivedAt = t.ArchivedAt.Format("2006-01-02 15:04")
			}
			fmt.Printf("%s  [%s]  %s  (%s)  archived:%s\n", t.Title, t.Type, project, t.ID, archivedAt)
		}
	case "import":
		fmt.Println("import: intended to load tasks/projects/events from external data. not implemented yet.")
	case "export":
		if err := st.Init(ctx); err != nil {
			fatal(err)
		}
		if err := runExport(ctx, st, args[1:]); err != nil {
			fatal(err)
		}
	default:
		fatal(fmt.Errorf("unknown command: %s", args[0]))
	}
}

func runExport(ctx context.Context, st store.Store, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mitori export <board|project|task|archive> [args] [--format markdown|json] [--out path]")
	}

	svc := iexport.NewService(st)
	switch args[0] {
	case "board":
		format, out, _, err := parseExportFlags("export board", args[1:])
		if err != nil {
			return err
		}
		data, err := svc.BuildBoard(ctx)
		if err != nil {
			return err
		}
		content, err := renderExport(format, data, iexport.RenderBoardMarkdown)
		if err != nil {
			return err
		}
		return writeExportOutput(out, content)
	case "project":
		format, out, rest, err := parseExportFlags("export project", args[1:])
		if err != nil {
			return err
		}
		if len(rest) < 1 {
			return fmt.Errorf("usage: mitori export project <project-name-or-id> [--format markdown|json] [--out path]")
		}
		data, err := svc.BuildProject(ctx, strings.Join(rest, " "))
		if err != nil {
			return err
		}
		content, err := renderExport(format, data, iexport.RenderProjectMarkdown)
		if err != nil {
			return err
		}
		return writeExportOutput(out, content)
	case "task":
		format, out, rest, err := parseExportFlags("export task", args[1:])
		if err != nil {
			return err
		}
		if len(rest) != 1 {
			return fmt.Errorf("usage: mitori export task <task-id> [--format markdown|json] [--out path]")
		}
		data, err := svc.BuildTask(ctx, rest[0])
		if err != nil {
			return err
		}
		content, err := renderExport(format, data, iexport.RenderTaskMarkdown)
		if err != nil {
			return err
		}
		return writeExportOutput(out, content)
	case "archive":
		format, out, _, err := parseExportFlags("export archive", args[1:])
		if err != nil {
			return err
		}
		data, err := svc.BuildArchive(ctx)
		if err != nil {
			return err
		}
		content, err := renderExport(format, data, iexport.RenderArchiveMarkdown)
		if err != nil {
			return err
		}
		return writeExportOutput(out, content)
	default:
		return fmt.Errorf("unknown export target: %s", args[0])
	}
}

func parseExportFlags(name string, args []string) (format string, out string, rest []string, err error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	formatPtr := fs.String("format", "markdown", "")
	outPtr := fs.String("out", "", "")
	if err := fs.Parse(args); err != nil {
		return "", "", nil, err
	}
	return strings.ToLower(strings.TrimSpace(*formatPtr)), strings.TrimSpace(*outPtr), fs.Args(), nil
}

func renderExport[T any](format string, data T, md func(T) string) (string, error) {
	switch format {
	case "markdown":
		return md(data), nil
	case "json":
		return iexport.RenderJSON(data)
	default:
		return "", fmt.Errorf("unsupported format: %s (use markdown or json)", format)
	}
}

func writeExportOutput(outPath, content string) error {
	if outPath == "" {
		fmt.Print(content)
		return nil
	}
	return os.WriteFile(outPath, []byte(content), 0o644)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "mitori: %v\n", err)
	os.Exit(1)
}
